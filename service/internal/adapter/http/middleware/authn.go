// Package middleware 承载 HTTP 层的横切关注点（鉴权、CORS、日志、恢复）。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/authn"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// ctxUserKey 是 gin.Context 中存放当前用户的键。
// 用自定义类型避免与其他中间件键冲突。
type ctxUserKey struct{}

// AuthNDeps 是 AuthN 中间件的依赖。
type AuthNDeps struct {
	Issuer  *authn.Issuer
	Revoker *authn.Revoker
	Users   appidentity.UserRepository
	Cache   appidentity.UserCache
}

// AuthN 校验 JWT 并把当前用户注入 context。
//
// 校验顺序（任一步失败即 401/400/404 并中断）：
//  1. 取出 Authorization: Bearer <token>
//  2. 解析 JWT（过期与无效返回不同错误，便于前端区分刷新与重登）
//  3. 比对令牌版本：落后于服务端即已被登出/改密/禁用吊销
//  4. 加载用户：先查缓存，未命中回源数据库并回填缓存
//  5. 状态检查：软删除视为不存在(404)，禁用返回 400（对齐 Python 的 400 Inactive user）
//
// 关于缓存用户的密码哈希：缓存值刻意不含 password_hash。
// 需要校验原密码的用例（修改密码）必须自己通过仓储**重新加载**，
// 不能依赖这里注入的用户对象——否则会拿到空哈希。
func AuthN(deps AuthNDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		raw := extractBearer(c)
		if raw == "" {
			abortAuth(c, authn.ErrMissingCredentials)
			return
		}

		claims, err := deps.Issuer.ParseAccessToken(raw)
		if err != nil {
			abortAuth(c, err)
			return
		}

		userID, err := claims.UserID()
		if err != nil {
			abortAuth(c, authn.ErrTokenInvalid)
			return
		}

		// 令牌版本落后说明已被吊销（登出/改密/禁用/批量操作）
		current, err := deps.Revoker.Version(ctx, userID)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		if claims.TokenVersion < current {
			abortAuth(c, authn.ErrTokenRevoked)
			return
		}

		user, err := deps.Cache.Get(ctx, userID)
		if err != nil {
			// 缓存故障一律降级为未命中，不阻断请求
			user = nil
		}
		if user == nil {
			user, err = deps.Users.GetByID(ctx, userID)
			if err != nil {
				response.AbortWithError(c, err)
				return
			}
			if user == nil {
				abortAuth(c, authn.ErrTokenInvalid)
				return
			}
			_ = deps.Cache.Set(ctx, user)
		}

		if user.IsDeleted() {
			response.AbortWithError(c, identity.ErrUserNotFound)
			return
		}
		if !user.IsActive() {
			response.AbortWithError(c, identity.ErrUserInactive)
			return
		}

		c.Set(stringKey(), user)
		c.Next()
	}
}

// CurrentUser 取出中间件注入的当前用户。未经过 AuthN 时返回 nil, false。
func CurrentUser(c *gin.Context) (*identity.User, bool) {
	v, ok := c.Get(stringKey())
	if !ok {
		return nil, false
	}
	u, ok := v.(*identity.User)
	if !ok || u == nil {
		return nil, false
	}
	return u, true
}

// MustUser 取出当前用户，不存在时直接中断请求。
// 只能在已挂载 AuthN 的路由上使用。
func MustUser(c *gin.Context) (*identity.User, bool) {
	u, ok := CurrentUser(c)
	if !ok {
		response.Abort(c, 401, "未认证")
		return nil, false
	}
	return u, true
}

// CurrentUserID 取出当前用户 ID。
func CurrentUserID(c *gin.Context) (int64, bool) {
	u, ok := CurrentUser(c)
	if !ok {
		return 0, false
	}
	return u.ID(), true
}

func extractBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

// abortAuth 写入 401 响应并附带 WWW-Authenticate 头。
// 头信息是 OAuth2 规范要求的，前端据此弹出重新登录。
func abortAuth(c *gin.Context, err error) {
	c.Header("WWW-Authenticate", "Bearer")
	response.AbortWithError(c, err)
}

func stringKey() string { return "middleware.authn.user" }
