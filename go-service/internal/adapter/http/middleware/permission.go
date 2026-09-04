package middleware

import (
	"github.com/gin-gonic/gin"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// RequirePermission 校验当前用户拥有指定权限码。
//
// 本轮**默认不在任何路由上启用**（沿用 Python 的管理员粗粒度拦截），
// 但能力保留，为后续细粒度鉴权做准备（03-api-contract.md §8）。
//
// 注意：权限码必须**每次从数据库查询**，不能依赖 AuthN 注入的用户对象——
// 该对象可能来自缓存，其中的角色不含权限列表。
func RequirePermission(permissions appidentity.PermissionRepository, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		user, ok := CurrentUser(c)
		if !ok {
			response.Abort(c, 401, "未认证")
			return
		}

		// 管理员直接放行，与 Python require_permission 的短路规则一致
		if user.IsAdmin() {
			c.Next()
			return
		}

		roleIDs := make([]int64, 0, len(user.Roles()))
		for _, r := range user.Roles() {
			roleIDs = append(roleIDs, r.ID())
		}
		if len(roleIDs) == 0 {
			response.AbortWithError(c, identity.ErrMissingPermission(code))
			return
		}

		codes, err := permissions.ListByRoleIDs(ctx, roleIDs)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		for _, owned := range codes {
			if owned == code {
				c.Next()
				return
			}
		}

		response.AbortWithError(c, identity.ErrMissingPermission(code))
	}
}
