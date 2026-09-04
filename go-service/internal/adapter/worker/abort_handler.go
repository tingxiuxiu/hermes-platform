package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/queue"
)

// StaleAborter 把心跳超时的 running execution 标为 aborted。
type StaleAborter interface {
	AbortStale(ctx context.Context, olderThan time.Duration) (int, []automation.Event, error)
}

// AbortHandler 消费 automation:abort-stale。
type AbortHandler struct {
	aborter StaleAborter
	timeout time.Duration
	logger  *slog.Logger
}

// NewAbortHandler 构造扫描处理器。timeout 为心跳超时阈值。
func NewAbortHandler(aborter StaleAborter, timeout time.Duration, logger *slog.Logger) *AbortHandler {
	if timeout <= 0 {
		timeout = time.Minute
	}
	return &AbortHandler{aborter: aborter, timeout: timeout, logger: logger}
}

// Register 把处理器挂到 ServeMux。
func (h *AbortHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(queue.TypeAutomationAbortStale, h.handle)
}

func (h *AbortHandler) handle(ctx context.Context, _ *asynq.Task) error {
	n, _, err := h.aborter.AbortStale(ctx, h.timeout)
	if err != nil {
		h.logger.Error("abort stale executions failed", "error", err)
		return err
	}
	if n > 0 {
		h.logger.Info("aborted stale executions", "count", n)
	}
	return nil
}
