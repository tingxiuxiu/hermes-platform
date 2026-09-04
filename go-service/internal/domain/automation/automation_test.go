package automation_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

var testBuildUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// ---------------------------------------------------------------------------
// Execution 状态机
// ---------------------------------------------------------------------------

func TestExecutionStatusValidation(t *testing.T) {
	valid := []string{"running", "completed", "failed", "aborted", "calculated"}
	for _, v := range valid {
		if _, err := automation.ParseExecutionStatus(v); err != nil {
			t.Errorf("status %q must be valid: %v", v, err)
		}
	}
	invalid := []string{"", "Running", "SUCCESS", "passed", "1"}
	for _, v := range invalid {
		if _, err := automation.ParseExecutionStatus(v); err == nil {
			t.Errorf("status %q must be rejected", v)
		}
	}
}

func TestExecutionStatusIsTerminal(t *testing.T) {
	if automation.ExecutionStatusRunning.IsTerminal() {
		t.Error("running must not be terminal")
	}
	for _, s := range []automation.ExecutionStatus{
		automation.ExecutionStatusCompleted,
		automation.ExecutionStatusFailed,
		automation.ExecutionStatusAborted,
		automation.ExecutionStatusCalculated,
	} {
		if !s.IsTerminal() {
			t.Errorf("%s must be terminal", s)
		}
	}
}

func TestExecutionTouchHeartbeatIgnoredWhenTerminal(t *testing.T) {
	start := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	e, err := automation.NewExecution(
		testBuildUID, "job-a", "http://jenkins/job-a",
		"proj-x", "app", "1.0.0", nil, 1, start,
	)
	if err != nil {
		t.Fatalf("new execution: %v", err)
	}
	if !e.TouchHeartbeat(start.Add(time.Second)) {
		t.Fatal("running execution must accept heartbeat")
	}
	if e.LastHeartbeatAt() == nil {
		t.Fatal("last_heartbeat_at must be set")
	}

	if _, err := e.Finish(automation.ExecutionStatusCompleted, start.Add(time.Minute)); err != nil {
		t.Fatalf("finish: %v", err)
	}
	if e.TouchHeartbeat(start.Add(2 * time.Minute)) {
		t.Fatal("terminal execution must ignore heartbeat")
	}
}

func TestNewExecutionDefaultsToRunning(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	e, err := automation.NewExecution(
		testBuildUID, "job-a", "http://jenkins/job-a",
		"proj-x", "app", "1.0.0", []string{"smoke"}, 10, start,
	)
	if err != nil {
		t.Fatalf("new execution: %v", err)
	}
	if e.Status() != automation.ExecutionStatusRunning {
		t.Errorf("status = %s, want running", e.Status())
	}
	if e.PlannedCasesCount() != 10 {
		t.Errorf("planned = %d, want 10", e.PlannedCasesCount())
	}
	if !e.StartTime().Equal(start) {
		t.Errorf("start = %v, want %v", e.StartTime(), start)
	}
}

func TestNewExecutionValidations(t *testing.T) {
	start := time.Now()

	tests := []struct {
		name    string
		mutate  func()
		wantErr error
	}{
		{"nil build_uid", func() {}, automation.ErrInvalidBuildUID},
		{"empty job_name", func() {}, automation.ErrJobNameRequired},
		{"empty project", func() {}, automation.ErrProjectNameRequired},
		{"empty software_name", func() {}, automation.ErrSoftwareNameRequired},
		{"empty version", func() {}, automation.ErrSoftwareVersionRequired},
		{"negative planned", func() {}, nil},
	}

	// 由于字段校验分散，逐个构造
	if _, err := automation.NewExecution(uuid.Nil, "job", "", "p", "s", "v", nil, 0, start); !errors.Is(err, errors.KindValidation) {
		t.Errorf("nil build_uid: want validation error, got %v", err)
	}
	if _, err := automation.NewExecution(testBuildUID, "", "", "p", "s", "v", nil, 0, start); !errors.Is(err, errors.KindValidation) {
		t.Errorf("empty job_name: want validation error, got %v", err)
	}
	if _, err := automation.NewExecution(testBuildUID, "job", "", "", "s", "v", nil, 0, start); !errors.Is(err, errors.KindValidation) {
		t.Errorf("empty project: want validation error, got %v", err)
	}
	if _, err := automation.NewExecution(testBuildUID, "job", "", "p", "", "v", nil, 0, start); !errors.Is(err, errors.KindValidation) {
		t.Errorf("empty software_name: want validation error, got %v", err)
	}
	if _, err := automation.NewExecution(testBuildUID, "job", "", "p", "s", "", nil, 0, start); !errors.Is(err, errors.KindValidation) {
		t.Errorf("empty version: want validation error, got %v", err)
	}
	if _, err := automation.NewExecution(testBuildUID, "job", "", "p", "s", "v", nil, -1, start); err == nil {
		t.Error("negative planned must be rejected")
	}
	_ = tests
}

func TestExecutionFinishComputesDuration(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 10, 5, 0, 0, time.UTC)

	e, _ := automation.NewExecution(testBuildUID, "job", "", "p", "s", "v", nil, 0, start)
	changed, err := e.Finish(automation.ExecutionStatusCompleted, end)
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	if !changed {
		t.Error("first finish must report changed")
	}
	if e.Status() != automation.ExecutionStatusCompleted {
		t.Errorf("status = %s", e.Status())
	}
	if e.EndTime() == nil || !e.EndTime().Equal(end) {
		t.Errorf("end = %v, want %v", e.EndTime(), end)
	}
	if e.Duration() == nil || *e.Duration() != 300 {
		t.Errorf("duration = %v, want 300", e.Duration())
	}
}

func TestExecutionCannotFinishBeforeStart(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 5, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)

	e, _ := automation.NewExecution(testBuildUID, "job", "", "p", "s", "v", nil, 0, start)
	if _, err := e.Finish(automation.ExecutionStatusCompleted, end); err == nil {
		t.Error("finishing before start must fail")
	}
}

// TestExecutionTerminalIsIdempotent 是幂等回归：
// 已进入终止态的 execution 再次收到 running 时静默忽略，不报错也不回退。
func TestExecutionTerminalIsIdempotent(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 10, 1, 0, 0, time.UTC)

	e, _ := automation.NewExecution(testBuildUID, "job", "", "p", "s", "v", nil, 0, start)
	if _, err := e.Finish(automation.ExecutionStatusCompleted, end); err != nil {
		t.Fatalf("first finish: %v", err)
	}

	// 再次收到 running：必须静默忽略，保持 completed
	changed, err := e.Finish(automation.ExecutionStatusRunning, time.Now())
	if err != nil {
		t.Fatalf("idempotent running must not error, got %v", err)
	}
	if changed {
		t.Error("idempotent running must report not-changed")
	}
	if e.Status() != automation.ExecutionStatusCompleted {
		t.Errorf("status regressed to %s", e.Status())
	}

	// 重复结束同一状态：也应幂等（不重复计算）
	changed, err = e.Finish(automation.ExecutionStatusCompleted, end.Add(time.Hour))
	if err != nil {
		t.Fatalf("duplicate finish: %v", err)
	}
	if changed {
		t.Error("duplicate finish must report not-changed")
	}
}

// TestExecutionRefreshCounts 是评审确认项 Q2 的回归：
// 通过率 = 成功 / 已结束用例；无已结束用例时为 0（非 NULL）。
func TestExecutionRefreshCounts(t *testing.T) {
	e, _ := automation.NewExecution(testBuildUID, "job", "", "p", "s", "v", nil, 0, time.Now())

	e.RefreshCounts(8, 1, 1)
	if e.PassCount() != 8 || e.FailureCount() != 1 || e.SkippedCount() != 1 {
		t.Errorf("counts = %d/%d/%d, want 8/1/1", e.PassCount(), e.FailureCount(), e.SkippedCount())
	}
	if e.PassRate() == nil || *e.PassRate() != 0.8 {
		t.Errorf("pass_rate = %v, want 0.8", e.PassRate())
	}

	// 无已结束用例：pass_rate 为 0
	e.RefreshCounts(0, 0, 0)
	if e.PassRate() == nil || *e.PassRate() != 0 {
		t.Errorf("pass_rate with no finished cases must be 0, got %v", e.PassRate())
	}
}

// ---------------------------------------------------------------------------
// StepPath
// ---------------------------------------------------------------------------

func TestParseStepPathValid(t *testing.T) {
	cases := []struct {
		raw    string
		depth  int
		index  int
		parent string
	}{
		{"0", 0, 0, ""},
		{"0.1", 1, 1, "0"},
		{"0.1.2", 2, 2, "0.1"},
		{"0.1.2.3.4", 4, 4, "0.1.2.3"},
		{"12", 0, 12, ""},
		{"3.0", 1, 0, "3"},
	}

	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			p, err := automation.ParseStepPath(c.raw)
			if err != nil {
				t.Fatalf("parse %q: %v", c.raw, err)
			}
			if p.Depth() != c.depth {
				t.Errorf("depth = %d, want %d", p.Depth(), c.depth)
			}
			if p.LastSegment() != c.index {
				t.Errorf("last segment = %d, want %d", p.LastSegment(), c.index)
			}
			parent, ok := p.Parent()
			if c.parent == "" {
				if ok {
					t.Errorf("%q must have no parent", c.raw)
				}
			} else {
				if !ok || parent.String() != c.parent {
					t.Errorf("parent = %q, want %q", parent.String(), c.parent)
				}
			}
		})
	}
}

func TestParseStepPathInvalid(t *testing.T) {
	invalid := []string{"", " ", "a", "0.a", "1.2.", ".1", "1..2", "-1", "0.1.2.3.4.5"}
	for _, raw := range invalid {
		if _, err := automation.ParseStepPath(raw); err == nil {
			t.Errorf("step path %q must be rejected", raw)
		}
	}
}

func TestParseStepPathRejectsSixthLevel(t *testing.T) {
	_, err := automation.ParseStepPath("0.1.2.3.4.5")
	if err == nil {
		t.Fatal("sixth-level step path must be rejected")
	}
	if !errors.Is(err, errors.KindValidation) {
		t.Fatalf("error = %v, want validation", err)
	}
	if got := errors.Message(err); got != "步骤层级不能超过 5 层" {
		t.Fatalf("message = %q, want 5-level limit", got)
	}
}

func TestParseStepPathNormalises(t *testing.T) {
	p, err := automation.ParseStepPath(" 0.1 ")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.String() != "0.1" {
		t.Errorf("normalised path = %q, want 0.1", p.String())
	}
}

// ---------------------------------------------------------------------------
// ExecutionItem
// ---------------------------------------------------------------------------

func TestNewItemDefaults(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	item, err := automation.NewItem(automation.NewItemCommand{
		BuildUID:      testBuildUID,
		CaseUID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		CaseKey:       "JIRA-100",
		CaseName:      "登录测试",
		AttemptNumber: 1,
		StartTime:     &start,
	})
	if err != nil {
		t.Fatalf("new item: %v", err)
	}
	if item.Status() != automation.CaseStatusRunning {
		t.Errorf("status = %s, want running", item.Status())
	}
	if !item.IsLatest() {
		t.Error("new item must be latest")
	}
	if item.AttemptNumber() != 1 {
		t.Errorf("attempt = %d, want 1", item.AttemptNumber())
	}
}

func TestNewItemValidations(t *testing.T) {
	base := automation.NewItemCommand{
		BuildUID:      testBuildUID,
		CaseUID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		CaseKey:       "JIRA-100",
		CaseName:      "登录测试",
		AttemptNumber: 1,
	}

	// 空 case_key
	bad := base
	bad.CaseKey = ""
	if _, err := automation.NewItem(bad); err == nil {
		t.Error("empty case_key must be rejected")
	}
	// 空 case_name
	bad = base
	bad.CaseName = ""
	if _, err := automation.NewItem(bad); err == nil {
		t.Error("empty case_name must be rejected")
	}
	// attempt 0
	bad = base
	bad.AttemptNumber = 0
	if _, err := automation.NewItem(bad); err == nil {
		t.Error("attempt 0 must be rejected")
	}
	// 非法 case_uid
	bad = base
	bad.CaseUID = uuid.Nil
	if _, err := automation.NewItem(bad); err == nil {
		t.Error("nil case_uid must be rejected")
	}
}

func TestItemFinishIsTerminalAndIdempotent(t *testing.T) {
	start := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 10, 1, 0, 0, time.UTC)

	item, _ := automation.NewItem(automation.NewItemCommand{
		BuildUID:      testBuildUID,
		CaseUID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		CaseKey:       "JIRA-100",
		CaseName:      "登录测试",
		AttemptNumber: 1,
		StartTime:     &start,
	})

	changed, err := item.Finish(automation.CaseStatusFailed, &end, "断言失败", "Traceback...")
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	if !changed {
		t.Error("first finish must report changed")
	}
	if item.Status() != automation.CaseStatusFailed {
		t.Errorf("status = %s", item.Status())
	}
	if item.ErrorMessage() != "断言失败" {
		t.Errorf("error message = %q", item.ErrorMessage())
	}
	if item.Duration() == nil || *item.Duration() != 60 {
		t.Errorf("duration = %v, want 60", item.Duration())
	}

	// 幂等：running 回退忽略
	changed, err = item.Finish(automation.CaseStatusRunning, nil, "", "")
	if err != nil {
		t.Fatalf("idempotent running: %v", err)
	}
	if changed {
		t.Error("idempotent running must report not-changed")
	}
	if item.Status() != automation.CaseStatusFailed {
		t.Errorf("status regressed to %s", item.Status())
	}
}

func TestCaseStatusValidation(t *testing.T) {
	for _, v := range []string{"running", "passed", "failed", "skipped", "broken"} {
		if _, err := automation.ParseCaseStatus(v); err != nil {
			t.Errorf("status %q must be valid: %v", v, err)
		}
	}
	for _, v := range []string{"", "PASSED", "error", "1"} {
		if _, err := automation.ParseCaseStatus(v); err == nil {
			t.Errorf("status %q must be rejected", v)
		}
	}
}

// ---------------------------------------------------------------------------
// 步骤树
// ---------------------------------------------------------------------------

func TestBuildStepTreeNestsAndFlagsOrphans(t *testing.T) {
	mk := func(path string) *automation.ExecutionStep {
		sp, _ := automation.ParseStepPath(path)
		s, err := automation.NewStep(automation.NewStepCommand{
			CaseUID:  testBuildUID,
			StepPath: sp,
			StepName: "步骤 " + path,
			Status:   automation.CaseStatusPassed,
		})
		if err != nil {
			t.Fatalf("new step %s: %v", path, err)
		}
		return s
	}

	// 根步骤 0 和 3；子步骤 0.1 / 0.1.2；
	// 真孤儿 1.2（父步骤 1 不在输入中）
	steps := []*automation.ExecutionStep{
		mk("0"), mk("0.1"), mk("0.1.2"),
		mk("3"),
		mk("1.2"),
	}

	roots, orphans := automation.BuildStepTree(steps)

	if orphans != 1 {
		t.Errorf("orphans = %d, want 1 (1.2 has no parent 1)", orphans)
	}
	if len(roots) != 2 {
		t.Fatalf("roots = %d, want 2 (0 and 3; orphan 1.2 must be dropped)", len(roots))
	}
	// 孤儿必须被丢弃，不能提为根节点
	for _, r := range roots {
		if r.Step.StepPath().String() == "1.2" {
			t.Error("orphan 1.2 must be dropped, not promoted to root")
		}
	}

	// 校验 0 → 0.1 → 0.1.2 嵌套
	root := roots[0]
	if root.Step.StepPath().String() != "0" {
		t.Fatalf("root = %s, want 0", root.Step.StepPath())
	}
	if len(root.Children) != 1 || len(root.Children[0].Children) != 1 {
		t.Fatalf("0.1.2 nested incorrectly: got %d children at 0, %d at 0.1",
			len(root.Children), len(root.Children[0].Children))
	}
	leaf := root.Children[0].Children[0]
	if leaf.Step.StepPath().String() != "0.1.2" {
		t.Errorf("leaf = %s, want 0.1.2", leaf.Step.StepPath())
	}
}

// ---------------------------------------------------------------------------
// Attachment
// ---------------------------------------------------------------------------

func TestAttachmentRequiresExactlyOneOwner(t *testing.T) {
	// 两个都有 → 拒
	_, err := automation.NewAttachment(automation.NewAttachmentCommand{
		ItemID: intPtr(1), StepID: intPtr(2),
		Type: automation.AttachmentTypeScreenshot, URL: "http://x/s.png",
	})
	if err == nil {
		t.Error("attachment with both item_id and step_id must be rejected")
	}

	// 两个都没有 → 拒
	_, err = automation.NewAttachment(automation.NewAttachmentCommand{
		Type: automation.AttachmentTypeLog, URL: "http://x/l.log",
	})
	if err == nil {
		t.Error("attachment with no owner must be rejected")
	}

	// 只有一个 → 通过
	a, err := automation.NewAttachment(automation.NewAttachmentCommand{
		ItemID: intPtr(1),
		Type:   automation.AttachmentTypeScreenshot, URL: "http://x/s.png",
	})
	if err != nil {
		t.Fatalf("item-scoped attachment must be valid: %v", err)
	}
	if a.ItemID() == nil || *a.ItemID() != 1 || a.StepID() != nil {
		t.Errorf("item attachment ownership wrong: item=%v step=%v", a.ItemID(), a.StepID())
	}
}

func TestAttachmentTypeAndURLValidation(t *testing.T) {
	_, err := automation.NewAttachment(automation.NewAttachmentCommand{
		ItemID: intPtr(1), Type: automation.AttachmentType("pdf"), URL: "http://x/f.pdf",
	})
	if err == nil {
		t.Error("invalid attachment type must be rejected")
	}

	_, err = automation.NewAttachment(automation.NewAttachmentCommand{
		ItemID: intPtr(1), Type: automation.AttachmentTypeOther, URL: "",
	})
	if err == nil {
		t.Error("empty url must be rejected")
	}
}

func intPtr(v int64) *int64 { return &v }
