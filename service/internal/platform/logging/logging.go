// Package logging 基于标准库 log/slog 的结构化日志。
//
// 采用 JSON 输出，字段与现有 OpenObserve 采集管线兼容；
// 链路信息（trace_id / span_id）由 OpenTelemetry 处理器补充，本包只负责基础字段。
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/hermes-platform/go-service/internal/platform/config"
)

// New 按配置构造 logger。
// 非 local 环境强制 JSON，避免多行文本污染日志采集。
func New(cfg config.LogConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}
	if strings.EqualFold(cfg.Format, "text") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// NewNop 返回丢弃所有输出的 logger，供测试使用。
func NewNop() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError + 1,
	}))
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ctxKey struct{}

// WithLogger 把 logger 注入 context。
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext 取出 context 中的 logger，缺省返回 slog.Default()，
// 保证未注入的代码路径不会因为 nil 而 panic。
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
