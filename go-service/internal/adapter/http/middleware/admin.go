package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// RequireAdmin 校验当前用户为管理员。
//
// 判定规则与 Python dependencies.get_current_active_admin 完全一致：
// 用户名为 "admin"，或拥有 admin / superadmin / administrator 任一角色。
// 判定逻辑本身在 identity.User.IsAdmin()，这里只做拦截，避免规则两处漂移。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			response.Abort(c, 401, "未认证")
			return
		}
		if !user.IsAdmin() {
			response.AbortWithError(c, identity.ErrAdminRequired)
			return
		}
		c.Next()
	}
}
