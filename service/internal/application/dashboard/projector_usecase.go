package dashboard

import (
	"context"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/dashboard"
)

// ProjectorUseCase 编排快照投影（T-5.3 / T-5.4）。
//
// 职责：读 automation 源表 → 用领域投影器算出快照 → 写回 dashboard 快照表。
// 事件驱动重建（asynq）在 Phase 6 接入；这里的 ProjectExecution 可被任务处理器复用。
type ProjectorUseCase struct {
	snapshots SnapshotRepository
	source    ExecutionReader
	clock     Clock
}

// NewProjectorUseCase 构造投影用例。
func NewProjectorUseCase(snapshots SnapshotRepository, source ExecutionReader, clock Clock) *ProjectorUseCase {
	return &ProjectorUseCase{snapshots: snapshots, source: source, clock: clock}
}

// ProjectExecution 重算单条 execution 的快照（execution + case 快照）。
// 供运行中 execution 卡片与 running-case 列表使用。
func (uc *ProjectorUseCase) ProjectExecution(ctx context.Context, executionID int64) error {
	in, err := uc.source.GetExecutionInput(ctx, executionID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	execSnapshot := dashboard.BuildExecutionSnapshot(*in, now)
	caseSnapshots := dashboard.BuildCaseSnapshots(*in, now)

	return uc.snapshots.ReplaceExecutionSnapshots(ctx, execSnapshot, caseSnapshots)
}

// RebuildGlobal 重算全局摘要 + 日趋势（近 days 天）。
func (uc *ProjectorUseCase) RebuildGlobal(ctx context.Context, days int) error {
	if days <= 0 {
		days = 7
	}
	now := uc.clock.Now()
	since := dashboard.DayKey(now).AddDate(0, 0, -(days - 1))

	terminalExecs, err := uc.source.ListExecutionsInRange(ctx, since)
	if err != nil {
		return err
	}
	terminalCases, err := uc.source.ListTerminalCasesInRange(ctx, since)
	if err != nil {
		return err
	}
	runningCount, err := uc.source.RunningExecutionsCount(ctx)
	if err != nil {
		return err
	}
	activeJobs, err := uc.source.ActivePipelinesCount(ctx)
	if err != nil {
		return err
	}

	// 全局摘要
	summary := dashboard.BuildSummarySnapshot(dashboard.SummaryInput{
		TerminalExecutions:     terminalExecs,
		TerminalCases:          terminalCases,
		ActiveJobsCount:        activeJobs,
		RunningExecutionsCount: runningCount,
		Now:                    now,
	})
	if err := uc.snapshots.UpsertSummary(ctx, summary); err != nil {
		return err
	}

	// 日趋势（补齐空日期）
	trends := dashboard.BuildTrendSnapshots(terminalCases, terminalExecs, now, days)
	for _, t := range trends {
		if err := uc.snapshots.UpsertTrend(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

// RefreshAll 全量刷新：全局摘要 + 每个运行中 execution 的快照。
// 供 fresh 判定后兜底调用。
func (uc *ProjectorUseCase) RefreshAll(ctx context.Context, days int) error {
	if err := uc.RebuildGlobal(ctx, days); err != nil {
		return err
	}

	// 从**源表**取 running execution（快照表可能是空的，不能作为投影来源）
	runningIDs, err := uc.source.ListRunningExecutionIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range runningIDs {
		if err := uc.ProjectExecution(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// Clock 是时间源端口。
type Clock interface {
	Now() time.Time
}
