package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/queue"
)

// noopEventPublisher 是一个占位的事件发布器。
//
// 集成测试不跑 asynq，用它替代真实桥接器——上报链路照常工作，
// 只是不投递 dashboard 刷新任务（读模型靠同步刷新兜底）。
type noopEventPublisher struct{}

func (noopEventPublisher) Publish(_ context.Context, _ automation.Event) error {
	return nil
}

// EventPublisher 是领域事件 → asynq 任务的桥接器（T-6.4）。
//
// 职责：把 automation 层发布的领域事件映射为 asynq 任务入队。
// automation 层**不感知 asynq**（ADR-0004）——它只调 EventPublisher.Publish，
// 具体的任务映射全在这里。
type EventPublisher struct {
	client *queue.Client
}

// NewEventPublisher 构造 asynq 事件桥接器。
func NewEventPublisher(client *queue.Client) *EventPublisher {
	return &EventPublisher{client: client}
}

// Publish 把领域事件映射为 asynq 任务。
//
// 事件 → 任务映射表：
//   - ExecutionCreated / ExecutionUpdated → dashboard:refresh:all（新增/结束 execution 影响全局摘要）
//   - ItemUpdated → dashboard:refresh:execution{execution_id}（用例状态影响该 execution 快照）
//
// 去重：dashboard 任务用 Unique(5m) 去重，防止高频上报造成投影风暴。
//
// 发布失败不阻断业务：写侧已完成，投影可由 scheduler 兜底重算（refresh:all 定时投递）。
func (p *EventPublisher) Publish(ctx context.Context, event automation.Event) error {
	switch e := event.(type) {
	case automation.ExecutionCreated, automation.ExecutionUpdated:
		return p.enqueueRefreshAll()

	case automation.ItemUpdated:
		// 注意：execution_id 可能为 0（ingest 用例未能反查），此时跳过——
		// ItemUpdated 只驱动单 execution 刷新，缺 execution_id 则依赖 refresh:all 兜底。
		return p.enqueueRefreshExecution(e.ExecutionID)

	default:
		return nil
	}
}

func (p *EventPublisher) enqueueRefreshAll() error {
	task, err := queue.NewDashboardRefreshAllTask(0)
	if err != nil {
		return fmt.Errorf("bootstrap: build refresh-all task: %w", err)
	}
	_, err = p.client.Enqueue(task, p.commonOptions()...)
	return err
}

func (p *EventPublisher) enqueueRefreshExecution(executionID int64) error {
	if executionID <= 0 {
		return nil
	}
	task, err := queue.NewDashboardRefreshExecutionTask(executionID)
	if err != nil {
		return fmt.Errorf("bootstrap: build refresh-execution task: %w", err)
	}
	_, err = p.client.Enqueue(task, p.commonOptions()...)
	return err
}

// commonOptions 是 dashboard 任务的通用选项：入 dashboard 队列、Unique(5m) 去重。
func (p *EventPublisher) commonOptions() []asynq.Option {
	return append(
		queue.CommonOptions(),
		queue.WithUnique(5*time.Minute),
	)
}
