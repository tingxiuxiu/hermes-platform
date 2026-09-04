//go:build !otel

// 本文件是 no-op 模式的占位实现，编译于默认构建（未启用 otel build tag）。
// 这样即使 otel 依赖未安装，整个服务也能编译运行（降级路径）。

package telemetry

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// instrumentHTTPWithOTel 的 no-op 版本：不注入 trace。
func instrumentHTTPWithOTel(_ string, _ string, _ *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// initOTel 的 no-op 版本：不初始化 TracerProvider。
func initOTel(_ string, _ string, _ *slog.Logger) {}

// resetShutdown 占位（otel 模式覆盖）。
var _ = context.Background
