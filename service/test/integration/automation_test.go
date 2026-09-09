package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	persist "github.com/hermes-platform/go-service/internal/adapter/persistence/automation"
	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	domainautomation "github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/test/harness"
)

// ---------------------------------------------------------------------------
// 工具
// ---------------------------------------------------------------------------

// serviceToken 返回配置的服务令牌（与 Python 侧一致的鉴权方式）。
func serviceToken(e *harness.Env) string { return e.Cfg.Automation.ServiceToken }

// reportHeaders 构造上报接口的鉴权头。
func reportHeaders(e *harness.Env) map[string]string {
	return map[string]string{"X-Service-Token": serviceToken(e)}
}

func doAutomation(t *testing.T, e *harness.Env, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	e.Router(t).ServeHTTP(w, req)
	return w
}

func decodeAutomation(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()
	return decode(t, w)
}

func nowUTC() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// createExecution 上报一次执行并返回其响应数据。
func createExecution(t *testing.T, e *harness.Env, buildUID string) executionData {
	t.Helper()

	body := `{
		"build_uid":"` + buildUID + `",
		"job_name":"job-integration",
		"job_url":"http://jenkins/job-integration",
		"project_name":"proj-x",
		"software_name":"app",
		"software_version":"1.0.0",
		"labels":["smoke"],
		"start_time":"` + nowUTC() + `",
		"planned_cases_count":2
	}`
	w := doAutomation(t, e, http.MethodPost, api(e)+"/automation/executions", body, reportHeaders(e))
	if w.Code != http.StatusOK {
		t.Fatalf("create execution: status=%d body=%s", w.Code, w.Body.String())
	}
	var data executionData
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode execution: %v", err)
	}
	return data
}

type executionData struct {
	ID       int64  `json:"id"`
	BuildUID string `json:"build_uid"`
	Status   string `json:"status"`
}

type itemData struct {
	ID       int64  `json:"id"`
	CaseUID  string `json:"case_uid"`
	Status   string `json:"status"`
	Attempts int    `json:"attempt_number"`
}

// createItem 上报一条用例。
func createItem(t *testing.T, e *harness.Env, buildUID, caseKey, caseUID string) itemData {
	t.Helper()

	body := `{
		"build_uid":"` + buildUID + `",
		"case_key":"` + caseKey + `",
		"case_name":"登录测试",
		"case_uid":"` + caseUID + `"
	}`
	w := doAutomation(t, e, http.MethodPost,
		api(e)+"/automation/executions/"+buildUID+"/items", body, reportHeaders(e))
	if w.Code != http.StatusOK {
		t.Fatalf("create item: status=%d body=%s", w.Code, w.Body.String())
	}
	var data itemData
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	return data
}

// updateItem 结束一条用例。
func updateItem(t *testing.T, e *harness.Env, buildUID, caseUID, status string) {
	t.Helper()

	body := `{"status":"` + status + `","end_time":"` + nowUTC() + `"}`
	w := doAutomation(t, e, http.MethodPatch,
		api(e)+"/automation/executions/"+buildUID+"/items/"+caseUID, body, reportHeaders(e))
	if w.Code != http.StatusOK {
		t.Fatalf("update item: status=%d body=%s", w.Code, w.Body.String())
	}
}

// finishExecution 结束一次执行。
func finishExecution(t *testing.T, e *harness.Env, buildUID, status string) {
	t.Helper()

	body := `{"status":"` + status + `","end_time":"` + nowUTC() + `"}`
	w := doAutomation(t, e, http.MethodPatch,
		api(e)+"/automation/executions/"+buildUID, body, reportHeaders(e))
	if w.Code != http.StatusOK {
		t.Fatalf("finish execution: status=%d body=%s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// ServiceToken 鉴权
// ---------------------------------------------------------------------------

func TestReportRequiresServiceToken(t *testing.T) {
	env := harness.Setup(t)

	body := `{"build_uid":"` + uuid.New().String() + `","job_name":"j","job_url":"u","project_name":"p","software_name":"s","software_version":"v","start_time":"` + nowUTC() + `"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/automation/executions", body, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("report without token: status=%d, want 401; body=%s", w.Code, w.Body.String())
	}

	// 错误令牌同样拒绝
	w = doAutomation(t, env, http.MethodPost, api(env)+"/automation/executions", body,
		map[string]string{"X-Service-Token": "wrong-token"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("report with wrong token: status=%d, want 401", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Execution 上报与幂等
// ---------------------------------------------------------------------------

func TestCreateExecutionStartsRunning(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()

	data := createExecution(t, env, buildUID)
	if data.Status != "running" {
		t.Errorf("status = %s, want running", data.Status)
	}
	if data.BuildUID != buildUID {
		t.Errorf("build_uid = %s, want %s", data.BuildUID, buildUID)
	}
}

// TestCreateExecutionIsIdempotent 是 D-13 的回归：
// 同一 build_uid 重复上报执行必须复用已有记录，不报错。
func TestCreateExecutionIsIdempotent(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()

	first := createExecution(t, env, buildUID)
	second := createExecution(t, env, buildUID)

	if second.ID != first.ID {
		t.Errorf("duplicate build_uid must reuse the same execution: first=%d second=%d",
			first.ID, second.ID)
	}
}

func TestUpdateExecutionFinishAndCount(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	// 上报两个用例并各自结束
	caseUID1, caseUID2 := uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, "JIRA-1", caseUID1)
	createItem(t, env, buildUID, "JIRA-2", caseUID2)
	updateItem(t, env, buildUID, caseUID1, "passed")
	updateItem(t, env, buildUID, caseUID2, "failed")

	// 结束执行
	finishExecution(t, env, buildUID, "completed")

	// 查询详情验证计数回填（Q2）
	rows := listExecutionsRaw(t, env)
	var found *executionRowData
	for i := range rows {
		if rows[i].BuildUID == buildUID {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatalf("execution %s not found in list", buildUID)
	}
	if found.Status != "completed" {
		t.Errorf("status = %s, want completed", found.Status)
	}
	if found.PassCount == nil || *found.PassCount != 1 {
		t.Errorf("pass_count = %v, want 1", found.PassCount)
	}
	if found.FailureCount == nil || *found.FailureCount != 1 {
		t.Errorf("failure_count = %v, want 1", found.FailureCount)
	}
	if found.PassRate == nil || *found.PassRate != 0.5 {
		t.Errorf("pass_rate = %v, want 0.5", found.PassRate)
	}
}

type executionRowData struct {
	ID           int64    `json:"id"`
	BuildUID     string   `json:"build_uid"`
	Status       string   `json:"status"`
	PassCount    *int     `json:"pass_count"`
	FailureCount *int     `json:"failure_count"`
	PassRate     *float64 `json:"pass_rate"`
}

func listExecutionsRaw(t *testing.T, e *harness.Env) []executionRowData {
	t.Helper()
	// 查询接口需要用户 JWT：用一个管理员账号
	admin := registerUser(t, e, "admin", "AdminStrongPass123!", "admin@example.com")
	return listExecutionsWithToken(t, e, admin.AccessToken)
}

// listExecutionsWithToken 用给定 token 查询执行列表。
// 供 E2E 复用已注册的 admin，避免同一测试内重复注册冲突。
func listExecutionsWithToken(t *testing.T, e *harness.Env, token string) []executionRowData {
	t.Helper()
	w := doAutomation(t, e, http.MethodGet, api(e)+"/automation/executions", "",
		bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("list executions: status=%d body=%s", w.Code, w.Body.String())
	}
	var list struct {
		Rows []executionRowData `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &list); err != nil {
		t.Fatalf("decode execution list: %v", err)
	}
	return list.Rows
}

// ---------------------------------------------------------------------------
// Item 上报与 attempt
// ---------------------------------------------------------------------------

// TestItemAttemptNumberIncrements 是 D-03 核心语义：
// 同一用例（同 build_uid + case_key）多次上报，attempt_number 递增且不覆盖历史。
func TestItemAttemptNumberIncrements(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	caseKey := "JIRA-RETRY"
	u1 := uuid.New().String()
	u2 := uuid.New().String()

	first := createItem(t, env, buildUID, caseKey, u1)
	second := createItem(t, env, buildUID, caseKey, u2)

	if first.Attempts != 1 {
		t.Errorf("first attempt = %d, want 1", first.Attempts)
	}
	if second.Attempts != 2 {
		t.Errorf("second attempt = %d, want 2 (D-03 attempt increment)", second.Attempts)
	}
}

// TestItemAttemptsQuery 验证 attempts 查询返回全部重试历史。
func TestItemAttemptsQuery(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	caseKey := "JIRA-HISTORY"
	u1, u2 := uuid.New().String(), uuid.New().String()
	createItem(t, env, buildUID, caseKey, u1)
	createItem(t, env, buildUID, caseKey, u2)
	updateItem(t, env, buildUID, u1, "passed")
	updateItem(t, env, buildUID, u2, "failed")

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/items/attempts?case_key="+caseKey, "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("attempts query: status=%d body=%s", w.Code, w.Body.String())
	}
	var list struct {
		Total int        `json:"total"`
		Rows  []itemData `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &list); err != nil {
		t.Fatalf("decode attempts: %v", err)
	}
	if list.Total != 2 {
		t.Errorf("attempts total = %d, want 2", list.Total)
	}
	// 按 attempt_number 升序
	if list.Rows[0].Attempts != 1 || list.Rows[1].Attempts != 2 {
		t.Errorf("attempts order wrong: %+v", list.Rows)
	}
	if list.Rows[0].Status != "passed" {
		t.Errorf("first attempt status = %s, want passed", list.Rows[0].Status)
	}
}

// ---------------------------------------------------------------------------
// Step 上报
// ---------------------------------------------------------------------------

func TestStepUpsertAndNestedQuery(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	caseUID := uuid.New().String()
	item := createItem(t, env, buildUID, "JIRA-STEPS", caseUID)

	// 批量上报步骤（含嵌套 + 一个重复路径验证幂等）
	stepsBody := `[
		{"case_uid":"` + caseUID + `","step_path":"0","step_name":"登录","status":"passed"},
		{"case_uid":"` + caseUID + `","step_path":"0.1","step_name":"输入账号","status":"passed"},
		{"case_uid":"` + caseUID + `","step_path":"0.1.2","step_name":"点登录","status":"passed"}
	]`
	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+buildUID+"/items/"+itoa(item.ID)+"/steps",
		stepsBody, reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("add steps: status=%d body=%s", w.Code, w.Body.String())
	}

	// 重复上报同一路径（幂等，不报错）
	w = doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+buildUID+"/items/"+itoa(item.ID)+"/steps",
		`[{"case_uid":"`+caseUID+`","step_path":"0","step_name":"登录(更新)","status":"failed"}]`,
		reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("re-update step: status=%d body=%s", w.Code, w.Body.String())
	}

	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	detail := doAutomation(t, env, http.MethodGet, api(env)+"/automation/items/"+caseUID, "",
		bearer(admin.AccessToken))
	if detail.Code != http.StatusOK {
		t.Fatalf("item detail: status=%d body=%s", detail.Code, detail.Body.String())
	}
	body := detail.Body.String()
	if !strings.Contains(body, "0.1.2") {
		t.Errorf("item must contain nested step 0.1.2, got %s", body)
	}
	if !strings.Contains(body, `"0"`) {
		t.Errorf("item must contain root step 0, got %s", body)
	}
}

// ---------------------------------------------------------------------------
// Pipeline 同步
// ---------------------------------------------------------------------------

func TestPipelineSyncedFromExecution(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	finishExecution(t, env, buildUID, "completed")

	// 流水线应随 execution 上报自动创建
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	w := doAutomation(t, env, http.MethodGet, api(env)+"/automation/pipelines", "",
		bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("list pipelines: status=%d body=%s", w.Code, w.Body.String())
	}
	var list struct {
		Rows []struct {
			JobName         string  `json:"job_name"`
			LastBuildStatus *string `json:"last_build_status"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &list); err != nil {
		t.Fatalf("decode pipelines: %v", err)
	}
	if len(list.Rows) != 1 {
		t.Fatalf("pipelines = %d, want 1", len(list.Rows))
	}
	if list.Rows[0].JobName != "job-integration" {
		t.Errorf("job_name = %s, want job-integration", list.Rows[0].JobName)
	}
	if list.Rows[0].LastBuildStatus == nil || *list.Rows[0].LastBuildStatus != "success" {
		t.Errorf("last_build_status = %v, want success", list.Rows[0].LastBuildStatus)
	}
}

func TestHeartbeatRequiresServiceToken(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+buildUID+"/heartbeat", "{}", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("heartbeat without token: status=%d, want 401", w.Code)
	}
}

func TestHeartbeatUpdatesRunningExecution(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+buildUID+"/heartbeat", "{}", reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("heartbeat: status=%d body=%s", w.Code, w.Body.String())
	}
	var data struct {
		LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err != nil {
		t.Fatalf("decode heartbeat: %v", err)
	}
	if data.LastHeartbeatAt.IsZero() {
		t.Fatal("last_heartbeat_at must be set")
	}
}

func TestHeartbeatUnknownExecutionNotFound(t *testing.T) {
	env := harness.Setup(t)
	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+uuid.New().String()+"/heartbeat", "{}", reportHeaders(env))
	if w.Code != http.StatusNotFound {
		t.Fatalf("heartbeat missing execution: status=%d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestHeartbeatIgnoredAfterFinish(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	finishExecution(t, env, buildUID, "completed")

	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/executions/"+buildUID+"/heartbeat", "{}", reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("heartbeat after finish: status=%d body=%s", w.Code, w.Body.String())
	}

	admin := registerUser(t, env, "hb-admin", "AdminStrongPass123!", "hb-admin@example.com")
	list := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions?build_uid="+buildUID, "", bearer(admin.AccessToken))
	if list.Code != http.StatusOK {
		t.Fatalf("list executions: status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Rows []struct {
			Status string `json:"status"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, list).Data, &payload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(payload.Rows) != 1 || payload.Rows[0].Status != "completed" {
		t.Fatalf("status after heartbeat = %+v, want completed", payload.Rows)
	}
}

func TestAbortStaleRunningExecutions(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	ctx := context.Background()
	if _, err := env.DB.Exec(ctx,
		`UPDATE test_executions SET last_heartbeat_at = now() - interval '2 minutes' WHERE build_uid = $1`,
		buildUID); err != nil {
		t.Fatalf("age heartbeat: %v", err)
	}

	n, _, err := env.App(t).Services.AutomationIngest.AbortStale(ctx, time.Minute)
	if err != nil {
		t.Fatalf("abort stale: %v", err)
	}
	if n != 1 {
		t.Fatalf("aborted = %d, want 1", n)
	}

	admin := registerUser(t, env, "abort-admin", "AdminStrongPass123!", "abort-admin@example.com")
	list := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions?build_uid="+buildUID, "", bearer(admin.AccessToken))
	if list.Code != http.StatusOK {
		t.Fatalf("list executions: status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Rows []struct {
			Status string `json:"status"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, list).Data, &payload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(payload.Rows) != 1 || payload.Rows[0].Status != "aborted" {
		t.Fatalf("status after abort = %+v, want aborted", payload.Rows)
	}
}

type recordingLivePublisher struct {
	events []domainautomation.Event
}

func (r *recordingLivePublisher) Publish(_ context.Context, event domainautomation.Event) error {
	r.events = append(r.events, event)
	return nil
}

type ingestTestClock struct{}

func (ingestTestClock) Now() time.Time { return time.Now().UTC() }

func TestIngestStepPublishesLiveEvent(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New()
	createExecution(t, env, buildUID.String())
	caseUID := uuid.New()
	createItem(t, env, buildUID.String(), "JIRA-LIVE", caseUID.String())

	live := &recordingLivePublisher{}
	ingest := appautomation.NewIngestUseCase(
		persist.NewExecutionRepo(env.DB),
		persist.NewItemRepo(env.DB),
		persist.NewStepRepo(env.DB),
		persist.NewAttachmentRepo(env.DB),
		persist.NewPipelineRepo(env.DB),
		ingestTestClock{},
		live,
	)

	path, err := domainautomation.ParseStepPath("0.1")
	if err != nil {
		t.Fatalf("parse path: %v", err)
	}
	if _, err := ingest.IngestStep(context.Background(), appautomation.IngestStepCommand{
		CaseUID:  caseUID,
		StepPath: path,
		StepName: "输入账号",
		Status:   domainautomation.CaseStatusRunning,
	}); err != nil {
		t.Fatalf("ingest step: %v", err)
	}

	if len(live.events) != 1 {
		t.Fatalf("live events = %d, want 1", len(live.events))
	}
	ev, ok := live.events[0].(domainautomation.StepUpserted)
	if !ok {
		t.Fatalf("event type = %T, want StepUpserted", live.events[0])
	}
	if ev.BuildUID != buildUID {
		t.Errorf("build_uid = %s, want %s", ev.BuildUID, buildUID)
	}
	if ev.CaseUID != caseUID {
		t.Errorf("case_uid = %s, want %s", ev.CaseUID, caseUID)
	}
	if ev.StepPath != "0.1" || ev.StepName != "输入账号" || ev.Status != "running" {
		t.Errorf("step event = %+v", ev)
	}
	if ev.CaseName != "登录测试" || ev.CaseKey != "JIRA-LIVE" {
		t.Errorf("step case identity = %+v, want 登录测试 / JIRA-LIVE", ev)
	}
}

func TestIngestItemPublishesLiveEvent(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New()
	createExecution(t, env, buildUID.String())
	caseUID := uuid.New()

	live := &recordingLivePublisher{}
	ingest := appautomation.NewIngestUseCase(
		persist.NewExecutionRepo(env.DB),
		persist.NewItemRepo(env.DB),
		persist.NewStepRepo(env.DB),
		persist.NewAttachmentRepo(env.DB),
		persist.NewPipelineRepo(env.DB),
		ingestTestClock{},
		live,
	)

	if _, err := ingest.IngestItem(context.Background(), appautomation.IngestItemCommand{
		BuildUID: buildUID,
		CaseUID:  caseUID,
		CaseKey:  "JIRA-ITEM-LIVE",
		CaseName: "登录测试",
		Status:   domainautomation.CaseStatusRunning,
	}); err != nil {
		t.Fatalf("ingest item: %v", err)
	}

	if len(live.events) != 1 {
		t.Fatalf("live events = %d, want 1", len(live.events))
	}
	ev, ok := live.events[0].(domainautomation.ItemUpdated)
	if !ok {
		t.Fatalf("event type = %T, want ItemUpdated", live.events[0])
	}
	if ev.BuildUID != buildUID || ev.CaseUID != caseUID {
		t.Errorf("uids = %s/%s, want %s/%s", ev.BuildUID, ev.CaseUID, buildUID, caseUID)
	}
	if ev.CaseName != "登录测试" || ev.CaseKey != "JIRA-ITEM-LIVE" || ev.Status != "running" {
		t.Errorf("item event = %+v", ev)
	}
}

func TestUpsertStepByCaseUID(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	caseUID := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-CASE-STEP", caseUID)

	body := `{"step_path":"0","step_name":"打开页面","status":"running","attachments":[{"attachment_type":"log","file_name":"x","url":"http://ignored.example/x.log"}]}`
	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/items/"+caseUID+"/steps", body, reportHeaders(env))
	if w.Code != http.StatusOK {
		t.Fatalf("upsert step: status=%d body=%s", w.Code, w.Body.String())
	}

	admin := registerUser(t, env, "step-admin", "AdminStrongPass123!", "step-admin@example.com")
	detail := doAutomation(t, env, http.MethodGet, api(env)+"/automation/items/"+caseUID, "",
		bearer(admin.AccessToken))
	if detail.Code != http.StatusOK {
		t.Fatalf("item detail: status=%d body=%s", detail.Code, detail.Body.String())
	}
	got := detail.Body.String()
	if !strings.Contains(got, "打开页面") {
		t.Errorf("item must contain upserted step, got %s", got)
	}
	if strings.Contains(got, "ignored.example") {
		t.Errorf("attachments must be ignored, got %s", got)
	}
}

func TestUpsertStepRejectsSixthLevel(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	caseUID := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-DEEP", caseUID)

	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/items/"+caseUID+"/steps",
		`{"step_path":"0.1.2.3.4.5","step_name":"过深","status":"running"}`,
		reportHeaders(env))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sixth-level step: status=%d, want 422; body=%s", w.Code, w.Body.String())
	}
}

func TestUpsertStepUnknownCaseNotFound(t *testing.T) {
	env := harness.Setup(t)
	w := doAutomation(t, env, http.MethodPost,
		api(env)+"/automation/items/"+uuid.New().String()+"/steps",
		`{"step_path":"0","step_name":"幽灵","status":"running"}`,
		reportHeaders(env))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown case: status=%d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestStreamEventsRequiresAuth(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/events", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("sse without token: status=%d, want 401", w.Code)
	}
}

func TestStepUpsertReachesSSE(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	caseUID := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-SSE", caseUID)

	srv := httptest.NewServer(env.Router(t))
	t.Cleanup(srv.Close)

	admin := registerUser(t, env, "sse-admin", "AdminStrongPass123!", "sse-admin@example.com")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		srv.URL+api(env)+"/automation/executions/"+buildUID+"/events", nil)
	if err != nil {
		t.Fatalf("new sse request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+admin.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("sse status=%d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	if _, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, ":")
	}); err != nil {
		t.Fatalf("wait connected: %v", err)
	}

	post, err := http.NewRequest(http.MethodPost,
		srv.URL+api(env)+"/automation/items/"+caseUID+"/steps",
		strings.NewReader(`{"step_path":"0","step_name":"SSE步骤","status":"running"}`))
	if err != nil {
		t.Fatalf("new post: %v", err)
	}
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("X-Service-Token", serviceToken(env))
	postResp, err := http.DefaultClient.Do(post)
	if err != nil {
		t.Fatalf("upsert for sse: %v", err)
	}
	_ = postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		t.Fatalf("upsert for sse: status=%d", postResp.StatusCode)
	}

	dataLine, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, "data:")
	})
	if err != nil {
		t.Fatalf("wait sse data: %v", err)
	}
	if !strings.Contains(dataLine, `"type":"step.upserted"`) ||
		!strings.Contains(dataLine, "SSE步骤") ||
		!strings.Contains(dataLine, "登录测试") {
		t.Fatalf("sse data = %q, want step.upserted with case and step name", dataLine)
	}
}

func TestItemAndExecutionUpdatesReachSSE(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	caseUID := uuid.New().String()

	srv := httptest.NewServer(env.Router(t))
	t.Cleanup(srv.Close)

	admin := registerUser(t, env, "sse-item-admin", "AdminStrongPass123!", "sse-item-admin@example.com")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		srv.URL+api(env)+"/automation/executions/"+buildUID+"/events", nil)
	if err != nil {
		t.Fatalf("new sse request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+admin.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("sse status=%d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	if _, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, ":")
	}); err != nil {
		t.Fatalf("wait connected: %v", err)
	}

	createItem(t, env, buildUID, "JIRA-SSE-ITEM", caseUID)
	createdLine, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, "data:")
	})
	if err != nil {
		t.Fatalf("wait item created sse: %v", err)
	}
	if !strings.Contains(createdLine, `"type":"item.updated"`) ||
		!strings.Contains(createdLine, "登录测试") ||
		!strings.Contains(createdLine, `"status":"running"`) {
		t.Fatalf("sse create item = %q, want item.updated with case name", createdLine)
	}

	updateItem(t, env, buildUID, caseUID, "passed")
	updatedLine, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, "data:")
	})
	if err != nil {
		t.Fatalf("wait item passed sse: %v", err)
	}
	if !strings.Contains(updatedLine, `"type":"item.updated"`) ||
		!strings.Contains(updatedLine, `"status":"passed"`) {
		t.Fatalf("sse update item = %q, want passed item.updated", updatedLine)
	}

	finishExecution(t, env, buildUID, "completed")
	finishedLine, err := waitSSELine(t, reader, 3*time.Second, func(line string) bool {
		return strings.HasPrefix(line, "data:")
	})
	if err != nil {
		t.Fatalf("wait execution sse: %v", err)
	}
	if !strings.Contains(finishedLine, `"type":"execution.updated"`) ||
		!strings.Contains(finishedLine, `"status":"completed"`) {
		t.Fatalf("sse finish execution = %q, want execution.updated completed", finishedLine)
	}
}

func waitSSELine(t *testing.T, reader *bufio.Reader, timeout time.Duration, match func(string) bool) (string, error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if match(strings.TrimRight(line, "\r\n")) {
			return line, nil
		}
	}
	return "", context.DeadlineExceeded
}

func TestLiveSnapshotAndExecutionRow(t *testing.T) {
	env := harness.Setup(t)
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)

	passedUID := uuid.New().String()
	runningUID := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-DONE", passedUID)
	createItem(t, env, buildUID, "JIRA-LIVE", runningUID)
	updateItem(t, env, buildUID, passedUID, "passed")

	for _, step := range []struct{ path, name, status string }{
		{"0", "打开页面", "passed"},
		{"0.1", "输入账号", "running"},
	} {
		body := `{"step_path":"` + step.path + `","step_name":"` + step.name + `","status":"` + step.status + `"}`
		w := doAutomation(t, env, http.MethodPost,
			api(env)+"/automation/items/"+runningUID+"/steps", body, reportHeaders(env))
		if w.Code != http.StatusOK {
			t.Fatalf("upsert %s: status=%d body=%s", step.path, w.Code, w.Body.String())
		}
	}

	admin := registerUser(t, env, "live-admin", "AdminStrongPass123!", "live-admin@example.com")

	execW := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID, "", bearer(admin.AccessToken))
	if execW.Code != http.StatusOK {
		t.Fatalf("get execution: status=%d body=%s", execW.Code, execW.Body.String())
	}
	var execRow struct {
		BuildUID string `json:"build_uid"`
		Status   string `json:"status"`
		Rows     any    `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, execW).Data, &execRow); err != nil {
		t.Fatalf("decode execution: %v", err)
	}
	if execRow.BuildUID != buildUID || execRow.Status != "running" {
		t.Fatalf("execution row = %+v", execRow)
	}
	if execRow.Rows != nil {
		t.Fatalf("GET execution must not return items rows, got %v", execRow.Rows)
	}

	listW := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/items", "", bearer(admin.AccessToken))
	if listW.Code != http.StatusOK {
		t.Fatalf("list items: status=%d body=%s", listW.Code, listW.Body.String())
	}
	listBody := listW.Body.String()
	if strings.Contains(listBody, "输入账号") {
		t.Fatalf("item summaries must not include step tree, got %s", listBody)
	}
	var list struct {
		Total int `json:"total"`
		Rows  []struct {
			CaseUID string `json:"case_uid"`
			Status  string `json:"status"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(decodeAutomation(t, listW).Data, &list); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	if list.Total != 2 || len(list.Rows) != 2 {
		t.Fatalf("items total=%d rows=%d, want 2", list.Total, len(list.Rows))
	}

	liveW := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/live", "", bearer(admin.AccessToken))
	if liveW.Code != http.StatusOK {
		t.Fatalf("live: status=%d body=%s", liveW.Code, liveW.Body.String())
	}
	var live struct {
		Execution struct {
			BuildUID string `json:"build_uid"`
			Status   string `json:"status"`
		} `json:"execution"`
		Items []struct {
			CaseUID string `json:"case_uid"`
			Status  string `json:"status"`
		} `json:"items"`
		CurrentItem *struct {
			CaseUID string `json:"case_uid"`
			Status  string `json:"status"`
			Steps   []struct {
				StepPath string `json:"step_path"`
				StepName string `json:"step_name"`
				Status   string `json:"status"`
				SubSteps []struct {
					StepPath string `json:"step_path"`
					StepName string `json:"step_name"`
					Status   string `json:"status"`
				} `json:"sub_steps"`
			} `json:"steps"`
		} `json:"current_item"`
	}
	if err := json.Unmarshal(decodeAutomation(t, liveW).Data, &live); err != nil {
		t.Fatalf("decode live: %v body=%s", err, liveW.Body.String())
	}
	if live.Execution.BuildUID != buildUID || live.Execution.Status != "running" {
		t.Fatalf("live execution = %+v", live.Execution)
	}
	if len(live.Items) != 2 {
		t.Fatalf("live items = %d, want 2", len(live.Items))
	}
	if live.CurrentItem == nil {
		t.Fatal("current_item is nil, want running case")
	}
	if live.CurrentItem.CaseUID != runningUID || live.CurrentItem.Status != "running" {
		t.Fatalf("current_item = %+v, want case %s running", live.CurrentItem, runningUID)
	}
	if len(live.CurrentItem.Steps) != 1 || live.CurrentItem.Steps[0].StepPath != "0" {
		t.Fatalf("current steps = %+v, want root 0", live.CurrentItem.Steps)
	}
	if len(live.CurrentItem.Steps[0].SubSteps) != 1 || live.CurrentItem.Steps[0].SubSteps[0].StepPath != "0.1" {
		t.Fatalf("nested step = %+v, want 0.1 running", live.CurrentItem.Steps[0].SubSteps)
	}

	updateItem(t, env, buildUID, runningUID, "passed")
	doneW := doAutomation(t, env, http.MethodGet,
		api(env)+"/automation/executions/"+buildUID+"/live", "", bearer(admin.AccessToken))
	var done struct {
		CurrentItem *struct{} `json:"current_item"`
	}
	if err := json.Unmarshal(decodeAutomation(t, doneW).Data, &done); err != nil {
		t.Fatalf("decode live after finish: %v", err)
	}
	if done.CurrentItem != nil {
		t.Fatalf("current_item after all finished = %+v, want null", done.CurrentItem)
	}
}

// ensure config 引用保持（避免误删依赖）。
var _ = config.EnvLocal
