package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/test/harness"
)

// 本文件是**缺陷回归测试集**（T-9.14）：用 D{编号} 命名的测试明确标注
// 每个已修复/已规避缺陷的回归验证，防止未来改动重新引入。
//
// 覆盖：D-04、D-05、D-07、D-08、D-13、D-14、D-15、D-16、D-17、D-18、D-26、S-1。

// ---------------------------------------------------------------------------
// D-04 / D-05：attempts 与 case history 返回 ExecutionItemRow 结构
// （Python 路由/响应模型错配，Go 按 service 实际行为返回 ExecutionItemRow 列表）
// ---------------------------------------------------------------------------

// TestD04D05AttemptsReturnsItemRow 验证 attempts 查询返回 ExecutionItemRow 结构。
func TestD04D05AttemptsReturnsItemRow(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1 := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-ATT", u1)
	updateItem(t, env, buildUID, u1, "passed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/items/attempts?case_key=JIRA-ATT", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("D-04/D-05 attempts query: status=%d body=%s", w.Code, w.Body.String())
	}

	// 响应必须是 ExecutionItemRow 结构（含 case_uid 等字段，而非缺 page/page_size 的查询参数模型）
	var list struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &list); err != nil {
		t.Fatalf("decode attempts: %v", err)
	}
	if len(list.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(list.Rows))
	}
	row := list.Rows[0]
	// ExecutionItemRow 应包含 case_uid / case_key / attempt_number / is_latest
	for _, field := range []string{"case_uid", "case_key", "attempt_number", "is_latest", "status"} {
		if _, ok := row[field]; !ok {
			t.Errorf("D-04/D-05 attempts row missing field %q (got %v)", field, row)
		}
	}
}

// ---------------------------------------------------------------------------
// D-07：pipelines 返回 PipelineListData（分页结构），非裸数组
// ---------------------------------------------------------------------------

// TestD07PipelinesReturnsPagedData 验证 pipelines 返回 total/page/page_size 分页结构。
func TestD07PipelinesReturnsPagedData(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	w := doAutomation(t, env, http.MethodGet, api(env)+"/automation/pipelines", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("D-07 pipelines: status=%d body=%s", w.Code, w.Body.String())
	}

	var list struct {
		Total    int `json:"total"`
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Rows     []any
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &list); err != nil {
		t.Fatalf("decode pipelines: %v", err)
	}
	if list.Page < 1 || list.PageSize < 1 {
		t.Errorf("D-07 pipelines must return page/page_size (got %+v)", list)
	}
}

// ---------------------------------------------------------------------------
// D-08：PATCH items/{case_uid} 能正常定位（Python 漏传 case_key 导致 500）
// ---------------------------------------------------------------------------

// TestD08UpdateItemLocatesByCaseUID 验证按 case_uid 更新用例不因缺 case_key 失败。
func TestD08UpdateItemLocatesByCaseUID(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1 := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-D08", u1)

	// 用 case_uid 更新，不传 case_key（Python D-08 会崩）
	body := `{"status":"passed","end_time":"` + nowUTC() + `"}`
	w := doAutomation(t, env, http.MethodPatch,
		api(env)+"/automation/executions/"+buildUID+"/items/"+u1, body, reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("D-08 update item by case_uid: status=%d body=%s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// D-13：execution 幂等创建（ON CONFLICT build_uid）
// ---------------------------------------------------------------------------

// TestD13ExecutionIdempotentCreate 验证重复上报 build_uid 复用已有记录。
func TestD13ExecutionIdempotentCreate(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()

	first := createExecution(t, env, buildUID)
	second := createExecution(t, env, buildUID)
	if first.ID != second.ID {
		t.Errorf("D-13 idempotent create: first=%d second=%d, must reuse", first.ID, second.ID)
	}
}

// ---------------------------------------------------------------------------
// D-14 / D-15：dashboard 字段命名漂移映射（started_at / item_id）
// ---------------------------------------------------------------------------

// TestD14D15DashboardFieldMapping 验证 overview 用 started_at、cases 用 item_id。
func TestD14D15DashboardFieldMapping(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	exec := createExecution(t, env, buildUID)
	u1 := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-D1415", u1)

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	// overview 的 running execution 应含 started_at（D-14）
	ov := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/dashboard/overview", "", bearer(admin.AccessToken))
	if ov.Code != http.StatusOK {
		t.Fatalf("D-14 overview: status=%d", ov.Code)
	}
	var ovData struct {
		RunningExecutions []map[string]any `json:"running_executions"`
	}
	if err := json.Unmarshal(decodeAutomation(t, ov).Data, &ovData); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if len(ovData.RunningExecutions) == 0 {
		t.Fatal("D-14 expected running execution in overview")
	}
	if _, ok := ovData.RunningExecutions[0]["started_at"]; !ok {
		t.Error("D-14 overview must use started_at field (not start_time)")
	}

	// cases 接口应含 item_id（D-15）
	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/dashboard/executions/"+itoa(exec.ID)+"/cases", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("D-15 cases: status=%d", w.Code)
	}
	var cases struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &cases); err != nil {
		t.Fatalf("decode cases: %v", err)
	}
	if len(cases.Items) == 0 {
		t.Fatal("D-15 expected cases")
	}
	if _, ok := cases.Items[0]["item_id"]; !ok {
		t.Error("D-15 cases must use item_id field (not case_id)")
	}
}

// ---------------------------------------------------------------------------
// D-16：状态归类（passed→success、failed+broken→failure、skipped→skipped）
// ---------------------------------------------------------------------------

// TestD16StatusCategorization 验证 dashboard 概览的失败计数按 failed+broken 归类。
func TestD16StatusCategorization(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1, u2, u3 := uuid.New().String(), uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, "JIRA-S1", u1)
	createItem(t, env, buildUID, "JIRA-S2", u2)
	createItem(t, env, buildUID, "JIRA-S3", u3)
	updateItem(t, env, buildUID, u1, "passed")
	updateItem(t, env, buildUID, u2, "failed")
	updateItem(t, env, buildUID, u3, "broken")
	finishExecution(t, env, buildUID, "completed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	ov := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/dashboard/overview", "", bearer(admin.AccessToken))
	if ov.Code != http.StatusOK {
		t.Fatalf("D-16 overview: status=%d", ov.Code)
	}
	var ovData struct {
		Summary struct {
			SuccessCases7d int `json:"success_cases_7d"`
			FailureCases7d int `json:"failure_cases_7d"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(decodeAutomation(t, ov).Data, &ovData); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if ovData.Summary.SuccessCases7d != 1 {
		t.Errorf("D-16 success = %d, want 1", ovData.Summary.SuccessCases7d)
	}
	// failed + broken 都归入 failure
	if ovData.Summary.FailureCases7d != 2 {
		t.Errorf("D-16 failure = %d, want 2 (failed + broken)", ovData.Summary.FailureCases7d)
	}
}

// ---------------------------------------------------------------------------
// D-17：execution labels 落库
// ---------------------------------------------------------------------------

// TestD17ExecutionLabelsPersisted 验证 execution 的 labels 写入并能读回。
func TestD17ExecutionLabelsPersisted(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()

	body := `{
		"build_uid":"` + buildUID + `","job_name":"job-labels","job_url":"http://u",
		"project_name":"p","software_name":"s","software_version":"v",
		"labels":["smoke","regression"],"start_time":"` + nowUTC() + `"
	}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/automation/executions", body, reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("D-17 create execution with labels: status=%d", w.Code)
	}

	// 查 execution 列表，验证 labels 读回
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	rows := listExecutionsWithToken(t, env, admin.AccessToken)
	var found *executionRowData
	for i := range rows {
		if rows[i].BuildUID == buildUID {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatal("D-17 execution not found")
	}
	// 重新查一次带 labels 的完整结构
	w2 := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions?build_uid="+buildUID, "", bearer(admin.AccessToken))
	var list struct {
		Rows []struct {
			Labels []string `json:"labels"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w2).Data, &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list.Rows) != 1 || len(list.Rows[0].Labels) != 2 {
		t.Errorf("D-17 labels = %v, want [smoke regression]", list.Rows)
	}
}

// ---------------------------------------------------------------------------
// D-18：execution 进入终止态时回填计数
// ---------------------------------------------------------------------------

// TestD18ExecutionCountsBackfilledOnFinish 验证结束 execution 回填 pass/fail 计数。
func TestD18ExecutionCountsBackfilledOnFinish(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1, u2 := uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, "JIRA-D18A", u1)
	createItem(t, env, buildUID, "JIRA-D18B", u2)
	updateItem(t, env, buildUID, u1, "passed")
	updateItem(t, env, buildUID, u2, "failed")
	finishExecution(t, env, buildUID, "completed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	rows := listExecutionsWithToken(t, env, admin.AccessToken)
	var found *executionRowData
	for i := range rows {
		if rows[i].BuildUID == buildUID {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatal("D-18 execution not found")
	}
	if found.PassCount == nil || *found.PassCount != 1 {
		t.Errorf("D-18 pass_count = %v, want 1", derefInt(found.PassCount))
	}
	if found.FailureCount == nil || *found.FailureCount != 1 {
		t.Errorf("D-18 failure_count = %v, want 1", derefInt(found.FailureCount))
	}
}

// ---------------------------------------------------------------------------
// D-26：写用 RETURNING，创建后能立即读回完整记录
// ---------------------------------------------------------------------------

// TestD26CreateThenReadBack 验证 item/step 创建后能立即读回（RETURNING 语义）。
func TestD26CreateThenReadBack(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1 := uuid.New().String()
	item := createItem(t, env, buildUID, "JIRA-D26", u1)

	// item 创建后立即可用（有 id、case_uid）
	if item.ID == 0 || item.CaseUID == "" {
		t.Errorf("D-26 created item missing id/case_uid: %+v", item)
	}

	// step 创建后能通过 execution detail 读回
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	updateItem(t, env, buildUID, u1, "passed")
	addNestedSteps(t, env, u1)
	detail := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/items/"+u1, "", bearer(admin.AccessToken))
	if !strings.Contains(detail.Body.String(), "0.1.2") {
		t.Error("D-26 steps not readable after create (RETURNING)")
	}
}

// ---------------------------------------------------------------------------
// S-1：sys_dict 同一 dict_type 可存多行
// ---------------------------------------------------------------------------

// TestS1SameDictTypeMultipleRows 验证 sys_dict 复合唯一允许同 type 多行不同 code。
func TestS1SameDictTypeMultipleRows(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	createDict(t, env, admin.AccessToken, "s1_status", 0, "运行中")
	createDict(t, env, admin.AccessToken, "s1_status", 1, "已通过")
	createDict(t, env, admin.AccessToken, "s1_status", 2, "已失败")

	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/system/dicts?dict_type=s1_status", "", bearer(admin.AccessToken))
	var items []dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("S-1 same dict_type rows = %d, want 3 (no single-col UNIQUE)", len(items))
	}
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
