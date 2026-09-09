package automation

import (
	"time"

	"github.com/google/uuid"
)

// 领域事件。
//
// automation 层**不感知 asynq**（ADR-0004）——它只发布领域事件，
// 由 bootstrap 层的桥接器（bootstrap/events.go）把事件映射成 asynq 任务。
// 这正是「写侧解耦读侧（Dashboard CQRS）」的关键：上报链路不知道也不关心
// 读模型怎么重算。

// Event 是所有领域事件的标记接口。
// 使用 marker 而非让每个事件实现一整套方法，是为了让下游可以用 type switch
// 分派，同时避免过度设计。
type Event interface {
	// eventKind 返回事件类型标识，用于日志与排队的队列路由。
	eventKind() string
}

// ExecutionCreated 在一次执行创建后发布。
type ExecutionCreated struct {
	BuildUID uuid.UUID
}

func (ExecutionCreated) eventKind() string { return "execution.created" }

// ExecutionUpdated 在一次执行状态/计数更新后发布。
type ExecutionUpdated struct {
	ExecutionID int64
	BuildUID    uuid.UUID
	Status      string
}

func (ExecutionUpdated) eventKind() string { return "execution.updated" }

// ItemUpdated 在一次用例创建或更新后发布。
type ItemUpdated struct {
	ExecutionID   int64
	BuildUID      uuid.UUID
	ItemID        int64
	CaseUID       uuid.UUID
	CaseKey       string
	CaseName      string
	AttemptNumber int
	Status        string
	StartTime     *time.Time
	EndTime       *time.Time
	Duration      *float64
	ErrorMessage  string
}

func (ItemUpdated) eventKind() string { return "item.updated" }

// StepUpserted 在步骤写入后发布，只走 live 通道，不进 asynq。
type StepUpserted struct {
	BuildUID uuid.UUID
	CaseUID  uuid.UUID
	CaseKey  string
	CaseName string
	StepPath string
	StepName string
	Status   string
}

func (StepUpserted) eventKind() string { return "step.upserted" }
