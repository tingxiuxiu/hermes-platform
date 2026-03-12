package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求方法
		method := c.Request.Method

		// 请求路由
		path := c.Request.URL.Path

		// 状态码
		statusCode := c.Writer.Status()

		// 客户端 IP
		clientIP := c.ClientIP()

		// 日志格式
		logMsg := fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %s",
			method,
			statusCode,
			latency,
			clientIP,
			path,
		)

		// 根据状态码设置日志级别
		switch {
		case statusCode >= 500:
			fmt.Println("[ERROR]" + logMsg)
		case statusCode >= 400:
			fmt.Println("[WARN]" + logMsg)
		default:
			fmt.Println("[INFO]" + logMsg)
		}
	}
}