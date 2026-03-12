package services

import (
	"testing"

	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
	"gorm.io/gorm"
)

// 模拟 TestRepository 实现
type mockTestRepository struct {
	tasks   map[uint]*models.TestTask
	details map[uint]*models.TestDetail
	steps   map[uint]*models.TestStepDetail
	records map[uint]*models.TestRecord
	nextID  uint
}

func (m *mockTestRepository) CreateTestTask(task *models.TestTask) error {
	m.nextID++
	task.ID = m.nextID
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTestRepository) GetTestTaskByID(id uint) (*models.TestTask, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (m *mockTestRepository) GetTestTaskByBuildID(buildID string) (*models.TestTask, error) {
	for _, task := range m.tasks {
		if task.BuildID == buildID {
			return task, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockTestRepository) UpdateTestTask(task *models.TestTask) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTestRepository) DeleteTestTask(id uint) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTestRepository) ListTestTasks(page, pageSize int, filters map[string]interface{}, sortBy string, sortDesc bool) ([]models.TestTask, int64, error) {
	var tasks []models.TestTask
	for _, task := range m.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, int64(len(tasks)), nil
}

func (m *mockTestRepository) CreateTestDetail(detail *models.TestDetail) error {
	m.nextID++
	detail.ID = m.nextID
	m.details[detail.ID] = detail
	return nil
}

func (m *mockTestRepository) GetTestDetailByID(id uint) (*models.TestDetail, error) {
	detail, exists := m.details[id]
	if !exists {
		return nil, nil
	}
	return detail, nil
}

func (m *mockTestRepository) UpdateTestDetail(detail *models.TestDetail) error {
	m.details[detail.ID] = detail
	return nil
}

func (m *mockTestRepository) DeleteTestDetail(id uint) error {
	delete(m.details, id)
	return nil
}

func (m *mockTestRepository) ListTestDetailsByTaskID(taskID uint, page, pageSize int) ([]models.TestDetail, int64, error) {
	var details []models.TestDetail
	for _, detail := range m.details {
		if detail.TestTaskID == taskID {
			details = append(details, *detail)
		}
	}
	return details, int64(len(details)), nil
}

func (m *mockTestRepository) CreateTestRecord(record *models.TestRecord) error {
	m.nextID++
	record.ID = m.nextID
	m.records[record.ID] = record
	return nil
}

func (m *mockTestRepository) GetTestRecordByID(id uint) (*models.TestRecord, error) {
	record, exists := m.records[id]
	if !exists {
		return nil, nil
	}
	return record, nil
}

func (m *mockTestRepository) UpdateTestRecord(record *models.TestRecord) error {
	m.records[record.ID] = record
	return nil
}

func (m *mockTestRepository) DeleteTestRecord(id uint) error {
	delete(m.records, id)
	return nil
}

func (m *mockTestRepository) ListTestRecordsByTaskID(taskID uint, page, pageSize int, recordType string) ([]models.TestRecord, int64, error) {
	var records []models.TestRecord
	for _, record := range m.records {
		if record.TestTaskID == taskID && (recordType == "" || record.RecordType == recordType) {
			records = append(records, *record)
		}
	}
	return records, int64(len(records)), nil
}

func (m *mockTestRepository) CreateTestStepDetail(step *models.TestStepDetail) error {
	m.nextID++
	step.ID = m.nextID
	m.steps[step.ID] = step
	return nil
}

func (m *mockTestRepository) GetTestStepDetailByID(id uint) (*models.TestStepDetail, error) {
	step, exists := m.steps[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return step, nil
}

func (m *mockTestRepository) UpdateTestStepDetail(step *models.TestStepDetail) error {
	m.steps[step.ID] = step
	return nil
}

func (m *mockTestRepository) DeleteTestStepDetail(id uint) error {
	delete(m.steps, id)
	return nil
}

func (m *mockTestRepository) ListTestStepDetailsByTestDetailID(testDetailID uint, page, pageSize int) ([]models.TestStepDetail, int64, error) {
	var steps []models.TestStepDetail
	for _, step := range m.steps {
		if step.TestDetailID == testDetailID {
			steps = append(steps, *step)
		}
	}
	return steps, int64(len(steps)), nil
}

func newMockTestRepository() repository.TestRepository {
	return &mockTestRepository{
		tasks:   make(map[uint]*models.TestTask),
		details: make(map[uint]*models.TestDetail),
		steps:   make(map[uint]*models.TestStepDetail),
		records: make(map[uint]*models.TestRecord),
		nextID:  0,
	}
}

func TestTestService_CreateTestTask(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	task := &models.TestTask{
		TaskName:   "Test Task",
		Status:     "pending",
		WorkerName: "test-worker",
		PlanKey:    "test-plan",
	}

	err := service.CreateTestTask(task)
	if err != nil {
		t.Errorf("CreateTestTask returned error: %v", err)
	}
	if task.ID == 0 {
		t.Error("CreateTestTask did not set task ID")
	}
}

func TestTestService_GetTestTaskByID(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:   "Test Task",
		Status:     "pending",
		WorkerName: "test-worker",
		PlanKey:    "test-plan",
	}

	err := service.CreateTestTask(task)
	if err != nil {
		t.Errorf("CreateTestTask returned error: %v", err)
	}

	retrievedTask, err := service.GetTestTaskByID(task.ID)
	if err != nil {
		t.Errorf("GetTestTaskByID returned error: %v", err)
	}
	if retrievedTask == nil {
		t.Error("GetTestTaskByID returned nil task")
	}
	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, retrievedTask.ID)
	}
	if retrievedTask.TaskName != task.TaskName {
		t.Errorf("Expected task name %s, got %s", task.TaskName, retrievedTask.TaskName)
	}
}

func TestTestService_ListTestTasks(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建多个测试任务
	for i := 0; i < 3; i++ {
		task := &models.TestTask{
			TaskName:   "Test Task " + string(rune('0'+i)),
			Status:     "pending",
			WorkerName: "test-worker",
			PlanKey:    "test-plan",
		}
		err := service.CreateTestTask(task)
		if err != nil {
			t.Errorf("CreateTestTask returned error: %v", err)
		}
	}

	// 列出测试任务
	tasks, total, err := service.ListTestTasks(1, 10, nil, "id", false)
	if err != nil {
		t.Errorf("ListTestTasks returned error: %v", err)
	}
	if len(tasks) == 0 {
		t.Error("ListTestTasks returned empty list")
	}
	if total == 0 {
		t.Error("ListTestTasks returned zero total")
	}
}

func TestTestService_GetTestTaskByBuildID(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试任务
	task := &models.TestTask{
		BuildID:    "test-buildid-123",
		TaskName:   "Test Task",
		Status:     "pending",
		WorkerName: "test-worker",
		PlanKey:    "test-plan",
	}

	err := service.CreateTestTask(task)
	if err != nil {
		t.Errorf("CreateTestTask returned error: %v", err)
	}

	// 通过 BuildID 获取测试任务
	retrievedTask, err := service.GetTestTaskByBuildID("test-buildid-123")
	if err != nil {
		t.Errorf("GetTestTaskByBuildID returned error: %v", err)
	}
	if retrievedTask == nil {
		t.Error("GetTestTaskByBuildID returned nil task")
	}
	if retrievedTask.BuildID != task.BuildID {
		t.Errorf("Expected task BuildID %s, got %s", task.BuildID, retrievedTask.BuildID)
	}
	if retrievedTask.TaskName != task.TaskName {
		t.Errorf("Expected task name %s, got %s", task.TaskName, retrievedTask.TaskName)
	}
}

func TestTestService_GetTestTaskProgressByBuildID(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试任务
	task := &models.TestTask{
		BuildID:     "test-buildid-456",
		TaskName:    "Test Task with Details",
		Status:      "running",
		WorkerName:  "test-worker",
		PlanKey:     "test-plan",
		TotalTests:  5,
		PassedTests: 2,
		FailedTests: 1,
		TestDetails: []models.TestDetail{
			{TestName: "Test 1", TestStatus: "passed"},
			{TestName: "Test 2", TestStatus: "passed"},
			{TestName: "Test 3", TestStatus: "failed"},
			{TestName: "Test 4", TestStatus: "pending"},
			{TestName: "Test 5", TestStatus: "pending"},
		},
	}

	err := service.CreateTestTask(task)
	if err != nil {
		t.Errorf("CreateTestTask returned error: %v", err)
	}

	// 获取测试任务进度
	progress, err := service.GetTestTaskProgressByBuildID("test-buildid-456")
	if err != nil {
		t.Errorf("GetTestTaskProgressByBuildID returned error: %v", err)
	}
	if progress == nil {
		t.Error("GetTestTaskProgressByBuildID returned nil progress")
	}
	if progress.BuildID != task.BuildID {
		t.Errorf("Expected progress BuildID %s, got %s", task.BuildID, progress.BuildID)
	}
	if progress.TaskName != task.TaskName {
		t.Errorf("Expected progress task name %s, got %s", task.TaskName, progress.TaskName)
	}
	// 验证进度计算
	expectedProgress := 60.0 // 3 out of 5 tests completed
	if progress.Progress != expectedProgress {
		t.Errorf("Expected progress %.2f, got %.2f", expectedProgress, progress.Progress)
	}
	// 验证通过率计算
	expectedPassRate := float64(2) / float64(3) * 100 // 2 out of 3 completed tests passed
	if progress.PassRate != expectedPassRate {
		t.Errorf("Expected pass rate %.2f, got %.2f", expectedPassRate, progress.PassRate)
	}
}

func TestTestService_CreateTestStepDetail(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试步骤详情
	step := &models.TestStepDetail{
		TestDetailID: 1,
		StepName:     "Test Step",
		StartTime:    1609459200000, // 2021-01-01 00:00:00 UTC
	}

	err := service.CreateTestStepDetail(step)
	if err != nil {
		t.Errorf("CreateTestStepDetail returned error: %v", err)
	}
	if step.ID == 0 {
		t.Error("CreateTestStepDetail did not set step ID")
	}
}

func TestTestService_UpdateTestStepDetail(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试步骤详情
	step := &models.TestStepDetail{
		TestDetailID: 1,
		StepName:     "Test Step",
		StartTime:    1609459200000, // 2021-01-01 00:00:00 UTC
	}

	err := service.CreateTestStepDetail(step)
	if err != nil {
		t.Errorf("CreateTestStepDetail returned error: %v", err)
	}

	// 更新测试步骤详情
	step.EndTime = 1609459210000 // 2021-01-01 00:00:10 UTC
	step.Passed = true
	step.Screenshot = "screenshot.png"
	step.VerificationArea = "{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}"

	err = service.UpdateTestStepDetail(step)
	if err != nil {
		t.Errorf("UpdateTestStepDetail returned error: %v", err)
	}

	// 验证执行时长计算
	expectedDuration := int64(10000) // 10 seconds
	if step.Duration != expectedDuration {
		t.Errorf("Expected duration %d, got %d", expectedDuration, step.Duration)
	}
}

func TestTestService_GetTestStepDetailByID(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建测试步骤详情
	step := &models.TestStepDetail{
		TestDetailID: 1,
		StepName:     "Test Step",
		StartTime:    1609459200000, // 2021-01-01 00:00:00 UTC
	}

	err := service.CreateTestStepDetail(step)
	if err != nil {
		t.Errorf("CreateTestStepDetail returned error: %v", err)
	}

	// 获取测试步骤详情
	retrievedStep, err := service.GetTestStepDetailByID(step.ID)
	if err != nil {
		t.Errorf("GetTestStepDetailByID returned error: %v", err)
	}
	if retrievedStep == nil {
		t.Error("GetTestStepDetailByID returned nil step")
	}
	if retrievedStep.ID != step.ID {
		t.Errorf("Expected step ID %d, got %d", step.ID, retrievedStep.ID)
	}
	if retrievedStep.StepName != step.StepName {
		t.Errorf("Expected step name %s, got %s", step.StepName, retrievedStep.StepName)
	}
}

func TestTestService_ListTestStepDetailsByTestDetailID(t *testing.T) {
	// 创建模拟的 testRepo
	testRepo := newMockTestRepository()
	service := NewTestService(testRepo)

	// 创建多个测试步骤详情
	for i := 0; i < 3; i++ {
		step := &models.TestStepDetail{
			TestDetailID: 1,
			StepName:     "Test Step " + string(rune('0'+i)),
			StartTime:    1609459200000 + int64(i*1000),
		}
		err := service.CreateTestStepDetail(step)
		if err != nil {
			t.Errorf("CreateTestStepDetail returned error: %v", err)
		}
	}

	// 列出测试步骤详情
	steps, total, err := service.ListTestStepDetailsByTestDetailID(1, 1, 10)
	if err != nil {
		t.Errorf("ListTestStepDetailsByTestDetailID returned error: %v", err)
	}
	if len(steps) == 0 {
		t.Error("ListTestStepDetailsByTestDetailID returned empty list")
	}
	if total == 0 {
		t.Error("ListTestStepDetailsByTestDetailID returned zero total")
	}
}
