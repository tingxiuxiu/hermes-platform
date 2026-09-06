// Package automation 承载 automation 上下文的应用层：上报与查询用例编排（ADR-0004）。
package automation

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/domain/automation"
)

// ---------------------------------------------------------------------------
// 持久化端口
// ---------------------------------------------------------------------------

// ExecutionRepository 是 execution 聚合的持久化端口。
type ExecutionRepository interface {
	// Create 写入一条执行。build_uid 冲突时**不报错**，返回已存在记录的 ID 与 false。
	// 这实现了 D-13（幂等创建）：插件重复上报同一 build_uid 时静默复用已有记录。
	Create(ctx context.Context, e *automation.TestExecution) (id int64, created bool, err error)

	// GetByBuildUID 按 build_uid 读取。
	GetByBuildUID(ctx context.Context, buildUID uuid.UUID) (*automation.TestExecution, error)

	// GetByID 按主键读取。
	GetByID(ctx context.Context, id int64) (*automation.TestExecution, error)

	// UpdateStatus 结束一条执行并回填计数。
	// 仅当 e.Status() 是终止态时由用例层调用；运行中的执行不应走到这里。
	UpdateStatus(ctx context.Context, e *automation.TestExecution) error

	// UpdateHeartbeat 刷新 last_heartbeat_at（仅 running）。
	UpdateHeartbeat(ctx context.Context, e *automation.TestExecution) error

	// ListStaleRunning 返回心跳早于 cutoff 的 running execution。
	ListStaleRunning(ctx context.Context, cutoff time.Time) ([]*automation.TestExecution, error)

	// List 分页查询。f 的筛选条件见 automation/ports 的 ListFilter。
	List(ctx context.Context, f ExecutionListFilter) ([]*automation.TestExecution, int64, error)

	// LatestByBuildUID 按时间倒序返回同一 build_uid 的所有执行（供重试会话合并）。
	LatestByBuildUID(ctx context.Context, buildUID uuid.UUID) ([]*automation.TestExecution, error)
}

// ExecutionListFilter 是执行列表的查询条件，对应 GET /automation/executions。
type ExecutionListFilter struct {
	BuildUID  *uuid.UUID
	JobName   string
	Project   string
	Status    *automation.ExecutionStatus
	Labels    []string
	StartFrom *time.Time
	StartTo   *time.Time
	Page      int
	PageSize  int
}

// ItemRepository 是用例聚合的持久化端口。
type ItemRepository interface {
	// Create 写入一条用例。返回数据库生成的 ID。
	Create(ctx context.Context, item *automation.ExecutionItem) (int64, error)

	// GetByCaseUID 按 case_uid 读取。
	GetByCaseUID(ctx context.Context, caseUID uuid.UUID) (*automation.ExecutionItem, error)

	// GetByID 按主键读取。
	GetByID(ctx context.Context, id int64) (*automation.ExecutionItem, error)

	// AttemptCount 返回某 build_uid + case_key 下已有的 attempt 数量。
	// 用于计算下一次 attempt_number（同一事务内，配合 Create）。
	AttemptCount(ctx context.Context, buildUID uuid.UUID, caseKey string) (int, error)

	// NextAttemptNumber 原子地取「下一个 attempt 序号」并立即占位。
	//
	// 这是 T-4.3 的并发关键：多个 pytest 进程可能同时为同一用例上报新 attempt，
	// 若用「先查 count 再插入」会撞唯一约束 `uq_execution_items_attempt`。
	// 用 pg_advisory_xact_lock 保证同一 build_uid+case_key 的 attempt 序号串行分配。
	NextAttemptNumber(ctx context.Context, buildUID uuid.UUID, caseKey string) (int, error)

	// CreateWithNextAttempt 在**单个事务**内分配 attempt 序号并插入用例。
	//
	// 这是并发的正确实现：pg_advisory_xact_lock 必须在与 INSERT 相同的事务内
	// 持有，锁才会在事务提交前一直生效。若锁与插入分离（NextAttemptNumber +
	// Create 两次独立调用），锁在 Exec 返回即释放，并发仍会撞唯一约束。
	CreateWithNextAttempt(ctx context.Context, buildUID uuid.UUID, item *automation.ExecutionItem) (int64, error)

	// Update 更新一条用例（状态、错误信息、标签等）。
	Update(ctx context.Context, item *automation.ExecutionItem) error

	// ListByBuildUID 返回某执行的全部用例（is_latest 任意），按 attempt 排序。
	ListByBuildUID(ctx context.Context, buildUID uuid.UUID) ([]*automation.ExecutionItem, error)

	// ListByBuildUIDAndCaseKey 返回某执行内某用例的全部 attempt，按 attempt_number 升序。
	ListByBuildUIDAndCaseKey(ctx context.Context, buildUID uuid.UUID, caseKey string) ([]*automation.ExecutionItem, error)

	// ListHistoryByCaseKey 返回某用例跨 execution 的历史结果。
	// 每个 execution 只取最终 attempt（is_latest = true），按时间倒序。
	ListHistoryByCaseKey(ctx context.Context, caseKey string) ([]*automation.ExecutionItem, error)

	// ListByBuildUIDLatest 返回某执行的**最新**用例（is_latest = true），用于汇总计数。
	ListByBuildUIDLatest(ctx context.Context, buildUID uuid.UUID) ([]*automation.ExecutionItem, error)
}

// StepRepository 是步骤聚合的持久化端口。
type StepRepository interface {
	// Upsert 幂等写入/更新一个步骤。
	// 以 (case_uid, step_path) 为幂等键：重复上报覆盖，不插入。
	// 返回是否为新插入（用于决定是否触发事件）。
	Upsert(ctx context.Context, step *automation.ExecutionStep) (created bool, err error)

	// ListByCaseUID 返回某用例的全部步骤，按 step_path 排序。
	ListByCaseUID(ctx context.Context, caseUID uuid.UUID) ([]*automation.ExecutionStep, error)
}

// AttachmentRepository 是附件记录的持久化端口。
type AttachmentRepository interface {
	// Create 写入一条附件记录。
	Create(ctx context.Context, a *automation.Attachment) (int64, error)

	// ListByItemID 返回某用例的全部附件（case 级，item_id 匹配）。
	ListByItemID(ctx context.Context, itemID int64) ([]*automation.Attachment, error)

	// ListByStepID 返回某步骤的全部附件（step 级）。
	ListByStepID(ctx context.Context, stepID int64) ([]*automation.Attachment, error)

	// ListByStepIDs 返回多个步骤的全部附件（step 级）。
	ListByStepIDs(ctx context.Context, stepIDs []int64) ([]*automation.Attachment, error)
}

// PipelineRepository 是流水线的持久化端口。
type PipelineRepository interface {
	// Upsert 按 job_name 幂等写入流水线。返回是否为新插入。
	Upsert(ctx context.Context, p *automation.JenkinsPipeline) (created bool, err error)

	// GetByJobName 按任务名读取。
	GetByJobName(ctx context.Context, jobName string) (*automation.JenkinsPipeline, error)

	// List 分页查询（模糊 job_name + 状态筛选）。
	List(ctx context.Context, f PipelineListFilter) ([]*automation.JenkinsPipeline, int64, error)
}

// PipelineListFilter 是流水线列表的查询条件，对应 GET /automation/pipelines。
type PipelineListFilter struct {
	JobName  string
	Status   *automation.PipelineStatus
	Page     int
	PageSize int
}

// ---------------------------------------------------------------------------
// 事件发布端口
// ---------------------------------------------------------------------------

// EventPublisher 发布领域事件。
// 由 bootstrap 层的桥接器实现（把事件映射为 asynq 任务），automation 层不感知 asynq。
type EventPublisher interface {
	Publish(ctx context.Context, event automation.Event) error
}

// LivePublisher 把步骤级变更推给正在看会话的客户端（Redis / 测试 fake）。
// 与 EventPublisher 分开，避免 asynq Unique(5m) 吞掉 1 秒级步骤更新。
type LivePublisher interface {
	Publish(ctx context.Context, event automation.Event) error
}

// LiveFrame 是一条 SSE 帧。
type LiveFrame struct {
	Event string
	JSON  []byte
}

// LiveSubscriber 订阅某次 execution 的 live 通道。
type LiveSubscriber interface {
	Subscribe(ctx context.Context, buildUID uuid.UUID) (<-chan LiveFrame, error)
}

// Clock 是时间源端口，由 platform/clock 实现。
type Clock interface {
	Now() time.Time
}
