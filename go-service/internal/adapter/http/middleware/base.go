package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/logging"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// RequestID 为每个请求生成标识并写入响应头。
// 排查线上问题时，前端报过来的 X-Request-ID 能直接对应到服务端日志。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

// Recover 把 panic 转成 500 响应，避免单个请求打挂整个进程。
// 堆栈打到日志里，但**不返回给调用方**。
func Recover(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("http handler panicked",
					"panic", rec,
					"path", c.FullPath(),
					"method", c.Request.Method,
					"request_id", c.GetString("request_id"),
				)
				response.Abort(c, 500, "internal server error")
			}
		}()
		c.Next()
	}
}

// Logger 输出访问日志。
// 刻意不记录请求体与响应体：登录接口会经过这里，记录 body 等于把密码写进日志。
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// allowedMethods 与 allowedHeaders 是 CORS 预检响应的固定值。
var (
	allowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowedHeaders = "Origin, Content-Type, Authorization, X-Service-Token, X-Request-ID"
)

// CORS 配置跨域。允许的来源来自配置（FRONTEND_HOST + BACKEND_CORS_ORIGINS）。
//
// 刻意不引入 gin-contrib/cors：我们的需求只是「按配置白名单放几个响应头」，
// 自己实现约 40 行，省掉一个第三方依赖与随之而来的版本/校验问题。
func CORS(cfg config.HTTPConfig) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.CORSOrigins)+2)
	for _, o := range cfg.CORSOrigins {
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	if cfg.FrontendHost != "" {
		allowed[cfg.FrontendHost] = struct{}{}
	}
	allowAll := len(allowed) == 0

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if _, ok := allowed[origin]; ok {
			// 命中白名单时回显具体来源，而不是 "*"，
			// 这样带 credentials 的请求才能通过浏览器校验
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			// 不在白名单：不写任何 CORS 头，浏览器会拦截
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// WithLogger 把 logger 注入 gin context，供 handler 通过 logging.FromContext 取用。
func WithLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(logging.WithLogger(c.Request.Context(), logger))
		c.Next()
	}
}

// newRequestID 生成请求标识：时间戳（纳秒，36 进制）+ 随机后缀。
// 前段有序便于日志按时间粗排序，后段随机避免多实例之间冲突。
func newRequestID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand 失败属于严重环境问题，退化为纯时间戳以保证请求不被阻塞
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + hex.EncodeToString(buf[:])
}
