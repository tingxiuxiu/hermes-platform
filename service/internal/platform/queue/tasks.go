package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// 任务类型常量（契约：worker 与 client 共享，不得各自硬编码）。
const (
	// TypeDashboardRefreshExecution 刷新单个 execution 的快照。
	TypeDashboardRefreshExecution = "dashboard:refresh:execution"
	// TypeDashboardRefreshAll 全量刷新 dashboard（全局摘要 + 运行中 execution）。
	TypeDashboardRefreshAll = "dashboard:refresh:all"
	// TypeAutomationAbortStale 把心跳超时的 running execution 标为 aborted。
	TypeAutomationAbortStale = "automation:abort-stale"
)

// DashboardRefreshExecutionPayload 是 dashboard:refresh:execution 的载荷。
type DashboardRefreshExecutionPayload struct {
	ExecutionID int64 `json:"execution_id"`
}

// DashboardRefreshAllPayload 是 dashboard:refresh:all 的载荷。
// 全量刷新没有必填参数，保留结构便于未来扩展（如指定天数）。
type DashboardRefreshAllPayload struct {
	// TrendDays 趋势天数；<=0 时用配置默认值。
	TrendDays int `json:"trend_days,omitempty"`
}

// NewDashboardRefreshExecutionTask 构造刷新单 execution 的任务。
// 启用 5 分钟唯一性：同一 execution 在 5 分钟内多次事件只触发一次投影，避免抖动风暴。
func NewDashboardRefreshExecutionTask(executionID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(DashboardRefreshExecutionPayload{ExecutionID: executionID})
	if err != nil {
		return nil, fmt.Errorf("queue: marshal execution refresh payload: %w", err)
	}
	return asynq.NewTask(TypeDashboardRefreshExecution, payload), nil
}

// ParseDashboardRefreshExecutionPayload 解析执行刷新任务载荷。
func ParseDashboardRefreshExecutionPayload(task *asynq.Task) (DashboardRefreshExecutionPayload, error) {
	var p DashboardRefreshExecutionPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return p, fmt.Errorf("queue: parse execution refresh payload: %w", err)
	}
	if p.ExecutionID <= 0 {
		return p, fmt.Errorf("queue: execution_id must be positive, got %d", p.ExecutionID)
	}
	return p, nil
}

// NewDashboardRefreshAllTask 构造全量刷新任务。
func NewDashboardRefreshAllTask(trendDays int) (*asynq.Task, error) {
	payload, err := json.Marshal(DashboardRefreshAllPayload{TrendDays: trendDays})
	if err != nil {
		return nil, fmt.Errorf("queue: marshal all-refresh payload: %w", err)
	}
	return asynq.NewTask(TypeDashboardRefreshAll, payload), nil
}

// ParseDashboardRefreshAllPayload 解析全量刷新任务载荷。
func ParseDashboardRefreshAllPayload(task *asynq.Task) (DashboardRefreshAllPayload, error) {
	var p DashboardRefreshAllPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return p, fmt.Errorf("queue: parse all-refresh payload: %w", err)
	}
	return p, nil
}

// NewAbortStaleTask 构造扫描僵尸 running execution 的任务。
func NewAbortStaleTask() (*asynq.Task, error) {
	return asynq.NewTask(TypeAutomationAbortStale, []byte(`{}`)), nil
}

// CommonOptions 是 dashboard 任务的通用选项：
// 入 dashboard 队列、重试 3 次、超时 30s、保留 1h。
func CommonOptions() []asynq.Option {
	return []asynq.Option{
		WithQueue(QueueDashboard),
		WithMaxRetry(3),
		WithTimeout(30 * time.Second),
		WithRetention(time.Hour),
	}
}

// AbortStaleOptions 是僵尸 execution 扫描任务的选项：入 default 队列、短 Unique 防重叠。
func AbortStaleOptions() []asynq.Option {
	return []asynq.Option{
		WithQueue(QueueDefault),
		WithMaxRetry(3),
		WithTimeout(30 * time.Second),
		WithRetention(time.Hour),
		WithUnique(10 * time.Second),
	}
}
