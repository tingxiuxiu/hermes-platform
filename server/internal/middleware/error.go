package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last()

			// 根据错误类型返回不同的状态码和消息
			switch err.Type {
			case gin.ErrorTypeBind:
				// 参数绑定错误
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "bad request",
					"error":   err.Error(),
				})
			case gin.ErrorTypePublic:
				// 公共错误
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "bad request",
					"error":   err.Error(),
				})
			default:
				// 服务器内部错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "internal server error",
					"error":   "Internal server error",
				})
			}

			// 终止后续处理
			c.Abort()
		}
	}
}