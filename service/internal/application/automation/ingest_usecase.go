package automation

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// 命令
// ---------------------------------------------------------------------------

// IngestExecutionCommand 是上报一次执行开始（pytest_sessionstart）的请求。
type IngestExecutionCommand struct {
	BuildUID          uuid.UUID
	JobName           string
	JobURL            string
	ProjectName       string
	SoftwareName      string
	SoftwareVersion   string
	Labels            []string
	PlannedCasesCount int
	StartTime         time.Time
}

// FinishExecutionCommand 是上报一次执行结束的请求。
type FinishExecutionCommand struct {
	BuildUID uuid.UUID
	Status   automation.ExecutionStatus
	EndTime  time.Time
	Labels   []string
}

// IngestItemCommand 是上报一条用例开始/结束的请求。
type IngestItemCommand struct {
	BuildUID       uuid.UUID
	CaseUID        uuid.UUID
	CaseKey        string
	CaseName       string
	Labels         []string
	Status         automation.CaseStatus
	StartTime      *time.Time
	EndTime        *time.Time
	ErrorMessage   string
	ErrorTraceback string
	Attachments    []automation.AttachmentSpec
}

// IngestStepCommand 是上报一个步骤的请求。
type IngestStepCommand struct {
	CaseUID  uuid.UUID
	StepPath automation.StepPath
	StepName string
	Status   automation.StepStatus
	Start    *time.Time
	End      *time.Time
	Duration *float64
}

// ---------------------------------------------------------------------------
// 用例
// ---------------------------------------------------------------------------

// IngestUseCase 编排上报类用例（T-3.8 / T-3.9 / T-3.10）。
//
// 所有方法都返回**发布过的事件**，供 handler 经 EventPublisher 派发。
// 事件是在写完成后收集的，不用在每个写方法里各自调 publisher——
// 这样如果事件发布失败，业务写入不会回滚（写侧与读模型解耦）。
type IngestUseCase struct {
	executions  ExecutionRepository
	items       ItemRepository
	steps       StepRepository
	attachments AttachmentRepository
	pipelines   PipelineRepository
	clock       Clock
	live        LivePublisher
}

type noopLivePublisher struct{}

func (noopLivePublisher) Publish(context.Context, automation.Event) error { return nil }

// NewIngestUseCase 构造上报用例。live 为 nil 时使用空发布器（测试 / worker 扫描）。
func NewIngestUseCase(
	executions ExecutionRepository,
	items ItemRepository,
	steps StepRepository,
	attachments AttachmentRepository,
	pipelines PipelineRepository,
	clock Clock,
	live LivePublisher,
) *IngestUseCase {
	if live == nil {
		live = noopLivePublisher{}
	}
	return &IngestUseCase{
		executions:  executions,
		items:       items,
		steps:       steps,
		attachments: attachments,
		pipelines:   pipelines,
		clock:       clock,
		live:        live,
	}
}

// IngestExecution 上报一次执行开始。
//
// 幂等：同一 build_uid 重复上报时复用已有记录，不报错（缺陷 D-13）。
func (uc *IngestUseCase) IngestExecution(ctx context.Context, cmd IngestExecutionCommand) ([]automation.Event, error) {
	start := cmd.StartTime
	if start.IsZero() {
		start = uc.clock.Now()
	}

	exec, err := automation.NewExecution(
		cmd.BuildUID, cmd.JobName, cmd.JobURL,
		cmd.ProjectName, cmd.SoftwareName, cmd.SoftwareVersion,
		cmd.Labels, cmd.PlannedCasesCount, start,
	)
	if err != nil {
		return nil, err
	}

	_, created, err := uc.executions.Create(ctx, exec)
	if err != nil {
		return nil, err
	}

	// 每条流水线记录要同步（job_name 幂等）。
	// 新 execution 才建；已存在则用最新信息刷新它。
	if created {
		if err := uc.syncPipeline(ctx, exec); err != nil {
			// 流水线同步失败不应让 execution 上报失败
			return nil, err
		}
	}

	var events []automation.Event
	if created {
		events = append(events, automation.ExecutionCreated{BuildUID: cmd.BuildUID})
	}
	return events, nil
}

// FinishExecution 上报一次执行结束。
//
// 幂等：已进入终止态的 execution 再次收到 running 时静默忽略（Execution.Finish）。
// 计数回填（评审确认项 Q2）：仅在 execution **首次**进入终止态时聚合一次。
func (uc *IngestUseCase) FinishExecution(ctx context.Context, cmd FinishExecutionCommand) ([]automation.Event, error) {
	exec, err := uc.executions.GetByBuildUID(ctx, cmd.BuildUID)
	if err != nil {
		return nil, err
	}

	if cmd.Labels != nil {
		exec.AddLabels(cmd.Labels)
	}

	end := cmd.EndTime
	if end.IsZero() {
		end = uc.clock.Now()
	}

	// 返回 false 表示幂等忽略（已是终止态又收到 running 或重复结束），不发事件
	changed, err := exec.Finish(cmd.Status, end)
	if err != nil {
		return nil, err
	}

	var events []automation.Event
	if changed {
		// 首次进入终止态：聚合计数回填（评审确认项 Q2）
		if err := uc.refreshCounts(ctx, exec); err != nil {
			return nil, err
		}

		if err := uc.executions.UpdateStatus(ctx, exec); err != nil {
			return nil, err
		}

		if err := uc.syncPipeline(ctx, exec); err != nil {
			return nil, err
		}

		ev := automation.ExecutionUpdated{
			ExecutionID: exec.ID(),
			BuildUID:    exec.BuildUID(),
			Status:      string(exec.Status()),
		}
		events = append(events, ev)
		uc.publishLive(ctx, ev)
	}
	return events, nil
}

// IngestItem 上报一条用例。
//
// 语义：
//   - 若 case_uid 已存在（同一 attempt 再次上报），更新它；
//   - 否则新建（attempt_number 由 NextAttemptNumber 原子分配）。
//
// 状态为 running 时创建；为终止态时创建并立即结束。
func (uc *IngestUseCase) IngestItem(ctx context.Context, cmd IngestItemCommand) ([]automation.Event, error) {
	// 先尝试按 case_uid 找到已有记录（同一 attempt 的重复上报）
	existing, err := uc.items.GetByCaseUID(ctx, cmd.CaseUID)
	if err != nil && !errors.Is(err, errors.KindNotFound) {
		return nil, err
	}

	var item *automation.ExecutionItem
	created := false
	if existing != nil {
		item = existing
	} else {
		item, err = automation.NewItem(automation.NewItemCommand{
			BuildUID: cmd.BuildUID,
			CaseUID:  cmd.CaseUID,
			CaseKey:  cmd.CaseKey,
			CaseName: cmd.CaseName,
			Labels:   cmd.Labels,
			// 占位值 1：真实序号由 CreateWithNextAttempt 在事务内
			// 通过 pg_advisory_xact_lock 分配后 SetAttemptNumber 覆盖。
			AttemptNumber: 1,
			StartTime:     cmd.StartTime,
		})
		if err != nil {
			return nil, err
		}

		// 原子创建：在单个事务内分配 attempt 序号并插入（并发安全，
		// 锁与 INSERT 同事务，不会撞唯一约束 uq_execution_items_attempt）。
		id, err := uc.items.CreateWithNextAttempt(ctx, cmd.BuildUID, item)
		if err != nil {
			return nil, err
		}
		// 重建实体以拿到数据库生成的 ID 与分配的 attempt_number
		item, err = uc.items.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		created = true
	}

	if cmd.Labels != nil {
		item.AddLabels(cmd.Labels)
	}
	if cmd.StartTime != nil {
		item.SetStartTime(cmd.StartTime)
	}

	var events []automation.Event

	// 状态为终止态时结束用例
	if cmd.Status.IsTerminal() {
		changed, err := item.Finish(cmd.Status, cmd.EndTime, cmd.ErrorMessage, cmd.ErrorTraceback)
		if err != nil {
			return nil, err
		}
		if changed {
			// 结束用例：持久化。已存在则更新，新建则已在上面的 Create 写入，这里再补状态
			if err := uc.items.Update(ctx, item); err != nil {
				return nil, err
			}
		} else if created {
			// 新建即终止（很少见）：状态在创建后已写入，无需额外操作
		}
	} else if !created {
		// running 状态但已存在的记录：只更新（可能只是补充标签/开始时间）
		if err := uc.items.Update(ctx, item); err != nil {
			return nil, err
		}
	}

	// 处理附件（Q1：case 级记 item_id）
	if len(cmd.Attachments) > 0 {
		if err := uc.registerItemAttachments(ctx, item.ID(), cmd.Attachments); err != nil {
			return nil, err
		}
	}

	// 解析该用例所属 execution 的 ID（用于驱动单 execution 快照刷新）
	var executionID int64
	if exec, err := uc.executions.GetByBuildUID(ctx, cmd.BuildUID); err == nil {
		executionID = exec.ID()
	}

	ev := itemUpdatedEvent(item, executionID)
	events = append(events, ev)
	uc.publishLive(ctx, ev)
	return events, nil
}

// Heartbeat 刷新一次 running execution 的存活时间。
// 已终止的 execution 静默忽略，避免 sessionfinish 与心跳竞态让插件失败。
func (uc *IngestUseCase) Heartbeat(ctx context.Context, buildUID uuid.UUID) (time.Time, error) {
	exec, err := uc.executions.GetByBuildUID(ctx, buildUID)
	if err != nil {
		return time.Time{}, err
	}

	now := uc.clock.Now()
	if !exec.TouchHeartbeat(now) {
		if at := exec.LastHeartbeatAt(); at != nil {
			return *at, nil
		}
		return now, nil
	}
	if err := uc.executions.UpdateHeartbeat(ctx, exec); err != nil {
		return time.Time{}, err
	}
	return now, nil
}

// AbortStale 把心跳早于 olderThan 的 running execution 标为 aborted。
func (uc *IngestUseCase) AbortStale(ctx context.Context, olderThan time.Duration) (int, []automation.Event, error) {
	if olderThan <= 0 {
		olderThan = time.Minute
	}
	cutoff := uc.clock.Now().Add(-olderThan)
	stale, err := uc.executions.ListStaleRunning(ctx, cutoff)
	if err != nil {
		return 0, nil, err
	}

	var events []automation.Event
	aborted := 0
	for _, exec := range stale {
		changed, err := exec.Finish(automation.ExecutionStatusAborted, uc.clock.Now())
		if err != nil {
			return aborted, events, err
		}
		if !changed {
			continue
		}
		if err := uc.refreshCounts(ctx, exec); err != nil {
			return aborted, events, err
		}
		if err := uc.executions.UpdateStatus(ctx, exec); err != nil {
			return aborted, events, err
		}
		if err := uc.syncPipeline(ctx, exec); err != nil {
			return aborted, events, err
		}
		aborted++
		ev := automation.ExecutionUpdated{
			ExecutionID: exec.ID(),
			BuildUID:    exec.BuildUID(),
			Status:      string(exec.Status()),
		}
		events = append(events, ev)
		uc.publishLive(ctx, ev)
	}
	return aborted, events, nil
}

// IngestStep 上报一个步骤。
//
// 幂等：以 (case_uid, step_path) 为键 upsert，重复上报覆盖（D-08）。
func (uc *IngestUseCase) IngestStep(ctx context.Context, cmd IngestStepCommand) ([]automation.Event, error) {
	item, err := uc.items.GetByCaseUID(ctx, cmd.CaseUID)
	if err != nil {
		return nil, err
	}

	step, err := automation.NewStep(automation.NewStepCommand{
		CaseUID:   cmd.CaseUID,
		StepPath:  cmd.StepPath,
		StepName:  cmd.StepName,
		Status:    cmd.Status,
		StartTime: cmd.Start,
		EndTime:   cmd.End,
		Duration:  cmd.Duration,
	})
	if err != nil {
		return nil, err
	}

	if parentPath, ok := cmd.StepPath.Parent(); ok {
		parentIndex := parentPath.LastSegment()
		step.SetParent(nil, &parentIndex)
	}

	if _, err := uc.steps.Upsert(ctx, step); err != nil {
		return nil, err
	}

	liveEvent := automation.StepUpserted{
		BuildUID: item.BuildUID(),
		CaseUID:  cmd.CaseUID,
		CaseKey:  item.CaseKey(),
		CaseName: item.CaseName(),
		StepPath: cmd.StepPath.String(),
		StepName: cmd.StepName,
		Status:   string(cmd.Status),
	}
	uc.publishLive(ctx, liveEvent)
	return []automation.Event{liveEvent}, nil
}

func (uc *IngestUseCase) publishLive(ctx context.Context, event automation.Event) {
	_ = uc.live.Publish(ctx, event)
}

func itemUpdatedEvent(item *automation.ExecutionItem, executionID int64) automation.ItemUpdated {
	return automation.ItemUpdated{
		ExecutionID:   executionID,
		BuildUID:      item.BuildUID(),
		ItemID:        item.ID(),
		CaseUID:       item.CaseUID(),
		CaseKey:       item.CaseKey(),
		CaseName:      item.CaseName(),
		AttemptNumber: item.AttemptNumber(),
		Status:        string(item.Status()),
		StartTime:     item.StartTime(),
		EndTime:       item.EndTime(),
		Duration:      item.Duration(),
		ErrorMessage:  item.ErrorMessage(),
	}
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// refreshCounts 按该 build_uid 下最新用例（is_latest = true）聚合计数回填。
func (uc *IngestUseCase) refreshCounts(ctx context.Context, exec *automation.TestExecution) error {
	items, err := uc.items.ListByBuildUIDLatest(ctx, exec.BuildUID())
	if err != nil {
		return err
	}

	var pass, failure, skipped int
	for _, item := range items {
		switch item.Status() {
		case automation.CaseStatusPassed:
			pass++
		case automation.CaseStatusFailed, automation.CaseStatusBroken:
			failure++
		case automation.CaseStatusSkipped:
			skipped++
		}
	}

	exec.RefreshCounts(pass, failure, skipped)
	return nil
}

// syncPipeline 用一次执行的最新信息刷新流水线记录。
func (uc *IngestUseCase) syncPipeline(ctx context.Context, exec *automation.TestExecution) error {
	pipeline, err := uc.pipelines.GetByJobName(ctx, exec.JobName())
	if err != nil && !errors.Is(err, errors.KindNotFound) {
		return err
	}
	if pipeline == nil {
		pipeline, err = automation.NewPipeline(exec.JobName(), exec.JobURL())
		if err != nil {
			return err
		}
	}
	pipeline.UpdateLastBuild(exec)
	_, err = uc.pipelines.Upsert(ctx, pipeline)
	return err
}

func (uc *IngestUseCase) registerItemAttachments(ctx context.Context, itemID int64, specs []automation.AttachmentSpec) error {
	for _, spec := range specs {
		att, err := automation.NewAttachment(automation.NewAttachmentCommand{
			ItemID:   &itemID,
			Type:     spec.Type,
			FileName: spec.FileName,
			URL:      spec.URL,
			MimeType: spec.MimeType,
		})
		if err != nil {
			return err
		}
		if _, err := uc.attachments.Create(ctx, att); err != nil {
			return err
		}
	}
	return nil
}
