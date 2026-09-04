// Package dashboard 承载 dashboard 上下文的 HTTP 处理器。
package dashboard

import (
	"strconv"

	"github.com/gin-gonic/gin"

	dtoautomation "github.com/hermes-platform/go-service/internal/adapter/http/dto/dashboard"
	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// Handlers 聚合 dashboard 的处理器（2 个接口）。
type Handlers struct {
	overview     *appdashboard.OverviewUseCase
	runningCases *appdashboard.RunningCasesUseCase
}

// NewHandlers 构造 dashboard 处理器。
func NewHandlers(
	overview *appdashboard.OverviewUseCase,
	runningCases *appdashboard.RunningCasesUseCase,
) *Handlers {
	return &Handlers{overview: overview, runningCases: runningCases}
}

// GetOverview 处理 GET /automation/dashboard/overview。
func (h *Handlers) GetOverview(c *gin.Context) {
	result, err := h.overview.GetOverview(c.Request.Context(), true)
	if err != nil {
		response.Error(c, err)
		return
	}

	trends := make([]dtoautomation.TrendItem, 0, len(result.Trends))
	for _, t := range result.Trends {
		trends = append(trends, dtoautomation.ToTrendItem(t))
	}

	running := make([]dtoautomation.RunningExecutionItem, 0, len(result.RunningExecutions))
	for _, e := range result.RunningExecutions {
		running = append(running, dtoautomation.ToRunningExecutionItem(e))
	}

	response.OKWithMessage(c, "获取 dashboard 概览成功", dtoautomation.OverviewData{
		Summary:           dtoautomation.ToSummaryItem(result.Summary),
		Trends:            trends,
		RunningExecutions: running,
	})
}

// GetRunningCases 处理 GET /automation/dashboard/executions/{execution_id}/cases。
func (h *Handlers) GetRunningCases(c *gin.Context) {
	executionID, err := strconv.ParseInt(c.Param("execution_id"), 10, 64)
	if err != nil || executionID <= 0 {
		response.Fail(c, 422, "路径参数不合法")
		return
	}

	result, err := h.runningCases.GetCases(c.Request.Context(), executionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dtoautomation.RunningCaseItem, 0, len(result.Items))
	for _, ci := range result.Items {
		items = append(items, dtoautomation.ToRunningCaseItem(ci))
	}

	response.OKWithMessage(c, "获取运行中 execution 用例成功", dtoautomation.RunningCaseListData{
		ExecutionID: result.ExecutionID,
		Items:       items,
	})
}
