package middleware

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/platform/authn"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// ServiceToken 校验上报接口的服务令牌（X-Service-Token）。
//
// 这是 pytest/CI Runner 上报数据的鉴权方式，与用户 JWT 完全独立。
// 比对用常量时间，避免通过响应时间侧信道探测令牌。
//
// 要求：令牌由配置 AUTOMATION_SERVICE_TOKEN 指定，环境里必须与 Python 侧一致，
// 否则插件上报会被拒。禁止为空（空令牌等于关闭鉴权）。
func ServiceToken(expected string) gin.HandlerFunc {
	if expected == "" {
		// 配置缺失是启动期错误，这里 fail-fast：任何上报都被拒，
		// 而不是静默放行造成安全洞。
		return func(c *gin.Context) {
			response.AbortWithError(c, authn.ErrServiceTokenInvalid)
		}
	}

	return func(c *gin.Context) {
		got := c.GetHeader("X-Service-Token")
		if got == "" || !constantTimeEqual(got, expected) {
			response.AbortWithError(c, authn.ErrServiceTokenInvalid)
			return
		}
		c.Next()
	}
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
