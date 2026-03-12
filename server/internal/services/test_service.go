package services

import (
	"fmt"
	"time"

	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
)

// TestTaskProgress 测试任务进度信息
type TestTaskProgress struct {
	BuildID        string  `json:"build_id"`
	TaskName       string  `json:"task_name"`
	Status         string  `json:"status"`
	Progress       float64 `json:"progress"`  // 进度百分比
	PassRate       float64 `json:"pass_rate"` // 通过率百分比
	TotalTests     int     `json:"total_tests"`
	PassedTests    int     `json:"passed_tests"`
	FailedTests    int     `json:"failed_tests"`
	CompletedTests int     `json:"completed_tests"`
}

// TestService 测试数据服务接口
type TestService interface {
	// 测试任务相关操作
	CreateTestTask(task *models.TestTask) error
	GetTestTaskByID(id uint) (*models.TestTask, error)
	GetTestTaskByBuildID(buildID string) (*models.TestTask, error)
	GetTestTaskProgressByBuildID(buildID string) (*TestTaskProgress, error)
	UpdateTestTask(task *models.TestTask) error
	DeleteTestTask(id uint) error
	ListTestTasks(page, pageSize int, filters map[string]interface{}, sortBy string, sortDesc bool) ([]models.TestTask, int64, error)

	// 测试详情相关操作
	CreateTestDetail(detail *models.TestDetail) error
	GetTestDetailByID(id uint) (*models.TestDetail, error)
	UpdateTestDetail(detail *models.TestDetail) error
	DeleteTestDetail(id uint) error
	ListTestDetailsByTaskID(taskID uint, page, pageSize int) ([]models.TestDetail, int64, error)

	// 测试记录相关操作
	CreateTestRecord(record *models.TestRecord) error
	GetTestRecordByID(id uint) (*models.TestRecord, error)
	UpdateTestRecord(record *models.TestRecord) error
	DeleteTestRecord(id uint) error
	ListTestRecordsByTaskID(taskID uint, page, pageSize int, recordType string) ([]models.TestRecord, int64, error)

	// 测试步骤详情相关操作
	CreateTestStepDetail(step *models.TestStepDetail) error
	GetTestStepDetailByID(id uint) (*models.TestStepDetail, error)
	UpdateTestStepDetail(step *models.TestStepDetail) error
	DeleteTestStepDetail(id uint) error
	ListTestStepDetailsByTestDetailID(testDetailID uint, page, pageSize int) ([]models.TestStepDetail, int64, error)
}

// testService 测试数据服务实现
type testService struct {
	repo  repository.TestRepository
	cache CacheService
}

// NewTestService 创建测试数据服务实例
func NewTestService(repo repository.TestRepository) TestService {
	return &testService{
		repo:  repo,
		cache: NewCacheService(),
	}
}

// 测试任务相关操作
func (s *testService) CreateTestTask(task *models.TestTask) error {
	return s.repo.CreateTestTask(task)
}

func (s *testService) GetTestTaskByID(id uint) (*models.TestTask, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("test_task:%d", id)
	var task models.TestTask
	err := s.cache.Get(cacheKey, &task)
	if err == nil && task.ID != 0 {
		return &task, nil
	}

	// 缓存未命中，从数据库获取
	taskPtr, err := s.repo.GetTestTaskByID(id)
	if err != nil {
		return nil, err
	}

	// 设置缓存，过期时间 5 分钟
	s.cache.Set(cacheKey, taskPtr, 5*time.Minute)

	return taskPtr, nil
}

func (s *testService) GetTestTaskByBuildID(buildID string) (*models.TestTask, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("test_task:buildid:%s", buildID)
	var task models.TestTask
	err := s.cache.Get(cacheKey, &task)
	if err == nil && task.ID != 0 {
		return &task, nil
	}

	// 缓存未命中，从数据库获取
	taskPtr, err := s.repo.GetTestTaskByBuildID(buildID)
	if err != nil {
		return nil, err
	}

	// 设置缓存，过期时间 5 分钟
	s.cache.Set(cacheKey, taskPtr, 5*time.Minute)

	return taskPtr, nil
}

func (s *testService) GetTestTaskProgressByBuildID(buildID string) (*TestTaskProgress, error) {
	// 获取测试任务
	task, err := s.GetTestTaskByBuildID(buildID)
	if err != nil {
		return nil, err
	}

	// 计算已完成的测试数
	completedTests := 0
	passedTests := 0
	for _, detail := range task.TestDetails {
		if detail.TestStatus == "passed" || detail.TestStatus == "failed" {
			completedTests++
			if detail.TestStatus == "passed" {
				passedTests++
			}
		}
	}

	// 计算进度和通过率
	var progress float64 = 0
	var passRate float64 = 0

	if task.TotalTests > 0 {
		progress = float64(completedTests) / float64(task.TotalTests) * 100
		if completedTests > 0 {
			passRate = float64(passedTests) / float64(completedTests) * 100
		}
	}

	// 构建进度信息
	progressInfo := &TestTaskProgress{
		BuildID:        task.BuildID,
		TaskName:       task.TaskName,
		Status:         task.Status,
		Progress:       progress,
		PassRate:       passRate,
		TotalTests:     task.TotalTests,
		PassedTests:    task.PassedTests,
		FailedTests:    task.FailedTests,
		CompletedTests: completedTests,
	}

	return progressInfo, nil
}

func (s *testService) UpdateTestTask(task *models.TestTask) error {
	err := s.repo.UpdateTestTask(task)
	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("test_task:%d", task.ID)
	s.cache.Delete(cacheKey)

	// 清除基于 BuildID 的缓存
	if task.BuildID != "" {
		buildIDCacheKey := fmt.Sprintf("test_task:buildid:%s", task.BuildID)
		s.cache.Delete(buildIDCacheKey)
	}

	return nil
}

func (s *testService) DeleteTestTask(id uint) error {
	// 先获取测试任务，以便清除基于 BuildID 的缓存
	task, err := s.repo.GetTestTaskByID(id)
	if err == nil && task.BuildID != "" {
		// 清除基于 BuildID 的缓存
		buildIDCacheKey := fmt.Sprintf("test_task:buildid:%s", task.BuildID)
		s.cache.Delete(buildIDCacheKey)
	}

	err = s.repo.DeleteTestTask(id)
	if err != nil {
		return err
	}

	// 清除基于 ID 的缓存
	cacheKey := fmt.Sprintf("test_task:%d", id)
	s.cache.Delete(cacheKey)

	return nil
}

func (s *testService) ListTestTasks(page, pageSize int, filters map[string]interface{}, sortBy string, sortDesc bool) ([]models.TestTask, int64, error) {
	return s.repo.ListTestTasks(page, pageSize, filters, sortBy, sortDesc)
}

// 测试详情相关操作
func (s *testService) CreateTestDetail(detail *models.TestDetail) error {
	return s.repo.CreateTestDetail(detail)
}

func (s *testService) GetTestDetailByID(id uint) (*models.TestDetail, error) {
	return s.repo.GetTestDetailByID(id)
}

func (s *testService) UpdateTestDetail(detail *models.TestDetail) error {
	return s.repo.UpdateTestDetail(detail)
}

func (s *testService) DeleteTestDetail(id uint) error {
	return s.repo.DeleteTestDetail(id)
}

func (s *testService) ListTestDetailsByTaskID(taskID uint, page, pageSize int) ([]models.TestDetail, int64, error) {
	return s.repo.ListTestDetailsByTaskID(taskID, page, pageSize)
}

// 测试记录相关操作
func (s *testService) CreateTestRecord(record *models.TestRecord) error {
	return s.repo.CreateTestRecord(record)
}

func (s *testService) GetTestRecordByID(id uint) (*models.TestRecord, error) {
	return s.repo.GetTestRecordByID(id)
}

func (s *testService) UpdateTestRecord(record *models.TestRecord) error {
	return s.repo.UpdateTestRecord(record)
}

func (s *testService) DeleteTestRecord(id uint) error {
	return s.repo.DeleteTestRecord(id)
}

func (s *testService) ListTestRecordsByTaskID(taskID uint, page, pageSize int, recordType string) ([]models.TestRecord, int64, error) {
	return s.repo.ListTestRecordsByTaskID(taskID, page, pageSize, recordType)
}

// 测试步骤详情相关操作
func (s *testService) CreateTestStepDetail(step *models.TestStepDetail) error {
	return s.repo.CreateTestStepDetail(step)
}

func (s *testService) GetTestStepDetailByID(id uint) (*models.TestStepDetail, error) {
	return s.repo.GetTestStepDetailByID(id)
}

func (s *testService) UpdateTestStepDetail(step *models.TestStepDetail) error {
	// 自动计算执行时长
	if step.EndTime > step.StartTime {
		step.Duration = step.EndTime - step.StartTime
	} else {
		step.Duration = 0
	}
	return s.repo.UpdateTestStepDetail(step)
}

func (s *testService) DeleteTestStepDetail(id uint) error {
	return s.repo.DeleteTestStepDetail(id)
}

func (s *testService) ListTestStepDetailsByTestDetailID(testDetailID uint, page, pageSize int) ([]models.TestStepDetail, int64, error) {
	return s.repo.ListTestStepDetailsByTestDetailID(testDetailID, page, pageSize)
}
