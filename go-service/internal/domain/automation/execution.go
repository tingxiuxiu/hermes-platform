// Package automation 承载 automation 上下文的领域模型。
//
// 该上下文是**写侧**：pytest 插件通过服务令牌上报执行过程。
// 所有状态枚举的取值都必须与 Python schemas.py 的 Literal 定义一致，
// 否则插件上报的状态会被拒（ADR-0001）。
package automation

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// ExecutionStatus 是一次执行的状态。
//
// 取值与 Python 的 `ExecutionStatus = Literal["running","completed","failed","aborted","calculated"]`
// 完全一致，不得增删。
type ExecutionStatus string

const (
	ExecutionStatusRunning    ExecutionStatus = "running"
	ExecutionStatusCompleted  ExecutionStatus = "completed"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusAborted    ExecutionStatus = "aborted"
	ExecutionStatusCalculated ExecutionStatus = "calculated" // 计算完成，ETL 已处理
)

// Valid 判断状态值是否合法。
func (s ExecutionStatus) Valid() bool {
	switch s {
	case ExecutionStatusRunning, ExecutionStatusCompleted,
		ExecutionStatusFailed, ExecutionStatusAborted, ExecutionStatusCalculated:
		return true
	default:
		return false
	}
}

// IsTerminal 判断是否终止态。
// 终止态的 execution 不再接受 running 状态回退（幂等忽略，见 Execution.Finish）。
func (s ExecutionStatus) IsTerminal() bool {
	return s != ExecutionStatusRunning
}

// ParseExecutionStatus 解析状态值。
func ParseExecutionStatus(raw string) (ExecutionStatus, error) {
	s := ExecutionStatus(strings.TrimSpace(raw))
	if !s.Valid() {
		return "", errors.Errorf(errors.KindValidation,
			"非法的 execution 状态: %q（允许 running/completed/failed/aborted/calculated）", raw)
	}
	return s, nil
}

// TestExecution 是一次 pytest 执行的聚合根。
type TestExecution struct {
	id                int64
	buildUID          uuid.UUID
	jobName           string
	jobURL            string
	projectName       string
	softwareName      string
	softwareVersion   string
	labels            []string
	status            ExecutionStatus
	startTime         time.Time
	endTime           *time.Time
	duration          *float64
	plannedCasesCount int
	passCount         int
	failureCount      int
	skippedCount      int
	passRate          *float64
	lastHeartbeatAt   *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

// ExecutionSnapshot 用于从数据库重建 execution。
type ExecutionSnapshot struct {
	ID                int64
	BuildUID          uuid.UUID
	JobName           string
	JobURL            string
	ProjectName       string
	SoftwareName      string
	SoftwareVersion   string
	Labels            []string
	Status            string
	StartTime         time.Time
	EndTime           *time.Time
	Duration          *float64
	PlannedCasesCount int
	PassCount         int
	FailureCount      int
	SkippedCount      int
	PassRate          *float64
	LastHeartbeatAt   *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewExecution 创建一次执行。创建时状态固定为 running。
func NewExecution(
	buildUID uuid.UUID,
	jobName, jobURL, projectName, softwareName, softwareVersion string,
	labels []string,
	plannedCasesCount int,
	startTime time.Time,
) (*TestExecution, error) {
	e := &TestExecution{}

	if buildUID == uuid.Nil {
		return nil, ErrInvalidBuildUID
	}
	e.buildUID = buildUID

	if err := e.setJobName(jobName); err != nil {
		return nil, err
	}
	if err := e.setProjectName(projectName); err != nil {
		return nil, err
	}
	if err := e.setSoftwareName(softwareName); err != nil {
		return nil, err
	}
	if err := e.setSoftwareVersion(softwareVersion); err != nil {
		return nil, err
	}

	e.jobURL = jobURL
	e.labels = normaliseLabels(labels)
	e.status = ExecutionStatusRunning
	e.startTime = startTime.UTC()
	if plannedCasesCount < 0 {
		return nil, errors.Errorf(errors.KindValidation, "计划用例数不能为负数: %d", plannedCasesCount)
	}
	e.plannedCasesCount = plannedCasesCount
	return e, nil
}

// RestoreExecution 从数据库重建实体。仅供持久化适配器使用。
func RestoreExecution(s ExecutionSnapshot) *TestExecution {
	status := ExecutionStatus(s.Status)
	if !status.Valid() {
		status = ExecutionStatusRunning
	}
	return &TestExecution{
		id:                s.ID,
		buildUID:          s.BuildUID,
		jobName:           s.JobName,
		jobURL:            s.JobURL,
		projectName:       s.ProjectName,
		softwareName:      s.SoftwareName,
		softwareVersion:   s.SoftwareVersion,
		labels:            s.Labels,
		status:            status,
		startTime:         s.StartTime,
		endTime:           s.EndTime,
		duration:          s.Duration,
		plannedCasesCount: s.PlannedCasesCount,
		passCount:         s.PassCount,
		failureCount:      s.FailureCount,
		skippedCount:      s.SkippedCount,
		passRate:          s.PassRate,
		lastHeartbeatAt:   s.LastHeartbeatAt,
		createdAt:         s.CreatedAt,
		updatedAt:         s.UpdatedAt,
	}
}

// ---- 读取器 ----

func (e *TestExecution) ID() int64                   { return e.id }
func (e *TestExecution) BuildUID() uuid.UUID         { return e.buildUID }
func (e *TestExecution) JobName() string             { return e.jobName }
func (e *TestExecution) JobURL() string              { return e.jobURL }
func (e *TestExecution) ProjectName() string         { return e.projectName }
func (e *TestExecution) SoftwareName() string        { return e.softwareName }
func (e *TestExecution) SoftwareVersion() string     { return e.softwareVersion }
func (e *TestExecution) Labels() []string            { return e.labels }
func (e *TestExecution) Status() ExecutionStatus     { return e.status }
func (e *TestExecution) StartTime() time.Time        { return e.startTime }
func (e *TestExecution) EndTime() *time.Time         { return e.endTime }
func (e *TestExecution) Duration() *float64          { return e.duration }
func (e *TestExecution) PlannedCasesCount() int      { return e.plannedCasesCount }
func (e *TestExecution) PassCount() int              { return e.passCount }
func (e *TestExecution) FailureCount() int           { return e.failureCount }
func (e *TestExecution) SkippedCount() int           { return e.skippedCount }
func (e *TestExecution) PassRate() *float64          { return e.passRate }
func (e *TestExecution) LastHeartbeatAt() *time.Time { return e.lastHeartbeatAt }
func (e *TestExecution) CreatedAt() time.Time        { return e.createdAt }
func (e *TestExecution) UpdatedAt() time.Time        { return e.updatedAt }

// TouchHeartbeat 记录进程仍存活。终止态调用无效。
func (e *TestExecution) TouchHeartbeat(at time.Time) bool {
	if e.status.IsTerminal() {
		return false
	}
	t := at.UTC()
	e.lastHeartbeatAt = &t
	return true
}

// ---- 行为 ----

// Finish 结束一次执行。
//
// 幂等规则：已进入终止态的 execution 再次收到 running 状态时**静默忽略**，
// 不报错也不回退。插件在异常路径下可能重复上报 running，报错会让插件失败。
func (e *TestExecution) Finish(status ExecutionStatus, endTime time.Time) (bool, error) {
	if !status.Valid() {
		return false, errors.Errorf(errors.KindValidation, "非法的 execution 状态: %q", status)
	}

	// 幂等：已处于终止态的 execution 不再接受任何结束操作。
	// 这覆盖两种重复上报：
	//   - 已终止又收到 running（回退）
	//   - 已 terminated 又收到 completed/failed（重复结束）
	// 插件在异常路径下可能重复上报，一律静默忽略，以**首次结束**为准，
	// 避免后到的数据覆盖先到且已计算的 duration。
	if e.status.IsTerminal() {
		return false, nil
	}
	if status == ExecutionStatusRunning {
		return false, nil
	}

	endTime = endTime.UTC()
	if endTime.Before(e.startTime) {
		return false, ErrEndTimeBeforeStart
	}

	e.status = status
	e.endTime = &endTime
	d := endTime.Sub(e.startTime).Seconds()
	e.duration = &d
	return true, nil
}

// RefreshCounts 按聚合结果回填计数与通过率（评审确认项 Q2）。
//
// 只在 execution 进入终止态时调用一次：运行中的计数没有意义
// （用例还在陆续上报），实时进度应由 Dashboard 快照提供。
//
// 通过率口径：成功数 / 已结束的用例数；没有已结束用例时为 0 而不是 NULL，
// 避免前端显示逻辑分叉。
func (e *TestExecution) RefreshCounts(pass, failure, skipped int) {
	e.passCount = pass
	e.failureCount = failure
	e.skippedCount = skipped

	finished := pass + failure + skipped
	rate := 0.0
	if finished > 0 {
		rate = float64(pass) / float64(finished)
	}
	e.passRate = &rate
}

// AddLabels 追加标签并去重。上报结束接口允许补充 labels。
func (e *TestExecution) AddLabels(labels []string) {
	e.labels = mergeLabels(e.labels, normaliseLabels(labels))
}

// ---- 内部校验 ----

func (e *TestExecution) setJobName(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrJobNameRequired
	}
	if len(v) > 255 {
		return errors.Errorf(errors.KindValidation, "任务名称长度不能超过 255 个字符")
	}
	e.jobName = v
	return nil
}

func (e *TestExecution) setProjectName(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrProjectNameRequired
	}
	if len(v) > 255 {
		return errors.Errorf(errors.KindValidation, "项目名称长度不能超过 255 个字符")
	}
	e.projectName = v
	return nil
}

func (e *TestExecution) setSoftwareName(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrSoftwareNameRequired
	}
	if len(v) > 100 {
		return errors.Errorf(errors.KindValidation, "被测软件名称长度不能超过 100 个字符")
	}
	e.softwareName = v
	return nil
}

func (e *TestExecution) setSoftwareVersion(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrSoftwareVersionRequired
	}
	if len(v) > 100 {
		return errors.Errorf(errors.KindValidation, "被测软件版本长度不能超过 100 个字符")
	}
	e.softwareVersion = v
	return nil
}

// normaliseLabels 清洗标签：去空白、去空串、截断到 64 字符（数据库列上限）。
func normaliseLabels(labels []string) []string {
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(labels))
	for _, l := range labels {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if len(l) > 64 {
			l = l[:64]
		}
		out = append(out, l)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// mergeLabels 合并两个标签集合并去重，保持原有顺序。
func mergeLabels(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, l := range append(append([]string{}, base...), extra...) {
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	return out
}
