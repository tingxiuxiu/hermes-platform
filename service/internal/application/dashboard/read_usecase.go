package dashboard

import (
	"context"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/dashboard"
)

// OverviewResult 是 dashboard 概览的查询结果（纯读模型，无 HTTP 概念）。
type OverviewResult struct {
	Summary           dashboard.SummarySnapshot
	Trends            []dashboard.TrendSnapshot
	RunningExecutions []dashboard.ExecutionSnapshot
}

// OverviewUseCase 编排 dashboard 概览查询（T-5.5）。
//
// 新鲜度策略（对齐 Python ensure_fresh_data）：
//  1. 快照不存在 → 全量重算；
//  2. 快照超过 TTL → 全量重算；
//  3. 源表有更新的数据（execution 更新晚于快照）→ 重算对应 execution。
//
// 带筛选参数时跳过快照、实时聚合（快照是粗粒度缓存，不支撑实时筛选）。
type OverviewUseCase struct {
	snapshots SnapshotRepository
	source    ExecutionReader
	projector *ProjectorUseCase
	clock     Clock
	// snapshotTTL 快照新鲜度阈值。
	snapshotTTL time.Duration
}

// NewOverviewUseCase 构造概览用例。
func NewOverviewUseCase(
	snapshots SnapshotRepository,
	source ExecutionReader,
	projector *ProjectorUseCase,
	clock Clock,
	snapshotTTL time.Duration,
) *OverviewUseCase {
	if snapshotTTL <= 0 {
		snapshotTTL = 30 * time.Second
	}
	return &OverviewUseCase{
		snapshots:   snapshots,
		source:      source,
		projector:   projector,
		clock:       clock,
		snapshotTTL: snapshotTTL,
	}
}

// GetOverview 返回概览。
// ensureFresh 为 true 时先按新鲜度策略重建，再读取。
func (uc *OverviewUseCase) GetOverview(ctx context.Context, ensureFresh bool) (*OverviewResult, error) {
	if ensureFresh {
		if err := uc.ensureFresh(ctx); err != nil {
			return nil, err
		}
	}

	summary, err := uc.snapshots.GetSummary(ctx)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		// 快照为空（首次），重建一次
		if err := uc.projector.RebuildGlobal(ctx, 7); err != nil {
			return nil, err
		}
		summary, err = uc.snapshots.GetSummary(ctx)
		if err != nil {
			return nil, err
		}
		if summary == nil {
			// 仍为空：返回空概览而不是报错（可能无任何数据）
			summary = &dashboard.SummarySnapshot{SnapshotKey: dashboard.SnapshotKey}
		}
	}

	since := uc.clock.Now().AddDate(0, 0, -6)
	trends, err := uc.snapshots.GetTrends(ctx, since)
	if err != nil {
		return nil, err
	}

	running, err := uc.snapshots.ListRunningExecutions(ctx)
	if err != nil {
		return nil, err
	}

	return &OverviewResult{
		Summary:           *summary,
		Trends:            trends,
		RunningExecutions: running,
	}, nil
}

// ensureFresh 按新鲜度策略重建快照。
func (uc *OverviewUseCase) ensureFresh(ctx context.Context) error {
	summary, err := uc.snapshots.GetSummary(ctx)
	if err != nil {
		return err
	}

	// 不存在或过期 → 全量重建（全局摘要 + 运行中 execution 快照）
	if summary == nil || uc.clock.Now().Sub(summary.UpdatedAt) > uc.snapshotTTL {
		return uc.projector.RefreshAll(ctx, 7)
	}

	// 存在且未过期：无需重建（异步事件驱动会在 Phase 6 补充）
	return nil
}

// RunningCasesResult 是运行中 execution 的 case 列表。
type RunningCasesResult struct {
	ExecutionID int64
	Items       []dashboard.CaseSnapshot
}

// RunningCasesUseCase 编排运行中 case 列表查询（T-5.6）。
type RunningCasesUseCase struct {
	snapshots   SnapshotRepository
	source      ExecutionReader
	projector   *ProjectorUseCase
	clock       Clock
	snapshotTTL time.Duration
}

// NewRunningCasesUseCase 构造运行中 case 列表用例。
func NewRunningCasesUseCase(
	snapshots SnapshotRepository,
	source ExecutionReader,
	projector *ProjectorUseCase,
	clock Clock,
	snapshotTTL time.Duration,
) *RunningCasesUseCase {
	if snapshotTTL <= 0 {
		snapshotTTL = 30 * time.Second
	}
	return &RunningCasesUseCase{
		snapshots:   snapshots,
		source:      source,
		projector:   projector,
		clock:       clock,
		snapshotTTL: snapshotTTL,
	}
}

// GetCases 返回某 execution 的 case 列表。
// 若快照过期或源数据更新，先重建该 execution 快照。
func (uc *RunningCasesUseCase) GetCases(ctx context.Context, executionID int64) (*RunningCasesResult, error) {
	if err := uc.ensureExecutionFresh(ctx, executionID); err != nil {
		return nil, err
	}

	cases, err := uc.snapshots.ListCasesByExecution(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return &RunningCasesResult{ExecutionID: executionID, Items: cases}, nil
}

func (uc *RunningCasesUseCase) ensureExecutionFresh(ctx context.Context, executionID int64) error {
	// 简化：直接重算该 execution 快照（源数据变化时投影是最新的）。
	// 事件驱动的增量刷新在 Phase 6 接入后，这里只做读。
	return uc.projector.ProjectExecution(ctx, executionID)
}
