package automation

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// MaxStepDepth 是步骤路径的最大段数，即最多 5 级 with step 嵌套。
// 路径 0.1.2.3.4 合法，0.1.2.3.4.5 拒绝。超过则视为滥用，调试期必须失败。
const MaxStepDepth = 5

// StepPath 是步骤在树中的位置，形如 `0` / `0.1` / `0.1.2`。
//
// 它是步骤上报的**幂等键**：与 case_uid 组成唯一约束
// `uq_execution_item_steps_path (case_uid, step_path)`，
// 重复上报同一路径会覆盖而不是插入。
//
// 用类型而不是裸 string，是为了让解析与校验只在一处发生。
type StepPath struct {
	segments []int
	raw      string
}

// ParseStepPath 解析步骤路径。
// 校验：非空、每段为整数、段数不超过 MaxStepDepth、各段非负。
func ParseStepPath(raw string) (StepPath, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return StepPath{}, ErrInvalidStepPath
	}

	parts := strings.Split(raw, ".")
	if len(parts) > MaxStepDepth {
		return StepPath{}, ErrStepPathTooDeep
	}

	segments := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return StepPath{}, ErrInvalidStepPath
		}
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 {
			return StepPath{}, ErrInvalidStepPath
		}
		segments = append(segments, v)
	}

	return StepPath{segments: segments, raw: raw}, nil
}

// String 返回规范化的路径字符串。
func (p StepPath) String() string {
	if p.raw != "" {
		return p.raw
	}
	return joinSegments(p.segments)
}

// Segments 返回路径各段。
func (p StepPath) Segments() []int { return p.segments }

// Depth 返回层级（根步骤为 0）。
func (p StepPath) Depth() int {
	if len(p.segments) == 0 {
		return 0
	}
	return len(p.segments) - 1
}

// LastSegment 返回最后一段，即 step_index。
func (p StepPath) LastSegment() int {
	if len(p.segments) == 0 {
		return 0
	}
	return p.segments[len(p.segments)-1]
}

// Parent 返回父路径。根步骤返回空路径与 false。
func (p StepPath) Parent() (StepPath, bool) {
	if len(p.segments) <= 1 {
		return StepPath{}, false
	}
	parent := StepPath{segments: p.segments[:len(p.segments)-1]}
	parent.raw = joinSegments(parent.segments)
	return parent, true
}

func joinSegments(segments []int) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, 0, len(segments))
	for _, s := range segments {
		parts = append(parts, strconv.Itoa(s))
	}
	return strings.Join(parts, ".")
}

// ExecutionStep 是一个测试步骤。
type ExecutionStep struct {
	id              int64
	caseUID         uuid.UUID
	parentStepID    *int64
	parentStepIndex *int
	stepIndex       int
	stepPath        StepPath
	depth           int
	stepName        string
	status          StepStatus
	startTime       *time.Time
	endTime         *time.Time
	duration        *float64
	createdAt       time.Time
	updatedAt       time.Time

	// attachments 是查询时的读模型扩展，不属于持久化字段。
	attachments []*Attachment
}

// StepSnapshot 用于从数据库重建步骤。
type StepSnapshot struct {
	ID              int64
	CaseUID         uuid.UUID
	ParentStepID    *int64
	ParentStepIndex *int
	StepIndex       int
	StepPath        string
	Depth           int
	StepName        string
	Status          string
	StartTime       *time.Time
	EndTime         *time.Time
	Duration        *float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewStepCommand 是创建步骤所需的输入。
type NewStepCommand struct {
	CaseUID   uuid.UUID
	StepPath  StepPath
	StepName  string
	Status    StepStatus
	StartTime *time.Time
	EndTime   *time.Time
	Duration  *float64
}

// NewStep 创建一个步骤。
// depth 由 step_path 推导，不由调用方传入，避免两者不一致。
func NewStep(cmd NewStepCommand) (*ExecutionStep, error) {
	s := &ExecutionStep{}

	if cmd.CaseUID == uuid.Nil {
		return nil, ErrInvalidCaseUID
	}
	s.caseUID = cmd.CaseUID

	if len(cmd.StepPath.segments) == 0 {
		return nil, ErrInvalidStepPath
	}
	s.stepPath = cmd.StepPath
	s.stepIndex = cmd.StepPath.LastSegment()
	s.depth = cmd.StepPath.Depth()

	name := strings.TrimSpace(cmd.StepName)
	if name == "" {
		return nil, ErrStepNameRequired
	}
	s.stepName = name

	status := StepStatus(strings.TrimSpace(string(cmd.Status)))
	if !status.Valid() {
		return nil, ErrInvalidStatus("步骤", string(cmd.Status))
	}
	s.status = status

	s.startTime = cmd.StartTime
	if cmd.EndTime != nil {
		t := cmd.EndTime.UTC()
		s.endTime = &t
	}
	s.duration = cmd.Duration

	// 未显式给出 duration 但两端时间都齐全时自动计算
	if s.duration == nil && s.startTime != nil && s.endTime != nil {
		d := s.endTime.Sub(*s.startTime).Seconds()
		s.duration = &d
	}
	return s, nil
}

// RestoreStep 从数据库重建实体。步骤路径解析失败时退化为单段 0，
// 保证脏数据不会让整棵步骤树加载失败。
func RestoreStep(s StepSnapshot) *ExecutionStep {
	path, err := ParseStepPath(s.StepPath)
	if err != nil {
		path = StepPath{segments: []int{0}, raw: s.StepPath}
	}
	status := StepStatus(s.Status)
	if !status.Valid() {
		status = CaseStatusRunning
	}
	return &ExecutionStep{
		id:              s.ID,
		caseUID:         s.CaseUID,
		parentStepID:    s.ParentStepID,
		parentStepIndex: s.ParentStepIndex,
		stepIndex:       s.StepIndex,
		stepPath:        path,
		depth:           path.Depth(),
		stepName:        s.StepName,
		status:          status,
		startTime:       s.StartTime,
		endTime:         s.EndTime,
		duration:        s.Duration,
		createdAt:       s.CreatedAt,
		updatedAt:       s.UpdatedAt,
	}
}

// ---- 读取器 ----

func (s *ExecutionStep) ID() int64             { return s.id }
func (s *ExecutionStep) CaseUID() uuid.UUID    { return s.caseUID }
func (s *ExecutionStep) ParentStepID() *int64  { return s.parentStepID }
func (s *ExecutionStep) ParentStepIndex() *int { return s.parentStepIndex }
func (s *ExecutionStep) StepIndex() int        { return s.stepIndex }
func (s *ExecutionStep) StepPath() StepPath    { return s.stepPath }
func (s *ExecutionStep) Depth() int            { return s.depth }
func (s *ExecutionStep) StepName() string      { return s.stepName }
func (s *ExecutionStep) Status() StepStatus    { return s.status }
func (s *ExecutionStep) StartTime() *time.Time { return s.startTime }
func (s *ExecutionStep) EndTime() *time.Time   { return s.endTime }
func (s *ExecutionStep) Duration() *float64    { return s.duration }
func (s *ExecutionStep) CreatedAt() time.Time  { return s.createdAt }
func (s *ExecutionStep) UpdatedAt() time.Time  { return s.updatedAt }

// Attachments 返回查询时填充的附件列表（读模型扩展）。
func (s *ExecutionStep) Attachments() []*Attachment { return s.attachments }

// SetAttachments 填充附件列表（查询用例层调用）。
func (s *ExecutionStep) SetAttachments(atts []*Attachment) { s.attachments = atts }

// ---- 行为 ----

// ApplyUpdate 用新上报的数据覆盖可变字段（幂等 upsert 用）。
// step_path 与 case_uid 是身份的一部分，不允许变更。
func (s *ExecutionStep) ApplyUpdate(status StepStatus, name string, startTime, endTime *time.Time, duration *float64) error {
	status = StepStatus(strings.TrimSpace(string(status)))
	if !status.Valid() {
		return ErrInvalidStatus("步骤", string(status))
	}
	s.status = status

	if name = strings.TrimSpace(name); name != "" {
		s.stepName = name
	}
	if startTime != nil {
		t := startTime.UTC()
		s.startTime = &t
	}
	if endTime != nil {
		t := endTime.UTC()
		s.endTime = &t
	}
	s.duration = duration
	if s.duration == nil && s.startTime != nil && s.endTime != nil {
		d := s.endTime.Sub(*s.startTime).Seconds()
		s.duration = &d
	}
	return nil
}

// SetParent 挂到父步骤下。parentID 为 nil 表示根节点。
func (s *ExecutionStep) SetParent(parentID *int64, parentIndex *int) {
	s.parentStepID = parentID
	s.parentStepIndex = parentIndex
}

// StepNode 是步骤树的节点，供查询接口返回嵌套结构。
type StepNode struct {
	Step     *ExecutionStep
	Children []StepNode
}

// BuildStepTree 把平铺的步骤列表组装成树。
//
// 与权限树不同，步骤树**允许丢弃**找不到父节点的孤儿：
// 插件可能乱序上报（先报子步骤后补父步骤），若把孤儿提为根节点，
// 同一个步骤会在树上出现两次（一次作为孤儿根、一次作为真正父节点的子节点）。
// 因此这里丢弃孤儿，并在返回中单独给出未挂载的步骤数量，由调用方决定是否告警。
//
// 实现要点：必须先用**指针**结构挂树、最后再一次性转值切片。
// 若直接对值切片「先 append 进 roots、后给子节点追加 Children」，
// append 进 roots 的是那一刻的副本，后续挂载对副本不可见，会导致
// 所有根节点的 Children 恒为空（与权限树的同类缺陷一致）。
func BuildStepTree(steps []*ExecutionStep) (roots []StepNode, orphans int) {
	byPath := make(map[string]*stepBuildNode, len(steps))
	nodes := make([]*stepBuildNode, 0, len(steps))

	for _, s := range steps {
		node := &stepBuildNode{step: s}
		nodes = append(nodes, node)
		byPath[s.StepPath().String()] = node
	}

	for _, node := range nodes {
		parentPath, ok := node.step.StepPath().Parent()
		if !ok {
			continue // 根节点：isChild 保持 false，天然是根
		}
		parent, ok := byPath[parentPath.String()]
		if !ok {
			// 孤儿：父步骤不存在。标记 isChild 让它在最后的根收集循环里被跳过，
			// 否则会因 isChild=false 被误当成根节点输出。
			orphans++
			node.isChild = true
			continue
		}
		parent.children = append(parent.children, node)
		node.isChild = true
	}

	for _, node := range nodes {
		if !node.isChild {
			roots = append(roots, node.toNode())
		}
	}
	return roots, orphans
}

// stepBuildNode 是建树过程中的中间结构，用指针保证子树能正确挂载。
type stepBuildNode struct {
	step     *ExecutionStep
	children []*stepBuildNode
	isChild  bool
}

func (n *stepBuildNode) toNode() StepNode {
	out := StepNode{
		Step:     n.step,
		Children: make([]StepNode, 0, len(n.children)),
	}
	for _, c := range n.children {
		out.Children = append(out.Children, c.toNode())
	}
	return out
}

// EnsureDepthIsConsistent 校验 depth 与 step_path 推导结果一致。
// 数据库里这两列冗余，写入前校验一次可避免不一致数据入库。
func (s *ExecutionStep) EnsureDepthIsConsistent() error {
	if s.depth != s.stepPath.Depth() {
		return errors.Errorf(errors.KindInternal,
			"步骤 depth(%d) 与 step_path(%s) 推导出的层级(%d) 不一致",
			s.depth, s.stepPath, s.stepPath.Depth())
	}
	return nil
}

// String 便于日志与错误信息中展示步骤。
func (s *ExecutionStep) String() string {
	return fmt.Sprintf("step(%s %s)", s.stepPath, s.stepName)
}
