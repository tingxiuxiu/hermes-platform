// Package automation 承载 automation 上下文的 HTTP 处理器。
//
// 职责边界：只做「HTTP ↔ 用例」的翻译。业务规则一律在用例/领域层。
package automation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	dtoautomation "github.com/hermes-platform/go-service/internal/adapter/http/dto/automation"
	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	domainautomation "github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// Handlers 聚合 automation 上下文的全部处理器。
type Handlers struct {
	ingest     *appautomation.IngestUseCase
	query      *appautomation.QueryUseCase
	publisher  appautomation.EventPublisher
	items      appautomation.ItemRepository
	subscriber appautomation.LiveSubscriber
}

// NewHandlers 构造 automation 处理器。subscriber 可为 nil（SSE 返回 503）。
func NewHandlers(
	ingest *appautomation.IngestUseCase,
	query *appautomation.QueryUseCase,
	publisher appautomation.EventPublisher,
	items appautomation.ItemRepository,
	subscriber appautomation.LiveSubscriber,
) *Handlers {
	return &Handlers{
		ingest: ingest, query: query, publisher: publisher,
		items: items, subscriber: subscriber,
	}
}

// publish 派发用例层返回的事件。发布失败只记日志，不阻塞业务响应。
func (h *Handlers) publish(c *gin.Context, events []domainautomation.Event) {
	for _, ev := range events {
		if err := h.publisher.Publish(c.Request.Context(), ev); err != nil {
			// 事件发布失败不影响上报结果：写侧已完成，读模型稍后重算即可
			_ = ev
		}
	}
}

// ---------------------------------------------------------------------------
// 上报类接口（ServiceToken 鉴权）
// ---------------------------------------------------------------------------

// CreateExecution 处理 POST /automation/executions（pytest_sessionstart）。
func (h *Handlers) CreateExecution(c *gin.Context) {
	var req dtoautomation.CreateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	planned := 0
	if req.PlannedCasesCount != nil {
		planned = *req.PlannedCasesCount
	}

	events, err := h.ingest.IngestExecution(c.Request.Context(), appautomation.IngestExecutionCommand{
		BuildUID:          req.BuildUID,
		JobName:           req.JobName,
		JobURL:            req.JobURL,
		ProjectName:       req.ProjectName,
		SoftwareName:      req.SoftwareName,
		SoftwareVersion:   req.SoftwareVersion,
		Labels:            req.Labels,
		PlannedCasesCount: planned,
		StartTime:         req.StartTime,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	h.publish(c, events)

	// 响应需返回创建的 execution 信息；用 build_uid 反查最新一条
	exec, err := h.query.ExecutionByBuildUID(c.Request.Context(), req.BuildUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "创建执行成功", dtoautomation.ToExecutionRow(exec))
}

// UpdateExecution 处理 PATCH /automation/executions/{build_uid}（pytest_sessionfinish）。
func (h *Handlers) UpdateExecution(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}

	var req dtoautomation.UpdateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	status := domainautomation.ExecutionStatusRunning
	if req.Status != nil {
		status = *req.Status
	}
	end := req.EndTime

	events, err := h.ingest.FinishExecution(c.Request.Context(), appautomation.FinishExecutionCommand{
		BuildUID: buildUID,
		Status:   status,
		EndTime:  derefTime(end),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	h.publish(c, events)

	exec, err := h.query.ExecutionByBuildUID(c.Request.Context(), buildUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "更新执行成功", dtoautomation.ToExecutionRow(exec))
}

// Heartbeat 处理 POST /automation/executions/{build_uid}/heartbeat。
func (h *Handlers) Heartbeat(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}

	at, err := h.ingest.Heartbeat(c.Request.Context(), buildUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "心跳已记录", dtoautomation.HeartbeatData{LastHeartbeatAt: at})
}

// CreateItem 处理 POST /automation/executions/{build_uid}/items。
func (h *Handlers) CreateItem(c *gin.Context) {
	buildUIDStr := c.Param("build_uid")
	buildUID, err := parseUUID(buildUIDStr)
	if err != nil {
		response.Fail(c, 422, "路径参数不合法")
		return
	}

	var req dtoautomation.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	// 路径参数优先，body 里的 build_uid 仅作兼容
	cmdBuildUID := buildUID
	if b, perr := parseUUID(req.BuildUID); perr == nil {
		cmdBuildUID = b
	}

	events, err := h.ingest.IngestItem(c.Request.Context(), appautomation.IngestItemCommand{
		BuildUID:  cmdBuildUID,
		CaseUID:   req.CaseUID,
		CaseKey:   req.CaseKey,
		CaseName:  req.CaseName,
		Labels:    req.Labels,
		Status:    domainautomation.CaseStatusRunning,
		StartTime: req.StartTime,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	h.publish(c, events)

	// 返回创建的 item 信息
	item, err := h.query.ItemByCaseUID(c.Request.Context(), req.CaseUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "创建用例成功", dtoautomation.ToExecutionItemRow(item))
}

// UpdateItem 处理 PATCH /automation/executions/{build_uid}/items/{case_uid}。
func (h *Handlers) UpdateItem(c *gin.Context) {
	caseUID, ok := pathUUID(c, "case_uid")
	if !ok {
		return
	}

	var req dtoautomation.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	events, err := h.ingest.IngestItem(c.Request.Context(), appautomation.IngestItemCommand{
		BuildUID:       uuidFromString(c.Param("build_uid")),
		CaseUID:        caseUID,
		Status:         req.Status,
		EndTime:        req.EndTime,
		ErrorMessage:   req.ErrorMessage,
		ErrorTraceback: req.ErrorTraceback,
		Attachments:    toAttachmentSpecs(req.Attachments),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	h.publish(c, events)

	item, err := h.query.ItemByCaseUID(c.Request.Context(), caseUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "更新用例成功", dtoautomation.ToExecutionItemRow(item))
}

// AddSteps 处理 POST /automation/executions/{build_uid}/items/{item_id}/steps。
func (h *Handlers) AddSteps(c *gin.Context) {
	itemID, ok := pathID(c, "item_id")
	if !ok {
		return
	}

	// 校验 item 存在（Python 版按 item_id 定位用例）。
	// 步骤体用 case_uid 标识归属；这里同时校验两者指向同一用例，防止误上报。
	item, err := h.query.ItemByID(c.Request.Context(), itemID)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req []dtoautomation.CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	for _, step := range req {
		// 校验 body 的 case_uid 与 item 一致
		if step.CaseUID != item.CaseUID() {
			response.Fail(c, 422, "步骤的 case_uid 与路径 item_id 不匹配")
			return
		}

		path, err := domainautomation.ParseStepPath(step.StepPath)
		if err != nil {
			response.Error(c, err)
			return
		}
		if _, err := h.ingest.IngestStep(c.Request.Context(), appautomation.IngestStepCommand{
			CaseUID:  step.CaseUID,
			StepPath: path,
			StepName: step.StepName,
			Status:   step.Status,
			Start:    step.StartTime,
			End:      step.EndTime,
			Duration: step.Duration,
		}); err != nil {
			response.Error(c, err)
			return
		}
	}

	response.OKWithMessage(c, "创建步骤成功", nil)
}

// UpsertStep 处理 POST /automation/items/{case_uid}/steps（live 单步上报）。
func (h *Handlers) UpsertStep(c *gin.Context) {
	caseUID, ok := pathUUID(c, "case_uid")
	if !ok {
		return
	}

	var req dtoautomation.UpsertStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	path, err := domainautomation.ParseStepPath(req.StepPath)
	if err != nil {
		response.Error(c, err)
		return
	}

	if _, err := h.ingest.IngestStep(c.Request.Context(), appautomation.IngestStepCommand{
		CaseUID:  caseUID,
		StepPath: path,
		StepName: req.StepName,
		Status:   req.Status,
		Start:    req.StartTime,
		End:      req.EndTime,
		Duration: req.Duration,
	}); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "上报步骤成功", nil)
}

// ---------------------------------------------------------------------------
// 查询类接口（用户 JWT 鉴权）
// ---------------------------------------------------------------------------

// ListPipelines 处理 GET /automation/pipelines。
func (h *Handlers) ListPipelines(c *gin.Context) {
	var q struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		JobName  string `form:"job_name"`
		Status   string `form:"status"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	var statusPtr *domainautomation.PipelineStatus
	if q.Status != "" {
		s := domainautomation.PipelineStatus(q.Status)
		statusPtr = &s
	}

	rows, total, err := h.query.ListPipelines(c.Request.Context(), appautomation.PipelineListFilter{
		JobName:  q.JobName,
		Status:   statusPtr,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	page, pageSize := defaultPage(q.Page, q.PageSize)
	items := make([]dtoautomation.PipelineRow, 0, len(rows))
	for _, p := range rows {
		items = append(items, dtoautomation.ToPipelineRow(p))
	}
	response.OKWithMessage(c, "获取自动化任务列表成功", dtoautomation.PipelineListData{
		Total: int(total), Page: page, PageSize: pageSize, Rows: items,
	})
}

// ListExecutions 处理 GET /automation/executions。
func (h *Handlers) ListExecutions(c *gin.Context) {
	var q struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		BuildUID string `form:"build_uid"`
		JobName  string `form:"job_name"`
		Status   string `form:"status"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	filter := appautomation.ExecutionListFilter{
		JobName:  q.JobName,
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	if q.BuildUID != "" {
		if u, err := parseUUID(q.BuildUID); err == nil {
			filter.BuildUID = &u
		}
	}
	if q.Status != "" {
		if s, err := domainautomation.ParseExecutionStatus(q.Status); err == nil {
			filter.Status = &s
		}
	}

	rows, total, err := h.query.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	page, pageSize := defaultPage(q.Page, q.PageSize)
	items := make([]dtoautomation.ExecutionRow, 0, len(rows))
	for _, e := range rows {
		items = append(items, dtoautomation.ToExecutionRow(e))
	}
	response.OKWithMessage(c, "获取执行列表成功", dtoautomation.ExecutionListData{
		Total: int(total), Page: page, PageSize: pageSize, Rows: items,
	})
}

// GetExecution 处理 GET /automation/executions/{build_uid}（仅 execution 行）。
func (h *Handlers) GetExecution(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}

	exec, err := h.query.ExecutionByBuildUID(c.Request.Context(), buildUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "获取执行详情成功", dtoautomation.ToExecutionRow(exec))
}

// ListExecutionItems 处理 GET /automation/executions/{build_uid}/items（分页摘要，不带步骤）。
func (h *Handlers) ListExecutionItems(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}

	var q struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	page, pageSize := defaultPage(q.Page, q.PageSize)
	items, total, err := h.query.ListExecutionItems(c.Request.Context(), buildUID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	rows := make([]dtoautomation.ItemSummaryRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, dtoautomation.ToItemSummary(item))
	}
	response.OKWithMessage(c, "获取用例列表成功", dtoautomation.ItemSummaryListData{
		Total: int(total), Page: page, PageSize: pageSize, Rows: rows,
	})
}

// GetLive 处理 GET /automation/executions/{build_uid}/live。
func (h *Handlers) GetLive(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}

	snap, err := h.query.Live(c.Request.Context(), buildUID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dtoautomation.ItemSummaryRow, 0, len(snap.Items))
	for _, item := range snap.Items {
		items = append(items, dtoautomation.ToItemSummary(item))
	}
	data := dtoautomation.LiveSnapshotData{
		Execution: dtoautomation.ToExecutionRow(snap.Execution),
		Items:     items,
	}
	if snap.CurrentItem != nil {
		cur := dtoautomation.ToLiveCurrentItem(snap.CurrentItem)
		data.CurrentItem = &cur
	}
	response.OKWithMessage(c, "获取实时快照成功", data)
}

// GetItem 处理 GET /automation/items/{case_uid}（用例 + 步骤树）。
func (h *Handlers) GetItem(c *gin.Context) {
	caseUID, ok := pathUUID(c, "case_uid")
	if !ok {
		return
	}

	item, err := h.query.ItemByCaseUID(c.Request.Context(), caseUID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "获取用例详情成功", dtoautomation.ToExecutionItemRow(item))
}

// StreamEvents 处理 GET /automation/executions/{build_uid}/events（SSE）。
func (h *Handlers) StreamEvents(c *gin.Context) {
	if h.subscriber == nil {
		response.Fail(c, http.StatusServiceUnavailable, "实时通道不可用")
		return
	}

	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}
	if _, err := h.query.ExecutionByBuildUID(c.Request.Context(), buildUID); err != nil {
		response.Error(c, err)
		return
	}

	frames, err := h.subscriber.Subscribe(c.Request.Context(), buildUID)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, "订阅实时通道失败")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if _, err := c.Writer.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case frame, ok := <-frames:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", frame.Event, frame.JSON); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

// GetCaseAttempts 处理 GET /automation/executions/{build_uid}/items/attempts。
func (h *Handlers) GetCaseAttempts(c *gin.Context) {
	buildUID, ok := pathUUID(c, "build_uid")
	if !ok {
		return
	}
	caseKey := c.Query("case_key")
	if caseKey == "" {
		response.Fail(c, 422, "缺少 case_key 查询参数")
		return
	}

	items, err := h.query.ListAttempts(c.Request.Context(), buildUID, caseKey)
	if err != nil {
		response.Error(c, err)
		return
	}

	rows := make([]dtoautomation.ExecutionItemRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, dtoautomation.ToExecutionItemRow(item))
	}
	response.OKWithMessage(c, "获取用例重试历史成功", dtoautomation.ExecutionItemListData{
		Total: len(rows), Page: 1, PageSize: len(rows), Rows: rows,
	})
}

// GetCaseHistory 处理 GET /automation/cases/history。
func (h *Handlers) GetCaseHistory(c *gin.Context) {
	caseKey := c.Query("case_key")
	if caseKey == "" {
		response.Fail(c, 422, "缺少 case_key 查询参数")
		return
	}

	items, err := h.query.ListCasesHistory(c.Request.Context(), caseKey)
	if err != nil {
		response.Error(c, err)
		return
	}

	rows := make([]dtoautomation.ExecutionItemRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, dtoautomation.ToExecutionItemRow(item))
	}
	response.OKWithMessage(c, "获取用例历史趋势成功", dtoautomation.ExecutionItemListData{
		Total: len(rows), Page: 1, PageSize: len(rows), Rows: rows,
	})
}
