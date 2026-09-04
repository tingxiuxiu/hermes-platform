package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/test/harness"
)

// TestE2EFullPytestSession 模拟一次完整的 pytest 会话（T-9.13）：
// 建 execution → 上报多 case（含重试）→ 上报 steps → 结束 execution →
// 通过查询接口与 Dashboard 概览验证数据完整链路。
func TestE2EFullPytestSession(t *testing.T) {
	env := harness.Setup(t)

	buildUID := uuid.New().String()

	// 1. pytest_sessionstart：创建 execution
	exec := createExecution(t, env, buildUID)
	if exec.Status != "running" {
		t.Fatalf("execution status = %s, want running", exec.Status)
	}

	// 2. 上报用例，记录各自的 case_uid
	//    case2 重试：同一 case_key 上报两次（不同 case_uid）
	okUID := uuid.New().String()
	retryFirstUID := uuid.New().String()
	retrySecondUID := uuid.New().String()
	failUID := uuid.New().String()

	createItem(t, env, buildUID, "JIRA-OK", okUID)
	createItem(t, env, buildUID, "JIRA-RETRY", retryFirstUID)
	createItem(t, env, buildUID, "JIRA-RETRY", retrySecondUID) // 重试
	createItem(t, env, buildUID, "JIRA-FAIL", failUID)

	// 3. 结束用例：case1 passed、case2 首次 failed 第二次 passed、case3 failed
	updateItem(t, env, buildUID, okUID, "passed")
	updateItem(t, env, buildUID, retryFirstUID, "failed")
	updateItem(t, env, buildUID, retrySecondUID, "passed")
	updateItem(t, env, buildUID, failUID, "failed")

	// 4. 上报 steps（给 case1 附加嵌套步骤）——需要 admin token，先注册一次
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	addNestedSteps(t, env, okUID)

	// 5. 结束 execution：completed
	finishExecution(t, env, buildUID, "completed")

	// 6. 验证：execution 查询返回计数回填
	rows := listExecutionsWithToken(t, env, admin.AccessToken)
	var found *executionRowData
	for i := range rows {
		if rows[i].BuildUID == buildUID {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatalf("execution %s not found after finish", buildUID)
	}
	// case2 最终 attempt 是 passed（重试成功），所以成功 2、失败 1
	if found.PassCount == nil || *found.PassCount != 2 {
		v := 0
		if found.PassCount != nil {
			v = *found.PassCount
		}
		t.Errorf("pass_count = %d, want 2 (含重试成功)", v)
	}
	if found.FailureCount == nil || *found.FailureCount != 1 {
		v := 0
		if found.FailureCount != nil {
			v = *found.FailureCount
		}
		t.Errorf("failure_count = %d, want 1", v)
	}

	// 7. 验证 attempts 查询：case2 应有 2 条 attempt 历史
	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/items/attempts?case_key=JIRA-RETRY", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("attempts query: status=%d", w.Code)
	}
	var attempts struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &attempts); err != nil {
		t.Fatalf("decode attempts: %v", err)
	}
	if attempts.Total != 2 {
		t.Errorf("JIRA-RETRY attempts = %d, want 2", attempts.Total)
	}

	// 8. 验证 execution 详情包含嵌套步骤树
	detail := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/items/"+okUID, "", bearer(admin.AccessToken))
	if detail.Code != http.StatusOK {
		t.Fatalf("item detail: status=%d", detail.Code)
	}
	if !containsStr(detail.Body.String(), "0.1.2") {
		t.Error("item detail must contain nested step 0.1.2")
	}

	// 9. 验证 Dashboard 概览能读到数据（execution 已结束，进入 7 天统计）
	overview := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/dashboard/overview", "", bearer(admin.AccessToken))
	if overview.Code != http.StatusOK {
		t.Fatalf("dashboard overview: status=%d", overview.Code)
	}
	var ov struct {
		Summary struct {
			TotalExecutionsCount int `json:"total_executions_count"`
			SuccessCases7d       int `json:"success_cases_7d"`
			FailureCases7d       int `json:"failure_cases_7d"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(decodeAutomation(t, overview).Data, &ov); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if ov.Summary.TotalExecutionsCount != 1 {
		t.Errorf("dashboard total_executions = %d, want 1", ov.Summary.TotalExecutionsCount)
	}
	if ov.Summary.SuccessCases7d != 2 {
		t.Errorf("dashboard success = %d, want 2", ov.Summary.SuccessCases7d)
	}
	if ov.Summary.FailureCases7d != 1 {
		t.Errorf("dashboard failure = %d, want 1", ov.Summary.FailureCases7d)
	}
}

// addNestedSteps 给指定 case_uid 上报嵌套步骤树。
func addNestedSteps(t *testing.T, e *harness.Env, caseUID string) {
	t.Helper()
	steps := []struct{ path, name string }{
		{"0", "登录"},
		{"0.1", "输入账号"},
		{"0.1.2", "点登录"},
	}
	for _, step := range steps {
		body := `{"step_path":"` + step.path + `","step_name":"` + step.name + `","status":"passed"}`
		w := doAutomation(t, e, http.MethodPost,
			api(e)+"/automation/items/"+caseUID+"/steps", body, reportHeaders(e))
		if w.Code != http.StatusOK {
			t.Fatalf("add step %s: status=%d body=%s", step.path, w.Code, w.Body.String())
		}
	}
}

func containsStr(s, sub string) bool {
	if len(s) < len(sub) {
		return false
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
