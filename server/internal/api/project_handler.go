package api

import (
	"strconv"

	"com.hermes.platform/internal/services"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := c.GetUint("user_id")
	project, err := h.projectService.CreateProject(req.Name, req.Description, userID)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid project ID")
		return
	}

	project, err := h.projectService.GetProject(uint(id))
	if err != nil {
		NotFound(c, "Project not found")
		return
	}

	Success(c, gin.H{"project": project})
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	projects, total, err := h.projectService.ListProjects(page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"projects":  projects,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid project ID")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	err = h.projectService.UpdateProject(uint(id), req.Name, req.Description)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Project updated successfully"})
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid project ID")
		return
	}

	err = h.projectService.DeleteProject(uint(id))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Project deleted successfully"})
}

func (h *ProjectHandler) CreateVersion(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid project ID")
		return
	}

	var req struct {
		Version     string `json:"version" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	version, err := h.projectService.CreateVersion(uint(projectID), req.Version, req.Description)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"version": version})
}

func (h *ProjectHandler) GetVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid version ID")
		return
	}

	version, err := h.projectService.GetVersion(uint(id))
	if err != nil {
		NotFound(c, "Version not found")
		return
	}

	Success(c, gin.H{"version": version})
}

func (h *ProjectHandler) ListVersions(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid project ID")
		return
	}

	versions, err := h.projectService.ListVersions(uint(projectID))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"versions": versions})
}

func (h *ProjectHandler) UpdateVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid version ID")
		return
	}

	var req struct {
		Version     string `json:"version" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	err = h.projectService.UpdateVersion(uint(id), req.Version, req.Description, status)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Version updated successfully"})
}

func (h *ProjectHandler) DeleteVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid version ID")
		return
	}

	err = h.projectService.DeleteVersion(uint(id))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Version deleted successfully"})
}

func (h *ProjectHandler) CreateTestPlan(c *gin.Context) {
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid version ID")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := c.GetUint("user_id")
	plan, err := h.projectService.CreateTestPlan(uint(versionID), req.Name, req.Description, userID)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"test_plan": plan})
}

func (h *ProjectHandler) GetTestPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test plan ID")
		return
	}

	plan, err := h.projectService.GetTestPlan(uint(id))
	if err != nil {
		NotFound(c, "Test plan not found")
		return
	}

	Success(c, gin.H{"test_plan": plan})
}

func (h *ProjectHandler) ListTestPlans(c *gin.Context) {
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid version ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	plans, total, err := h.projectService.ListTestPlans(uint(versionID), page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"test_plans": plans,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func (h *ProjectHandler) UpdateTestPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test plan ID")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	err = h.projectService.UpdateTestPlan(uint(id), req.Name, req.Description, status)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Test plan updated successfully"})
}

func (h *ProjectHandler) DeleteTestPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test plan ID")
		return
	}

	err = h.projectService.DeleteTestPlan(uint(id))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Test plan deleted successfully"})
}

func (h *ProjectHandler) CreateTestCase(c *gin.Context) {
	planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test plan ID")
		return
	}

	var req struct {
		Name             string `json:"name" binding:"required"`
		Status           string `json:"status"`
		Remark           string `json:"remark"`
		PreCondition     string `json:"pre_condition"`
		StepDescription  string `json:"step_description"`
		ExpectedResult   string `json:"expected_result"`
		Priority         int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	priority := req.Priority
	if priority == 0 {
		priority = 3
	}

	userID := c.GetUint("user_id")
	testCase, err := h.projectService.CreateTestCase(
		uint(planID),
		req.Name,
		req.Status,
		req.Remark,
		req.PreCondition,
		req.StepDescription,
		req.ExpectedResult,
		priority,
		userID,
	)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"test_case": testCase})
}

func (h *ProjectHandler) GetTestCase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test case ID")
		return
	}

	testCase, err := h.projectService.GetTestCase(uint(id))
	if err != nil {
		NotFound(c, "Test case not found")
		return
	}

	Success(c, gin.H{"test_case": testCase})
}

func (h *ProjectHandler) GetTestCaseByCaseKey(c *gin.Context) {
	caseKey := c.Param("caseKey")

	testCase, err := h.projectService.GetTestCaseByCaseKey(caseKey)
	if err != nil {
		NotFound(c, "Test case not found")
		return
	}

	Success(c, gin.H{"test_case": testCase})
}

func (h *ProjectHandler) ListTestCases(c *gin.Context) {
	planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test plan ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	testCases, total, err := h.projectService.ListTestCases(uint(planID), page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"test_cases": testCases,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func (h *ProjectHandler) UpdateTestCase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test case ID")
		return
	}

	var req struct {
		Name             string `json:"name" binding:"required"`
		Status           string `json:"status"`
		Remark           string `json:"remark"`
		PreCondition     string `json:"pre_condition"`
		StepDescription  string `json:"step_description"`
		ExpectedResult   string `json:"expected_result"`
		Priority         int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	priority := req.Priority
	if priority == 0 {
		priority = 3
	}

	err = h.projectService.UpdateTestCase(
		uint(id),
		req.Name,
		req.Status,
		req.Remark,
		req.PreCondition,
		req.StepDescription,
		req.ExpectedResult,
		priority,
	)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Test case updated successfully"})
}

func (h *ProjectHandler) DeleteTestCase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid test case ID")
		return
	}

	err = h.projectService.DeleteTestCase(uint(id))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Test case deleted successfully"})
}

type TestCaseStatus string

const (
	TestCaseStatusDeveloping TestCaseStatus = "developing"
	TestCaseStatusReady      TestCaseStatus = "ready"
	TestCaseStatusSkip       TestCaseStatus = "skip"
)

func (h *ProjectHandler) GetTestCaseStatuses(c *gin.Context) {
	statuses := []gin.H{
		{"value": "developing", "label": "开发中", "label_en": "Developing"},
		{"value": "ready", "label": "就绪", "label_en": "Ready"},
		{"value": "skip", "label": "跳过", "label_en": "Skip"},
	}
	Success(c, gin.H{"statuses": statuses})
}

func (h *ProjectHandler) GetPriorities(c *gin.Context) {
	priorities := []gin.H{
		{"value": 1, "label": "高", "label_en": "High"},
		{"value": 2, "label": "中", "label_en": "Medium"},
		{"value": 3, "label": "低", "label_en": "Low"},
	}
	Success(c, gin.H{"priorities": priorities})
}

func (h *ProjectHandler) GetTestPlanStatuses(c *gin.Context) {
	statuses := []gin.H{
		{"value": "draft", "label": "草稿", "label_en": "Draft"},
		{"value": "active", "label": "进行中", "label_en": "Active"},
		{"value": "completed", "label": "已完成", "label_en": "Completed"},
	}
	Success(c, gin.H{"statuses": statuses})
}

func (h *ProjectHandler) GetVersionStatuses(c *gin.Context) {
	statuses := []gin.H{
		{"value": "active", "label": "激活", "label_en": "Active"},
		{"value": "archived", "label": "归档", "label_en": "Archived"},
	}
	Success(c, gin.H{"statuses": statuses})
}
