package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/test/harness"
)

// setupDashboardData 上报一个 running execution + 若干用例，并返回 admin 令牌。
// 用于验证 dashboard 概览能正确投影。
func setupDashboardData(t *testing.T, env *harness.Env) (string, int64, string) {
	t.Helper()

	buildUID := uuid.New().String()
	exec := createExecution(t, env, buildUID)

	// 2 个用例：1 结束(passed) + 1 running
	u1, u2 := uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, "JIRA-DASH-1", u1)
	createItem(t, env, buildUID, "JIRA-DASH-2", u2)
	updateItem(t, env, buildUID, u1, "passed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	return admin.AccessToken, exec.ID, buildUID
}

func TestDashboardOverview(t *testing.T) {
	env := harness.Setup(t)
	token, _, buildUID := setupDashboardData(t, env)

	w := doAutomation(t, env, http.MethodGet, api(env)+"/automation/dashboard/overview", "",
		bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard overview: status=%d body=%s", w.Code, w.Body.String())
	}

	var data struct {
		Summary struct {
			TotalExecutionsCount   int `json:"total_executions_count"`
			RunningExecutionsCount int `json:"running_executions_count"`
			SuccessCases7d         int `json:"success_cases_7d"`
			FailureCases7d         int `json:"failure_cases_7d"`
		} `json:"summary"`
		RunningExecutions []struct {
			ExecutionID      int64    `json:"execution_id"`
			BuildUID         string   `json:"build_uid"`
			Status           string   `json:"status"`
			ProgressPercent  float64  `json:"progress_percent"`
			RunningCaseNames []string `json:"running_case_names"`
		} `json:"running_executions"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode overview: %v", err)
	}

	// execution 未结束（running），所以不进入 7 天终止态计数
	if data.Summary.RunningExecutionsCount != 1 {
		t.Errorf("running_executions = %d, want 1", data.Summary.RunningExecutionsCount)
	}
	if len(data.RunningExecutions) != 1 {
		t.Fatalf("running_executions list = %d, want 1", len(data.RunningExecutions))
	}
	re := data.RunningExecutions[0]
	if re.BuildUID != buildUID {
		t.Errorf("build_uid = %s, want %s", re.BuildUID, buildUID)
	}
	if re.Status != "running" {
		t.Errorf("status = %s, want running", re.Status)
	}
	if re.ProgressPercent != 50 { // completed=1(pre), planned=2 → 50%
		t.Errorf("progress = %v, want 50", re.ProgressPercent)
	}
	// running_case_names 应包含未结束的用例
	if len(re.RunningCaseNames) != 1 || re.RunningCaseNames[0] != "登录测试" {
		t.Errorf("running_case_names = %v, want [登录测试]", re.RunningCaseNames)
	}
}

func TestDashboardRunningCases(t *testing.T) {
	env := harness.Setup(t)
	token, execID, _ := setupDashboardData(t, env)

	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/dashboard/executions/"+itoa(execID)+"/cases", "",
		bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("running cases: status=%d body=%s", w.Code, w.Body.String())
	}

	var data struct {
		ExecutionID int64 `json:"execution_id"`
		Items       []struct {
			ItemID  int64  `json:"item_id"`
			CaseKey string `json:"case_key"`
			Status  string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode running cases: %v", err)
	}
	if data.ExecutionID != execID {
		t.Errorf("execution_id = %d, want %d", data.ExecutionID, execID)
	}
	if len(data.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(data.Items))
	}
	// 两个用例各自生效 attempt
	statuses := map[string]string{}
	for _, item := range data.Items {
		statuses[item.CaseKey] = item.Status
	}
	if statuses["JIRA-DASH-1"] != "passed" {
		t.Errorf("JIRA-DASH-1 status = %s, want passed", statuses["JIRA-DASH-1"])
	}
	if statuses["JIRA-DASH-2"] != "running" {
		t.Errorf("JIRA-DASH-2 status = %s, want running", statuses["JIRA-DASH-2"])
	}
}

func TestDashboardOverviewAfterExecutionFinished(t *testing.T) {
	env := harness.Setup(t)

	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1, u2 := uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, "JIRA-FINISH-1", u1)
	createItem(t, env, buildUID, "JIRA-FINISH-2", u2)
	updateItem(t, env, buildUID, u1, "passed")
	updateItem(t, env, buildUID, u2, "failed")
	finishExecution(t, env, buildUID, "completed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	w := doAutomation(t, env, http.MethodGet, api(env)+"/automation/dashboard/overview", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard overview: status=%d body=%s", w.Code, w.Body.String())
	}

	var data struct {
		Summary struct {
			TotalExecutionsCount int     `json:"total_executions_count"`
			SuccessCases7d       int     `json:"success_cases_7d"`
			FailureCases7d       int     `json:"failure_cases_7d"`
			PassRate7d           float64 `json:"pass_rate_7d"`
		} `json:"summary"`
		Trends []struct {
			StatDate       string `json:"stat_date"`
			ExecutionTotal int    `json:"execution_total"`
			SuccessCases   int    `json:"success_cases"`
			FailureCases   int    `json:"failure_cases"`
		} `json:"trends"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if data.Summary.TotalExecutionsCount != 1 {
		t.Errorf("total_executions = %d, want 1", data.Summary.TotalExecutionsCount)
	}
	if data.Summary.SuccessCases7d != 1 {
		t.Errorf("success = %d, want 1", data.Summary.SuccessCases7d)
	}
	if data.Summary.FailureCases7d != 1 {
		t.Errorf("failure = %d, want 1", data.Summary.FailureCases7d)
	}
	if data.Summary.PassRate7d != 0.5 {
		t.Errorf("pass_rate = %v, want 0.5", data.Summary.PassRate7d)
	}
	if len(data.Trends) == 0 {
		t.Fatal("trends empty, want daily buckets for the last 7 days")
	}
	last := data.Trends[len(data.Trends)-1]
	if last.ExecutionTotal != 1 || last.SuccessCases != 1 || last.FailureCases != 1 {
		t.Errorf("today trend = %+v, want exec=1 success=1 failure=1", last)
	}
}
