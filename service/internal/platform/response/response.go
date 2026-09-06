// Package response 统一 HTTP 响应信封。
//
// 契约严格复刻 Python 版 core/base_response.py 的 BaseResponse：
//
//	{ "code": 200, "message": "登录成功", "data": {...}, "success": true }
//
// 字段顺序、success 默认值（true）、data 为 nil 时输出 null（而非省略字段）
// 都必须与 Python 侧一致，否则前端解析会出错（ADR-0001）。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// Envelope 是所有接口的统一响应体。字段顺序即 JSON 输出顺序，不得调整。
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

// OK 返回业务成功响应，HTTP 状态码固定 200。
// 业务 code 与 HTTP 状态码保持一致（Python 侧也是 200）。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		Success: true,
	})
}

// OKWithMessage 返回带自定义消息的业务成功响应。
// Python 侧各接口都有中文消息文案（如「获取用户列表成功」），必须逐个对齐。
func OKWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
		Success: true,
	})
}

// Fail 返回业务失败响应。status 为 HTTP 状态码，code 默认与 status 相同。
func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{
		Code:    status,
		Message: message,
		Data:    nil,
		Success: false,
	})
}

// FailWithCode 用于业务 code 需要与 HTTP 状态码解耦的场景。
func FailWithCode(c *gin.Context, status, code int, message string) {
	c.JSON(status, Envelope{
		Code:    code,
		Message: message,
		Data:    nil,
		Success: false,
	})
}

// Error 把领域错误翻译为响应。
// Kind 决定 HTTP 状态码与业务 code，消息取自 errors.Message（内部错误已脱敏）。
func Error(c *gin.Context, err error) {
	kind := errors.KindOf(err)
	status := kind.HTTPStatus()
	c.JSON(status, Envelope{
		Code:    kind.Code(),
		Message: errors.Message(err),
		Data:    nil,
		Success: false,
	})
}

// Abort 用于中间件：写入错误响应后立即中断后续 handler。
func Abort(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{
		Code:    status,
		Message: message,
		Data:    nil,
		Success: false,
	})
	c.Abort()
}

// AbortWithError 用于中间件：写入领域错误响应后立即中断。
func AbortWithError(c *gin.Context, err error) {
	Error(c, err)
	c.Abort()
}
