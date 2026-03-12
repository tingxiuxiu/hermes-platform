package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionChecker 权限检查接口
type PermissionChecker interface {
	CheckPermission(userID uint, permissionCode string) (bool, error)
}

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permissionChecker PermissionChecker, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户 ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "unauthorized",
				"error":   "User not authenticated",
			})
			c.Abort()
			return
		}

		// 检查用户是否具有所需的权限
		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "internal server error",
				"error":   "Invalid user ID format",
			})
			c.Abort()
			return
		}

		hasPermission, err := permissionChecker.CheckPermission(uid, permissionCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "internal server error",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "forbidden",
				"error":   "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
