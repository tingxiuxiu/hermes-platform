// Package dashboard 承载 dashboard 上下文的领域模型（CQRS 读模型）。
//
// 这是**读侧**：数据由投影器从 automation 源表重算后写入快照表，
// 本包只定义快照实体与投影输入结构，不做任何业务决策。
package dashboard

import (
	"time"
)

// SnapshotKey 是全局摘要快照的主键值（固定 "global"）。
const SnapshotKey = "global"

// SummarySnapshot 对应 automation_dashboard_summary_snapshots。
type SummarySnapshot struct {
	SnapshotKey            string
	ActiveJobsCount        int
	TotalExecutionsCount   int
	RunningExecutionsCount int
	SuccessCases7d         int
	FailureCases7d         int
	SkippedCases7d         int
	PassRate7d             float64
	AvgExecutionDuration7d *float64
	UpdatedAt              time.Time
}

// TrendSnapshot 对应 automation_dashboard_daily_trends（按 stat_date 主键）。
type TrendSnapshot struct {
	StatDate       time.Time // 仅日期部分有效
	ExecutionTotal int
	SuccessCases   int
	FailureCases   int
	SkippedCases   int
	RunningCases   int
}

// ExecutionSnapshot 对应 automation_dashboard_execution_snapshots。
//
// 命名漂移修正（D-14）：数据库列是 start_time，DTO 是 started_at；
// 实体内部统一用 StartedAt，映射到 DTO 时再改字段名。
type ExecutionSnapshot struct {
	ExecutionID         int64
	BuildUID            string
	JobID               int64
	JobName             string
	JobURL              string
	Status              string
	StartedAt           time.Time
	Duration            *float64
	PreCasesCount       int
	CompletedCasesCount int
	SuccessCasesCount   int
	FailureCasesCount   int
	SkippedCasesCount   int
	RunningCasesCount   int
	ProgressPercent     float64
	RunningCaseNames    []string
	UpdatedAt           time.Time
}

// CaseSnapshot 对应 automation_dashboard_execution_item_snapshots。
//
// 命名漂移修正（D-15）：数据库列是 case_id（即 execution_items.id），
// DTO 字段是 item_id。实体内部统一用 ItemID，映射到 DTO 时改字段名。
type CaseSnapshot struct {
	ID            int64
	ExecutionID   int64
	ItemID        int64
	CaseKey       string
	CaseName      string
	AttemptNumber int
	Status        string
	StartedAt     *time.Time
	EndedAt       *time.Time
	Duration      *float64
	ErrorMessage  string
	UpdatedAt     time.Time
}

// ExecutionInput 是 execution 快照投影器所需的源数据（从 automation 侧读取）。
type ExecutionInput struct {
	ExecutionID   int64
	BuildUID      string
	JobID         int64
	JobName       string
	JobURL        string
	Status        string
	StartedAt     time.Time
	Duration      *float64
	PreCasesCount int
	// Cases 是该 execution 下全部用例（is_latest 任意）。
	Cases []CaseInput
}

// CaseInput 是 case 快照投影器所需的单条用例数据。
type CaseInput struct {
	ItemID        int64
	CaseKey       string
	CaseName      string
	AttemptNumber int
	Status        string
	StartedAt     *time.Time
	EndedAt       *time.Time
	Duration      *float64
	ErrorMessage  string
}

// TerminalExecution 是汇总投影所需的终止态 execution 数据。
type TerminalExecution struct {
	ExecutionID int64
	JobID       int64
	JobName     string
	JobURL      string
	Status      string
	StartedAt   time.Time
	Duration    *float64
}

// TerminalCase 是汇总投影所需的终止态用例数据。
type TerminalCase struct {
	ExecutionID int64
	CaseKey     string
	Status      string
	StartedAt   *time.Time
}
