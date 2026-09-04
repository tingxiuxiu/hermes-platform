package automation

import (
	"context"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/domain/automation"
)

// QueryUseCase 编排查询类用例（T-3.11）。
type QueryUseCase struct {
	executions  ExecutionRepository
	items       ItemRepository
	steps       StepRepository
	attachments AttachmentRepository
	pipelines   PipelineRepository
}

// NewQueryUseCase 构造查询用例。
func NewQueryUseCase(
	executions ExecutionRepository,
	items ItemRepository,
	steps StepRepository,
	attachments AttachmentRepository,
	pipelines PipelineRepository,
) *QueryUseCase {
	return &QueryUseCase{
		executions:  executions,
		items:       items,
		steps:       steps,
		attachments: attachments,
		pipelines:   pipelines,
	}
}

// ListPipelines 查询流水线列表（D-07）。
func (uc *QueryUseCase) ListPipelines(ctx context.Context, f PipelineListFilter) ([]*automation.JenkinsPipeline, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	return uc.pipelines.List(ctx, f)
}

// ListExecutions 查询执行列表。
func (uc *QueryUseCase) ListExecutions(ctx context.Context, f ExecutionListFilter) ([]*automation.TestExecution, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	return uc.executions.List(ctx, f)
}

// ExecutionByBuildUID 按 build_uid 查询单条执行。
func (uc *QueryUseCase) ExecutionByBuildUID(ctx context.Context, buildUID uuid.UUID) (*automation.TestExecution, error) {
	return uc.executions.GetByBuildUID(ctx, buildUID)
}

// LiveSnapshot 是会话页首屏：execution + 用例摘要 + 当前 running 用例的步骤树。
type LiveSnapshot struct {
	Execution   *automation.TestExecution
	Items       []*automation.ExecutionItem
	CurrentItem *automation.ExecutionItem
}

// Live 组装 live 快照。items 不带步骤；current_item 带步骤树（无 running 则为 nil）。
func (uc *QueryUseCase) Live(ctx context.Context, buildUID uuid.UUID) (*LiveSnapshot, error) {
	exec, err := uc.executions.GetByBuildUID(ctx, buildUID)
	if err != nil {
		return nil, err
	}
	items, err := uc.items.ListByBuildUID(ctx, buildUID)
	if err != nil {
		return nil, err
	}
	current := pickCurrentRunningItem(items)
	if current != nil {
		if err := uc.attachStepsAndAttachments(ctx, current); err != nil {
			return nil, err
		}
	}
	return &LiveSnapshot{Execution: exec, Items: items, CurrentItem: current}, nil
}

// pickCurrentRunningItem 取 status=running 且 start_time 最新的一条。
// v1 不应出现多条 running；若出现，取最新 start_time，再比 ID。
func pickCurrentRunningItem(items []*automation.ExecutionItem) *automation.ExecutionItem {
	var current *automation.ExecutionItem
	for _, item := range items {
		if item.Status() != automation.CaseStatusRunning {
			continue
		}
		if current == nil || runningItemNewer(item, current) {
			current = item
		}
	}
	return current
}

func runningItemNewer(a, b *automation.ExecutionItem) bool {
	as, bs := a.StartTime(), b.StartTime()
	if as != nil && bs != nil {
		if as.After(*bs) {
			return true
		}
		if as.Before(*bs) {
			return false
		}
	} else if as != nil && bs == nil {
		return true
	} else if as == nil && bs != nil {
		return false
	}
	return a.ID() > b.ID()
}

// ItemByCaseUID 按 case_uid 查询单条用例（含步骤树与附件）。
func (uc *QueryUseCase) ItemByCaseUID(ctx context.Context, caseUID uuid.UUID) (*automation.ExecutionItem, error) {
	item, err := uc.items.GetByCaseUID(ctx, caseUID)
	if err != nil {
		return nil, err
	}
	if err := uc.attachStepsAndAttachments(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// ItemByID 按主键查询单条用例（不含步骤树与附件，用于步骤上报的归属校验）。
func (uc *QueryUseCase) ItemByID(ctx context.Context, id int64) (*automation.ExecutionItem, error) {
	return uc.items.GetByID(ctx, id)
}

// ExecutionDetail 查询单条执行（不含用例列表）。
func (uc *QueryUseCase) ExecutionDetail(ctx context.Context, id int64) (*automation.TestExecution, error) {
	exec, err := uc.executions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

// ListExecutionItems 查询一次执行下的用例列表（摘要，不带步骤树）。
func (uc *QueryUseCase) ListExecutionItems(
	ctx context.Context,
	buildUID uuid.UUID,
	page, pageSize int,
) ([]*automation.ExecutionItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	if _, err := uc.executions.GetByBuildUID(ctx, buildUID); err != nil {
		return nil, 0, err
	}

	items, err := uc.items.ListByBuildUID(ctx, buildUID)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

// ListAttempts 查询某用例在**一次执行内**的多次 attempt（D-03）。
//
// Python 侧通过 `WHERE case_key = ? AND build_uid = ?` 反查（缺陷 D-04 保持现状）。
// 返回按 attempt_number 升序。
func (uc *QueryUseCase) ListAttempts(ctx context.Context, buildUID uuid.UUID, caseKey string) ([]*automation.ExecutionItem, error) {
	return uc.items.ListByBuildUIDAndCaseKey(ctx, buildUID, caseKey)
}

// ListCasesHistory 查询某用例跨 execution 的历史结果（D-05）。
// 每个 execution 只取该用例的最终 attempt（is_latest = true）。
func (uc *QueryUseCase) ListCasesHistory(ctx context.Context, caseKey string) ([]*automation.ExecutionItem, error) {
	return uc.items.ListHistoryByCaseKey(ctx, caseKey)
}

// ListExecutionItemsAll 返回一次执行的全部用例（不分页），供 dashboard 投影等内部使用。
func (uc *QueryUseCase) ListExecutionItemsAll(ctx context.Context, buildUID uuid.UUID) ([]*automation.ExecutionItem, error) {
	return uc.items.ListByBuildUID(ctx, buildUID)
}

// attachStepsAndAttachments 给单个用例填充其步骤树与附件。
func (uc *QueryUseCase) attachStepsAndAttachments(ctx context.Context, item *automation.ExecutionItem) error {
	steps, err := uc.steps.ListByCaseUID(ctx, item.CaseUID())
	if err != nil {
		return err
	}
	item.SetSteps(steps)

	attachments, err := uc.attachments.ListByItemID(ctx, item.ID())
	if err != nil {
		return err
	}
	item.SetAttachments(attachments)

	// 每个步骤也带附件（step 级，Q1）
	for _, step := range steps {
		stepAtts, err := uc.attachments.ListByStepID(ctx, step.ID())
		if err != nil {
			return err
		}
		step.SetAttachments(stepAtts)
	}
	return nil
}
