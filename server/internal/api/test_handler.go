package api

import (
	"strconv"

	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/services"
	"github.com/gin-gonic/gin"
)

// TestHandler 测试数据处理程序
type TestHandler struct {
	service services.TestService
}

// NewTestHandler 创建测试数据处理程序实例
func NewTestHandler(service services.TestService) *TestHandler {
	return &TestHandler{service: service}
}

// 测试任务相关路由处理

// CreateTestTask 创建测试任务
// @Summary 创建测试任务
// @Description 创建新的测试任务执行结果，首次创建需提供 buildID, taskname, status, starttime, totaltest, worker_name, plan_key 字段
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param task body models.TestTask true "测试任务信息"
// @Success 201 {object} models.TestTask
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks [post]
func (h *TestHandler) CreateTestTask(c *gin.Context) {
	var task models.TestTask
	if err := c.ShouldBindJSON(&task); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 验证必要字段
	if task.BuildID == "" || task.TaskName == "" || task.Status == "" || task.StartTime == 0 || task.TotalTests == 0 || task.WorkerName == "" || task.PlanKey == "" {
		BadRequest(c, "Missing required fields: buildID, taskname, status, starttime, totaltest, worker_name, plan_key")
		return
	}

	// 自动填充默认值
	if task.EndTime == 0 {
		task.EndTime = 0
	}
	if task.Duration == 0 {
		task.Duration = 0
	}
	if task.PassedTests == 0 {
		task.PassedTests = 0
	}
	if task.FailedTests == 0 {
		task.FailedTests = 0
	}

	if err := h.service.CreateTestTask(&task); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Created(c, task)
}

// GetTestTaskByID 获取测试任务详情
// @Summary 获取测试任务详情
// @Description 根据 ID 获取测试任务执行结果详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试任务 ID"
// @Success 200 {object} models.TestTask
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/{id} [get]
func (h *TestHandler) GetTestTaskByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid task ID")
		return
	}

	task, err := h.service.GetTestTaskByID(uint(id))
	if err != nil {
		NotFound(c, "Task not found")
		return
	}

	Success(c, task)
}

// UpdateTestTask 更新测试任务
// @Summary 更新测试任务
// @Description 更新测试任务执行结果，仅允许更新 endtime, passedtests, failedtest 字段
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param id path int true "测试任务 ID"
// @Param task body models.TestTask true "测试任务信息"
// @Success 200 {object} models.TestTask
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/{id} [put]
func (h *TestHandler) UpdateTestTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid task ID")
		return
	}

	// 获取现有测试任务
	existingTask, err := h.service.GetTestTaskByID(uint(id))
	if err != nil {
		NotFound(c, "Task not found")
		return
	}

	// 仅接受需要更新的字段
	var updateData struct {
		EndTime     int64 `json:"end_time"`
		PassedTests int   `json:"passed_tests"`
		FailedTests int   `json:"failed_tests"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 更新字段
	existingTask.EndTime = updateData.EndTime
	existingTask.PassedTests = updateData.PassedTests
	existingTask.FailedTests = updateData.FailedTests

	// 自动计算 duration
	if existingTask.EndTime > existingTask.StartTime {
		existingTask.Duration = existingTask.EndTime - existingTask.StartTime
	} else {
		existingTask.Duration = 0
	}

	if err := h.service.UpdateTestTask(existingTask); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, existingTask)
}

// DeleteTestTask 删除测试任务
// @Summary 删除测试任务
// @Description 删除测试任务执行结果
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试任务 ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/{id} [delete]
func (h *TestHandler) DeleteTestTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid task ID")
		return
	}

	if err := h.service.DeleteTestTask(uint(id)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Task deleted successfully"})
}

// ListTestTasks 列出测试任务
// @Summary 列出测试任务
// @Description 分页列出测试任务执行结果，支持筛选和排序
// @Tags 测试管理
// @Produce json
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页大小，默认 10"
// @Param status query string false "状态筛选"
// @Param sort_by query string false "排序字段"
// @Param sort_desc query bool false "是否倒序排序"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks [get]
func (h *TestHandler) ListTestTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	sortBy := c.DefaultQuery("sort_by", "")
	sortDesc, _ := strconv.ParseBool(c.DefaultQuery("sort_desc", "true"))

	// 构建筛选条件
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	tasks, total, err := h.service.ListTestTasks(page, pageSize, filters, sortBy, sortDesc)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"tasks":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 测试详情相关路由处理

// CreateTestDetail 创建测试详情
// @Summary 创建测试详情
// @Description 创建新的测试结果详情，支持通过 buildid 关联到 testtask
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param detail body models.TestDetail true "测试详情信息"
// @Success 201 {object} models.TestDetail
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-details [post]
func (h *TestHandler) CreateTestDetail(c *gin.Context) {
	// 接受包含 buildid 的请求
	var request struct {
		TestName      string `json:"test_name"`       // case key
		TestStartTime int64  `json:"test_start_time"` // 开始时间
		BuildID       string `json:"build_id"`        // 用于关联到 testtask
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 验证必要字段
	if request.TestName == "" || request.TestStartTime == 0 || request.BuildID == "" {
		BadRequest(c, "Missing required fields: test_name, test_start_time, build_id")
		return
	}

	// 通过 buildid 获取 testtask
	task, err := h.service.GetTestTaskByBuildID(request.BuildID)
	if err != nil {
		NotFound(c, "Test task not found for the provided build_id")
		return
	}

	// 创建测试详情
	detail := &models.TestDetail{
		TestTaskID:    task.ID,
		TestName:      request.TestName,
		TestStatus:    "running", // 初始状态为 running
		TestStartTime: request.TestStartTime,
		TestEndTime:   0,
		Duration:      0,
	}

	if err := h.service.CreateTestDetail(detail); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Created(c, detail)
}

// GetTestDetailByID 获取测试详情
// @Summary 获取测试详情
// @Description 根据 ID 获取测试结果详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试详情 ID"
// @Success 200 {object} models.TestDetail
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-details/{id} [get]
func (h *TestHandler) GetTestDetailByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid detail ID")
		return
	}

	detail, err := h.service.GetTestDetailByID(uint(id))
	if err != nil {
		NotFound(c, "Detail not found")
		return
	}

	Success(c, detail)
}

// UpdateTestDetail 更新测试详情
// @Summary 更新测试详情
// @Description 更新测试结果详情，支持更新 endtime 和 pass/failed 状态
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param id path int true "测试详情 ID"
// @Param detail body models.TestDetail true "测试详情信息"
// @Success 200 {object} models.TestDetail
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-details/{id} [put]
func (h *TestHandler) UpdateTestDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid detail ID")
		return
	}

	// 获取现有测试详情
	existingDetail, err := h.service.GetTestDetailByID(uint(id))
	if err != nil {
		NotFound(c, "Detail not found")
		return
	}

	// 仅接受需要更新的字段
	var updateData struct {
		TestEndTime  int64  `json:"test_end_time"`
		TestStatus   string `json:"test_status"` // passed, failed
		ErrorMessage string `json:"error_message"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 更新字段
	existingDetail.TestEndTime = updateData.TestEndTime
	existingDetail.TestStatus = updateData.TestStatus
	existingDetail.ErrorMessage = updateData.ErrorMessage

	// 自动计算执行时长
	if existingDetail.TestEndTime > existingDetail.TestStartTime {
		existingDetail.Duration = existingDetail.TestEndTime - existingDetail.TestStartTime
	} else {
		existingDetail.Duration = 0
	}

	if err := h.service.UpdateTestDetail(existingDetail); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, existingDetail)
}

// DeleteTestDetail 删除测试详情
// @Summary 删除测试详情
// @Description 删除测试结果详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试详情 ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-details/{id} [delete]
func (h *TestHandler) DeleteTestDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid detail ID")
		return
	}

	if err := h.service.DeleteTestDetail(uint(id)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Detail deleted successfully"})
}

// ListTestDetailsByTaskID 列出测试详情
// @Summary 列出测试详情
// @Description 分页列出指定测试任务的测试结果详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试任务 ID"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页大小，默认 10"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/{id}/details [get]
func (h *TestHandler) ListTestDetailsByTaskID(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid task ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	details, total, err := h.service.ListTestDetailsByTaskID(uint(taskID), page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"details":   details,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 测试记录相关路由处理

// CreateTestRecord 创建测试记录
// @Summary 创建测试记录
// @Description 创建新的测试记录数据
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param record body models.TestRecord true "测试记录信息"
// @Success 201 {object} models.TestRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-records [post]
func (h *TestHandler) CreateTestRecord(c *gin.Context) {
	var record models.TestRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.service.CreateTestRecord(&record); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Created(c, record)
}

// GetTestRecordByID 获取测试记录
// @Summary 获取测试记录
// @Description 根据 ID 获取测试记录数据
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试记录 ID"
// @Success 200 {object} models.TestRecord
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-records/{id} [get]
func (h *TestHandler) GetTestRecordByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid record ID")
		return
	}

	record, err := h.service.GetTestRecordByID(uint(id))
	if err != nil {
		NotFound(c, "Record not found")
		return
	}

	Success(c, record)
}

// UpdateTestRecord 更新测试记录
// @Summary 更新测试记录
// @Description 更新测试记录数据
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param id path int true "测试记录 ID"
// @Param record body models.TestRecord true "测试记录信息"
// @Success 200 {object} models.TestRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-records/{id} [put]
func (h *TestHandler) UpdateTestRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid record ID")
		return
	}

	var record models.TestRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		BadRequest(c, err.Error())
		return
	}

	record.ID = uint(id)
	if err := h.service.UpdateTestRecord(&record); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, record)
}

// DeleteTestRecord 删除测试记录
// @Summary 删除测试记录
// @Description 删除测试记录数据
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试记录 ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-records/{id} [delete]
func (h *TestHandler) DeleteTestRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid record ID")
		return
	}

	if err := h.service.DeleteTestRecord(uint(id)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Record deleted successfully"})
}

// ListTestRecordsByTaskID 列出测试记录
// @Summary 列出测试记录
// @Description 分页列出指定测试任务的测试记录数据，支持记录类型筛选
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试任务 ID"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页大小，默认 10"
// @Param record_type query string false "记录类型筛选"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/{id}/records [get]
func (h *TestHandler) ListTestRecordsByTaskID(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid task ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	recordType := c.DefaultQuery("record_type", "")

	records, total, err := h.service.ListTestRecordsByTaskID(uint(taskID), page, pageSize, recordType)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTestTaskProgressByBuildID 获取测试任务进度
// @Summary 获取测试任务进度
// @Description 根据 BuildID 获取测试任务的运行进度和通过率
// @Tags 测试管理
// @Produce json
// @Param buildid path string true "测试任务 BuildID"
// @Success 200 {object} services.TestTaskProgress
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-tasks/buildid/{buildid}/progress [get]
func (h *TestHandler) GetTestTaskProgressByBuildID(c *gin.Context) {
	buildID := c.Param("buildid")
	if buildID == "" {
		BadRequest(c, "Invalid BuildID")
		return
	}

	progress, err := h.service.GetTestTaskProgressByBuildID(buildID)
	if err != nil {
		NotFound(c, "Test task not found")
		return
	}

	Success(c, progress)
}

// 测试步骤详情相关路由处理

// CreateTestStepDetail 创建测试步骤详情
// @Summary 创建测试步骤详情
// @Description 创建新的测试步骤详情，关联到测试详情
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param step body models.TestStepDetail true "测试步骤详情信息"
// @Success 201 {object} models.TestStepDetail
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-step-details [post]
func (h *TestHandler) CreateTestStepDetail(c *gin.Context) {
	var step models.TestStepDetail
	if err := c.ShouldBindJSON(&step); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 验证必要字段
	if step.StepName == "" || step.StartTime == 0 || step.TestDetailID == 0 {
		BadRequest(c, "Missing required fields: step_name, start_time, test_detail_id")
		return
	}

	// 自动填充默认值
	if step.EndTime == 0 {
		step.EndTime = 0
	}
	if step.Duration == 0 {
		step.Duration = 0
	}

	if err := h.service.CreateTestStepDetail(&step); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Created(c, step)
}

// GetTestStepDetailByID 获取测试步骤详情
// @Summary 获取测试步骤详情
// @Description 根据 ID 获取测试步骤详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试步骤详情 ID"
// @Success 200 {object} models.TestStepDetail
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-step-details/{id} [get]
func (h *TestHandler) GetTestStepDetailByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid step ID")
		return
	}

	step, err := h.service.GetTestStepDetailByID(uint(id))
	if err != nil {
		NotFound(c, "Step not found")
		return
	}

	Success(c, step)
}

// UpdateTestStepDetail 更新测试步骤详情
// @Summary 更新测试步骤详情
// @Description 更新测试步骤详情，包括结束时间、是否通过、截图、验证区域等
// @Tags 测试管理
// @Accept json
// @Produce json
// @Param id path int true "测试步骤详情 ID"
// @Param step body models.TestStepDetail true "测试步骤详情信息"
// @Success 200 {object} models.TestStepDetail
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-step-details/{id} [put]
func (h *TestHandler) UpdateTestStepDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid step ID")
		return
	}

	// 获取现有测试步骤详情
	existingStep, err := h.service.GetTestStepDetailByID(uint(id))
	if err != nil {
		NotFound(c, "Step not found")
		return
	}

	// 仅接受需要更新的字段
	var updateData struct {
		EndTime          int64  `json:"end_time"`
		Passed           bool   `json:"passed"`
		Screenshot       string `json:"screenshot"`
		VerificationArea string `json:"verification_area"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 更新字段
	existingStep.EndTime = updateData.EndTime
	existingStep.Passed = updateData.Passed
	existingStep.Screenshot = updateData.Screenshot
	existingStep.VerificationArea = updateData.VerificationArea

	if err := h.service.UpdateTestStepDetail(existingStep); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, existingStep)
}

// DeleteTestStepDetail 删除测试步骤详情
// @Summary 删除测试步骤详情
// @Description 删除测试步骤详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试步骤详情 ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-step-details/{id} [delete]
func (h *TestHandler) DeleteTestStepDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid step ID")
		return
	}

	if err := h.service.DeleteTestStepDetail(uint(id)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Step deleted successfully"})
}

// ListTestStepDetailsByTestDetailID 列出测试步骤详情
// @Summary 列出测试步骤详情
// @Description 分页列出指定测试详情的测试步骤详情
// @Tags 测试管理
// @Produce json
// @Param id path int true "测试详情 ID"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页大小，默认 10"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/test-details/{id}/steps [get]
func (h *TestHandler) ListTestStepDetailsByTestDetailID(c *gin.Context) {
	testDetailID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test detail ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	steps, total, err := h.service.ListTestStepDetailsByTestDetailID(uint(testDetailID), page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"steps":     steps,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
