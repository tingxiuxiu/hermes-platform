package dashboard_test

import (
	"testing"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/dashboard"
)

var now = time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

// ---------------------------------------------------------------------------
// ProgressPercent
// ---------------------------------------------------------------------------

func TestProgressPercent(t *testing.T) {
	cases := []struct {
		completed, pre int
		want           float64
	}{
		{5, 10, 50},
		{0, 10, 0},
		{10, 10, 100},
		{3, 0, 0}, // 预计为 0 → 0
		{0, 0, 0},
	}
	for _, c := range cases {
		if got := dashboard.ProgressPercent(c.completed, c.pre); got != c.want {
			t.Errorf("progress(%d/%d) = %v, want %v", c.completed, c.pre, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// BuildExecutionSnapshot
// ---------------------------------------------------------------------------

func TestBuildExecutionSnapshotCountsAndNames(t *testing.T) {
	in := dashboard.ExecutionInput{
		ExecutionID:   1,
		JobID:         7,
		JobName:       "job-a",
		Status:        "running",
		StartedAt:     now,
		PreCasesCount: 6,
		Cases: []dashboard.CaseInput{
			{ItemID: 1, CaseName: "case1", Status: "passed"},
			{ItemID: 2, CaseName: "case2", Status: "failed"},
			{ItemID: 3, CaseName: "case3", Status: "broken"},
			{ItemID: 4, CaseName: "case4", Status: "skipped"},
			{ItemID: 5, CaseName: "case5", Status: "running"},
			{ItemID: 6, CaseName: "case6", Status: "running"},
			// 故意多一个 running，验证 running_case_names 只取前 5
			{ItemID: 7, CaseName: "case7", Status: "running"},
		},
	}

	s := dashboard.BuildExecutionSnapshot(in, now)

	if s.SuccessCasesCount != 1 {
		t.Errorf("success = %d, want 1", s.SuccessCasesCount)
	}
	if s.FailureCasesCount != 2 {
		t.Errorf("failure = %d, want 2 (failed + broken)", s.FailureCasesCount)
	}
	if s.SkippedCasesCount != 1 {
		t.Errorf("skipped = %d, want 1", s.SkippedCasesCount)
	}
	if s.RunningCasesCount != 3 {
		t.Errorf("running = %d, want 3", s.RunningCasesCount)
	}
	if s.CompletedCasesCount != 4 {
		t.Errorf("completed = %d, want 4", s.CompletedCasesCount)
	}
	// running_case_names 取前 5 个 running 的名字（这里只有 3 个）
	if len(s.RunningCaseNames) != 3 {
		t.Errorf("running_case_names = %v, want 3 entries", s.RunningCaseNames)
	}
}

// ---------------------------------------------------------------------------
// BuildCaseSnapshots（D-15：只投影最新 attempt）
// ---------------------------------------------------------------------------

func TestBuildCaseSnapshotsKeepsLatestAttempt(t *testing.T) {
	in := dashboard.ExecutionInput{
		ExecutionID: 1,
		Cases: []dashboard.CaseInput{
			{ItemID: 1, CaseKey: "JIRA-1", AttemptNumber: 1, Status: "failed"},
			{ItemID: 2, CaseKey: "JIRA-1", AttemptNumber: 2, Status: "passed"},
			{ItemID: 3, CaseKey: "JIRA-2", AttemptNumber: 1, Status: "skipped"},
		},
	}

	snapshots := dashboard.BuildCaseSnapshots(in, now)

	if len(snapshots) != 2 {
		t.Fatalf("snapshots = %d, want 2 (only latest attempt per case)", len(snapshots))
	}

	// JIRA-1 应取 attempt 2（passed），ItemID 为 2
	var jira1 *dashboard.CaseSnapshot
	for i := range snapshots {
		if snapshots[i].CaseKey == "JIRA-1" {
			jira1 = &snapshots[i]
		}
	}
	if jira1 == nil {
		t.Fatal("JIRA-1 missing from snapshots")
	}
	if jira1.AttemptNumber != 2 || jira1.Status != "passed" {
		t.Errorf("JIRA-1 = attempt %d status %s, want attempt 2 passed", jira1.AttemptNumber, jira1.Status)
	}
	if jira1.ItemID != 2 {
		t.Errorf("JIRA-1 item_id = %d, want 2 (D-15 latest attempt id)", jira1.ItemID)
	}
}

// ---------------------------------------------------------------------------
// CaseStatusCategory（D-16）
// ---------------------------------------------------------------------------

func TestCaseStatusCategory(t *testing.T) {
	cases := map[string]string{
		"passed":  "success",
		"failed":  "failure",
		"broken":  "failure",
		"skipped": "skipped",
		"running": "",
		"unknown": "",
	}
	for status, want := range cases {
		if got := dashboard.CaseStatusCategory(status); got != want {
			t.Errorf("category(%s) = %q, want %q", status, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// BuildSummarySnapshot
// ---------------------------------------------------------------------------

func TestBuildSummarySnapshot(t *testing.T) {
	in := dashboard.SummaryInput{
		TerminalExecutions: []dashboard.TerminalExecution{
			{ExecutionID: 1, Duration: ptr(100.0)},
			{ExecutionID: 2, Duration: ptr(200.0)},
			{ExecutionID: 3, Duration: nil},
		},
		TerminalCases: []dashboard.TerminalCase{
			{CaseKey: "a", Status: "passed"},
			{CaseKey: "b", Status: "failed"},
			{CaseKey: "c", Status: "broken"},
			{CaseKey: "d", Status: "skipped"},
		},
		ActiveJobsCount:        5,
		RunningExecutionsCount: 2,
		Now:                    now,
	}

	s := dashboard.BuildSummarySnapshot(in)

	if s.TotalExecutionsCount != 3 {
		t.Errorf("total_executions = %d, want 3", s.TotalExecutionsCount)
	}
	if s.ActiveJobsCount != 5 {
		t.Errorf("active_jobs = %d, want 5", s.ActiveJobsCount)
	}
	if s.RunningExecutionsCount != 2 {
		t.Errorf("running = %d, want 2", s.RunningExecutionsCount)
	}
	if s.SuccessCases7d != 1 {
		t.Errorf("success = %d, want 1", s.SuccessCases7d)
	}
	if s.FailureCases7d != 2 {
		t.Errorf("failure = %d, want 2", s.FailureCases7d)
	}
	if s.SkippedCases7d != 1 {
		t.Errorf("skipped = %d, want 1", s.SkippedCases7d)
	}
	// pass_rate = 1/(1+2+1) = 0.25
	if s.PassRate7d != 0.25 {
		t.Errorf("pass_rate = %v, want 0.25", s.PassRate7d)
	}
	// 平均耗时 = (100+200)/2 = 150（nil 不计入分母）
	if s.AvgExecutionDuration7d == nil || *s.AvgExecutionDuration7d != 150 {
		t.Errorf("avg_duration = %v, want 150", s.AvgExecutionDuration7d)
	}
}

func TestBuildSummarySnapshotNoFinishedCases(t *testing.T) {
	s := dashboard.BuildSummarySnapshot(dashboard.SummaryInput{Now: now})
	if s.PassRate7d != 0 {
		t.Errorf("pass_rate with no cases must be 0, got %v", s.PassRate7d)
	}
	if s.AvgExecutionDuration7d != nil {
		t.Errorf("avg_duration with no executions must be nil, got %v", *s.AvgExecutionDuration7d)
	}
}

// ---------------------------------------------------------------------------
// BuildTrendSnapshots（补齐空日期）
// ---------------------------------------------------------------------------

func TestBuildTrendSnapshotsFillsGaps(t *testing.T) {
	end := now // 2026-08-31

	cases := []dashboard.TerminalCase{
		{ExecutionID: 1, CaseKey: "a", Status: "passed", StartedAt: &now},
		{ExecutionID: 1, CaseKey: "b", Status: "failed", StartedAt: &now},
	}
	execs := []dashboard.TerminalExecution{
		{ExecutionID: 1, Status: "completed", StartedAt: now},
	}

	trends := dashboard.BuildTrendSnapshots(cases, execs, end, 7)

	if len(trends) != 7 {
		t.Fatalf("trends = %d, want 7 (gaps filled)", len(trends))
	}
	// 最后一天（今天）应有数据
	last := trends[6]
	if last.ExecutionTotal != 1 || last.SuccessCases != 1 || last.FailureCases != 1 {
		t.Errorf("last day = %+v, want exec=1 success=1 failure=1", last)
	}
	// 空日期应为 0
	first := trends[0]
	if first.ExecutionTotal != 0 || first.SuccessCases != 0 {
		t.Errorf("empty day = %+v, want all zero", first)
	}
}

func ptr(v float64) *float64 { return &v }
