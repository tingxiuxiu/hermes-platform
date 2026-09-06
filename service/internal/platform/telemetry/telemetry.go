// Package telemetry 封装 OpenTelemetry 接入（T-8.2）。
//
// 设计：`OTEL_EXPORTER_OTLP_ENDPOINT` 为空时**降级为 no-op**，不阻塞启动。
// 本机（无 OTLP collector）走 no-op，生产（配置了端点）才真正导出 trace。
//
// 这样本地开发与集成测试不需要真实 trace 后端，也不会因为无法连接 collector
// 而启动失败——满足「不阻塞启动」的验收标准。
package telemetry

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Shutdown 在进程退出时刷新并关闭 trace provider。
// no-op 模式下是个空操作；完整模式下会刷新并关闭 exporter。
var Shutdown = func(context.Context) error { return nil }

// InstrumentHTTP 返回一个 HTTP 中间件。
// 当配置了 OTLP 端点时返回 otelgin 中间件；否则返回透传 no-op。
func InstrumentHTTP(endpoint string, serviceName string, logger *slog.Logger) gin.HandlerFunc {
	if endpoint == "" {
		// 降级：无 OTLP collector，不注入 trace。
		return func(c *gin.Context) { c.Next() }
	}
	return instrumentHTTPWithOTel(endpoint, serviceName, logger)
}

// Init 在启动期初始化全局 TracerProvider（完整模式下）。
// no-op 模式无操作。
func Init(endpoint, serviceName string, logger *slog.Logger) {
	if endpoint == "" {
		logger.Info("telemetry: OTEL_EXPORTER_OTLP_ENDPOINT empty, tracing disabled")
		return
	}
	initOTel(endpoint, serviceName, logger)
}
