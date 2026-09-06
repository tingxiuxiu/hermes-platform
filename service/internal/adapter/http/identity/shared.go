package identity

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/platform/response"
)

// pathID 解析路径参数中的正整数 ID。
// 非法时写入 422 响应并返回 false——避免每个 handler 重复写这段解析与错误处理。
func pathID(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, 422, "路径参数不合法")
		return 0, false
	}
	return id, true
}
