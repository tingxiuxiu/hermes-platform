// Package worker 承载 asynq 任务处理器（worker 消费端）。
package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/hermes-platform/go-service/internal/platform/queue"
)

// DashboardHandler 消费 dashboard 快照投影任务。
//
// 这些任务由 bootstrap 的 EventPublisher 桥接器投递（写侧事件驱动），
// 也由 scheduler 定时投递（refresh:all 兜底）。
type DashboardHandler struct {
	// projector 复用 dashboard 应用层的投影用例。
	projector DashboardProjector
	logger    *slog.Logger
}

// DashboardProjector 是投影用例的接口（由 bootstrap 提供具体实现）。
// 用接口隔离，避免 worker 直接依赖 dashboard 应用层具体类型，便于测试注入 fake。
type DashboardProjector interface {
	ProjectExecution(ctx context.Context, executionID int64) error
	RefreshAll(ctx context.Context, days int) error
}

// NewDashboardHandler 构造 dashboard 任务处理器。
func NewDashboardHandler(projector DashboardProjector, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{projector: projector, logger: logger}
}

// Register 把处理器注册到 ServeMux。
func (h *DashboardHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(queue.TypeDashboardRefreshExecution, h.handleRefreshExecution)
	mux.HandleFunc(queue.TypeDashboardRefreshAll, h.handleRefreshAll)
}

// Handle 按任务类型分发。公开方法便于单测直接调用（绕过 ServeMux）。
func (h *DashboardHandler) Handle(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case queue.TypeDashboardRefreshExecution:
		return h.handleRefreshExecution(ctx, task)
	case queue.TypeDashboardRefreshAll:
		return h.handleRefreshAll(ctx, task)
	default:
		h.logger.Warn("unknown task type, ignoring", "type", task.Type())
		return nil
	}
}

func (h *DashboardHandler) handleRefreshExecution(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.ParseDashboardRefreshExecutionPayload(task)
	if err != nil {
		// 载荷损坏：不重试（重试也无意义）
		h.logger.Error("invalid refresh-execution task payload", "error", err)
		return nil
	}

	h.logger.Info("dashboard refresh execution",
		"execution_id", payload.ExecutionID, "type", task.Type())
	if err := h.projector.ProjectExecution(ctx, payload.ExecutionID); err != nil {
		// 返回 error 触发 asynq 重试（最多 3 次）
		h.logger.Error("dashboard refresh execution failed",
			"execution_id", payload.ExecutionID, "error", err)
		return err
	}
	return nil
}

func (h *DashboardHandler) handleRefreshAll(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.ParseDashboardRefreshAllPayload(task)
	if err != nil {
		h.logger.Error("invalid refresh-all task payload", "error", err)
		return nil
	}

	days := payload.TrendDays
	if days <= 0 {
		days = 7 // 默认 7 天
	}

	h.logger.Info("dashboard refresh all", "trend_days", days, "type", task.Type())
	if err := h.projector.RefreshAll(ctx, days); err != nil {
		h.logger.Error("dashboard refresh all failed", "error", err)
		return err
	}
	return nil
}
