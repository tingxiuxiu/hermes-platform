//go:build otel

// 本文件是完整 OpenTelemetry 实现，编译于启用 otel build tag 时：
//
//	go build -tags otel ./cmd/api
//
// 默认构建（无 tag）走 noop_otel.go 的降级路径，因此本文件引用的
// otel 依赖只在显式开启时参与编译。

package telemetry

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// instrumentHTTPWithOTel 返回 otelgin 中间件（完整模式）。
func instrumentHTTPWithOTel(endpoint string, serviceName string, _ *slog.Logger) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// initOTel 初始化全局 TracerProvider 与 OTLP exporter（完整模式）。
func initOTel(endpoint string, serviceName string, logger *slog.Logger) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		logger.Error("telemetry: create OTLP exporter failed, tracing disabled",
			"error", err)
		return
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		logger.Error("telemetry: build resource failed", "error", err)
		return
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// 覆盖包级 Shutdown，让进程退出时能刷新并关闭 exporter。
	Shutdown = func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}

	logger.Info("telemetry: OTLP tracing enabled", "endpoint", endpoint)
}
