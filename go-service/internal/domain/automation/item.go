package automation

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// CaseStatus 是用例（以及步骤）的状态。
//
// 取值与 Python 的 `CaseStatus = Literal["running","passed","failed","skipped","broken"]`
// 完全一致。StepStatus 在 Python 侧就是 CaseStatus 的别名，这里同样复用。
type CaseStatus string

const (
	CaseStatusRunning CaseStatus = "running"
	CaseStatusPassed  CaseStatus = "passed"
	CaseStatusFailed  CaseStatus = "failed"
	CaseStatusSkipped CaseStatus = "skipped"
	CaseStatusBroken  CaseStatus = "broken"
)

// StepStatus 是步骤状态，取值与 CaseStatus 相同。
type StepStatus = CaseStatus

// Valid 判断状态值是否合法。
func (s CaseStatus) Valid() bool {
	switch s {
	case CaseStatusRunning, CaseStatusPassed, CaseStatusFailed,
		CaseStatusSkipped, CaseStatusBroken:
		return true
	default:
		return false
	}
}

// IsTerminal 判断是否终止态（running 之外的都是终止态）。
func (s CaseStatus) IsTerminal() bool { return s != CaseStatusRunning }

// ParseCaseStatus 解析状态值。
func ParseCaseStatus(raw string) (CaseStatus, error) {
	s := CaseStatus(strings.TrimSpace(raw))
	if !s.Valid() {
		return "", errors.Errorf(errors.KindValidation,
			"非法的用例状态: %q（允许 running/passed/failed/skipped/broken）", raw)
	}
	return s, nil
}

// ParseStepStatus 解析步骤状态值。
func ParseStepStatus(raw string) (StepStatus, error) { return ParseCaseStatus(raw) }

// ExecutionItem 是一次用例执行记录。
//
// 关键领域规则（缺陷 D-02，评审确认保持现状）：
//   - 每个 attempt 生成**新的** case_uid，但 case_key 不变；
//   - 新建 attempt 时**不翻转**旧 attempt 的 is_latest。
//
// 这让「某次执行中已结束的用例」查询必须靠 `is_latest = false` 反查
// （缺陷 D-03，同样保持现状），需要在用例层与仓储层各自注释清楚，
// 避免后人误以为漏写了翻转逻辑。
type ExecutionItem struct {
	id             int64
	buildUID       uuid.UUID
	caseUID        uuid.UUID
	caseKey        string
	caseName       string
	labels         []string
	attemptNumber  int
	isLatest       bool
	status         CaseStatus
	startTime      *time.Time
	endTime        *time.Time
	duration       *float64
	errorMessage   string
	errorTraceback string
	createdAt      time.Time
	updatedAt      time.Time

	// steps / attachments 是查询时的读模型扩展，不属于持久化字段。
	// 用例层查询接口填充它们，供 DTO 层直接读取。
	steps       []*ExecutionStep
	attachments []*Attachment
}

// ItemSnapshot 用于从数据库重建用例。
type ItemSnapshot struct {
	ID             int64
	BuildUID       uuid.UUID
	CaseUID        uuid.UUID
	CaseKey        string
	CaseName       string
	Labels         []string
	AttemptNumber  int
	IsLatest       bool
	Status         string
	StartTime      *time.Time
	EndTime        *time.Time
	Duration       *float64
	ErrorMessage   string
	ErrorTraceback string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewItemCommand 是创建用例所需的输入。
type NewItemCommand struct {
	BuildUID uuid.UUID
	CaseUID  uuid.UUID
	CaseKey  string
	CaseName string
	Labels   []string
	// AttemptNumber 由用例层根据「同 build_uid + case_key 已有 attempt 数」计算后传入，
	// 实体本身不做自增——自增需要查询，属于用例/仓储职责。
	AttemptNumber int
	StartTime     *time.Time
}

// NewItem 创建一条用例记录。
func NewItem(cmd NewItemCommand) (*ExecutionItem, error) {
	item := &ExecutionItem{}

	if cmd.BuildUID == uuid.Nil {
		return nil, ErrInvalidBuildUID
	}
	item.buildUID = cmd.BuildUID

	if cmd.CaseUID == uuid.Nil {
		return nil, ErrInvalidCaseUID
	}
	item.caseUID = cmd.CaseUID

	key := strings.TrimSpace(cmd.CaseKey)
	if key == "" {
		return nil, ErrCaseKeyRequired
	}
	if len(key) > 128 {
		return nil, errors.Errorf(errors.KindValidation, "case_key 长度不能超过 128 个字符")
	}
	item.caseKey = key

	name := strings.TrimSpace(cmd.CaseName)
	if name == "" {
		return nil, ErrCaseNameRequired
	}
	item.caseName = name

	item.labels = normaliseLabels(cmd.Labels)

	if cmd.AttemptNumber < 1 {
		return nil, ErrInvalidAttemptNumber
	}
	item.attemptNumber = cmd.AttemptNumber

	item.isLatest = true
	item.status = CaseStatusRunning
	item.startTime = cmd.StartTime
	return item, nil
}

// RestoreItem 从数据库重建实体。仅供持久化适配器使用。
func RestoreItem(s ItemSnapshot) *ExecutionItem {
	status := CaseStatus(s.Status)
	if !status.Valid() {
		status = CaseStatusRunning
	}
	return &ExecutionItem{
		id:             s.ID,
		buildUID:       s.BuildUID,
		caseUID:        s.CaseUID,
		caseKey:        s.CaseKey,
		caseName:       s.CaseName,
		labels:         s.Labels,
		attemptNumber:  s.AttemptNumber,
		isLatest:       s.IsLatest,
		status:         status,
		startTime:      s.StartTime,
		endTime:        s.EndTime,
		duration:       s.Duration,
		errorMessage:   s.ErrorMessage,
		errorTraceback: s.ErrorTraceback,
		createdAt:      s.CreatedAt,
		updatedAt:      s.UpdatedAt,
	}
}

// ---- 读取器 ----

func (i *ExecutionItem) ID() int64           { return i.id }
func (i *ExecutionItem) BuildUID() uuid.UUID { return i.buildUID }
func (i *ExecutionItem) CaseUID() uuid.UUID  { return i.caseUID }
func (i *ExecutionItem) CaseKey() string     { return i.caseKey }
func (i *ExecutionItem) CaseName() string    { return i.caseName }
func (i *ExecutionItem) Labels() []string    { return i.labels }
func (i *ExecutionItem) AttemptNumber() int  { return i.attemptNumber }

// SetAttemptNumber 设置 attempt 序号。
// 仅供仓储在事务内分配序号后回填（CreateWithNextAttempt）。
func (i *ExecutionItem) SetAttemptNumber(n int) { i.attemptNumber = n }

func (i *ExecutionItem) IsLatest() bool         { return i.isLatest }
func (i *ExecutionItem) Status() CaseStatus     { return i.status }
func (i *ExecutionItem) StartTime() *time.Time  { return i.startTime }
func (i *ExecutionItem) EndTime() *time.Time    { return i.endTime }
func (i *ExecutionItem) Duration() *float64     { return i.duration }
func (i *ExecutionItem) ErrorMessage() string   { return i.errorMessage }
func (i *ExecutionItem) ErrorTraceback() string { return i.errorTraceback }
func (i *ExecutionItem) CreatedAt() time.Time   { return i.createdAt }
func (i *ExecutionItem) UpdatedAt() time.Time   { return i.updatedAt }

// Steps 返回查询时填充的步骤列表（读模型扩展）。
func (i *ExecutionItem) Steps() []*ExecutionStep { return i.steps }

// Attachments 返回查询时填充的附件列表（读模型扩展）。
func (i *ExecutionItem) Attachments() []*Attachment { return i.attachments }

// SetSteps 填充步骤列表（查询用例层调用）。
func (i *ExecutionItem) SetSteps(steps []*ExecutionStep) { i.steps = steps }

// SetAttachments 填充附件列表（查询用例层调用）。
func (i *ExecutionItem) SetAttachments(atts []*Attachment) { i.attachments = atts }

// ---- 行为 ----

// Finish 结束一条用例。
//
// 幂等规则同 execution：已终止的用例不再接受任何结束操作，
// 以**首次结束**为准，避免重复上报覆盖已计算的 duration。
func (i *ExecutionItem) Finish(
	status CaseStatus,
	endTime *time.Time,
	errorMessage, errorTraceback string,
) (bool, error) {
	if !status.Valid() {
		return false, ErrInvalidStatus("用例", string(status))
	}
	if i.status.IsTerminal() {
		return false, nil
	}
	if status == CaseStatusRunning {
		return false, nil
	}

	i.status = status
	i.errorMessage = strings.TrimSpace(errorMessage)
	i.errorTraceback = strings.TrimSpace(errorTraceback)

	if endTime != nil {
		t := endTime.UTC()
		i.endTime = &t
		if i.startTime != nil {
			d := t.Sub(*i.startTime).Seconds()
			i.duration = &d
		}
	}
	return true, nil
}

// SetStartTime 补记开始时间。插件在 session 开始时可能还没有用例粒度的开始时间。
func (i *ExecutionItem) SetStartTime(t *time.Time) {
	if t == nil {
		return
	}
	v := t.UTC()
	i.startTime = &v
}

// AddLabels 追加标签并去重。
func (i *ExecutionItem) AddLabels(labels []string) {
	i.labels = mergeLabels(i.labels, normaliseLabels(labels))
}
