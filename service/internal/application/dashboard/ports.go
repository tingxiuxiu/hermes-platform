// Package dashboard 承载 dashboard 上下文的应用层：快照投影编排与读查询（CQRS 读侧）。
//
// 读侧依赖 automation 的 ExecutionReader（读源表），写快照表。
// 严格依赖方向：application/dashboard → domain/dashboard + application/automation（只读）。
package dashboard

import (
	"context"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/dashboard"
)

// ---------------------------------------------------------------------------
// 快照仓储端口
// ---------------------------------------------------------------------------

// SnapshotRepository 是 dashboard 快照表的持久化端口。
type SnapshotRepository interface {
	// UpsertSummary 幂等写入全局摘要（snapshot_key 固定 "global"）。
	UpsertSummary(ctx context.Context, s dashboard.SummarySnapshot) error

	// UpsertTrend 幂等写入一条日趋势（stat_date 主键）。
	UpsertTrend(ctx context.Context, t dashboard.TrendSnapshot) error

	// ReplaceExecutionSnapshots 单事务重建某 execution 的快照：
	// 写 execution 快照 + 全量 case 快照 + 删除该 execution 已失效的 case 行。
	ReplaceExecutionSnapshots(ctx context.Context, exec dashboard.ExecutionSnapshot, cases []dashboard.CaseSnapshot) error

	// GetSummary 读取全局摘要。不存在返回 nil, nil。
	GetSummary(ctx context.Context) (*dashboard.SummarySnapshot, error)

	// GetTrends 返回最近 N 天的日趋势（升序）。
	GetTrends(ctx context.Context, since time.Time) ([]dashboard.TrendSnapshot, error)

	// ListRunningExecutions 返回状态为 running 的 execution 快照。
	ListRunningExecutions(ctx context.Context) ([]dashboard.ExecutionSnapshot, error)

	// ListCasesByExecution 返回某 execution 的全部 case 快照。
	ListCasesByExecution(ctx context.Context, executionID int64) ([]dashboard.CaseSnapshot, error)
}

// ---------------------------------------------------------------------------
// automation 源数据读取端口（读侧依赖）
// ---------------------------------------------------------------------------

// ExecutionReader 是 dashboard 读取 automation 源表的端口。
// 实现方复用 automation 的 ExecutionRepository / ItemRepository / PipelineRepository。
type ExecutionReader interface {
	// ListExecutionsInRange 返回某时间窗口内（含终止态与 running）的 execution 元数据。
	ListExecutionsInRange(ctx context.Context, since time.Time) ([]dashboard.TerminalExecution, error)

	// ListTerminalCasesInRange 返回某时间窗口内的终止态用例（用于摘要与趋势聚合）。
	ListTerminalCasesInRange(ctx context.Context, since time.Time) ([]dashboard.TerminalCase, error)

	// GetExecutionInput 返回单条 execution 的完整投影输入（execution + 全部用例）。
	GetExecutionInput(ctx context.Context, executionID int64) (*dashboard.ExecutionInput, error)

	// RunningExecutionsCount 返回当前 running 状态的 execution 数。
	RunningExecutionsCount(ctx context.Context) (int, error)

	// ListRunningExecutionIDs 返回当前 running 状态的 execution ID 列表。
	// 供 RefreshAll 投影运行中 execution 的快照。
	ListRunningExecutionIDs(ctx context.Context) ([]int64, error)

	// ActivePipelinesCount 返回启用中的流水线数。
	ActivePipelinesCount(ctx context.Context) (int, error)
}
