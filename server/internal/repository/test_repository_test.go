package repository

import (
	"testing"

	"com.hermes.platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTestRepository_CreateTestTask(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}

	// 测试创建测试任务
	err = repo.CreateTestTask(task)
	if err != nil {
		t.Errorf("Failed to create test task: %v", err)
	}

	// 验证测试任务是否创建成功
	var createdTask models.TestTask
	db.First(&createdTask, task.ID)
	if createdTask.TaskName != task.TaskName {
		t.Errorf("Expected task name %s, got %s", task.TaskName, createdTask.TaskName)
	}
}

func TestTestRepository_GetTestTaskByID(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}
	db.Create(task)

	// 测试根据 ID 查找测试任务
	foundTask, err := repo.GetTestTaskByID(task.ID)
	if err != nil {
		t.Errorf("Failed to find test task by ID: %v", err)
	}

	if foundTask == nil {
		t.Error("Expected test task, got nil")
	}

	if foundTask.ID != task.ID {
		t.Errorf("Expected ID %d, got %d", task.ID, foundTask.ID)
	}

	// 测试查找不存在的测试任务
	_, err = repo.GetTestTaskByID(999)
	if err == nil {
		t.Error("Expected error when finding nonexistent test task, got nil")
	}
}

func TestTestRepository_UpdateTestTask(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}
	db.Create(task)

	// 更新测试任务
	task.TaskName = "Updated Test Task"
	task.Status = "in_progress"
	err = repo.UpdateTestTask(task)
	if err != nil {
		t.Errorf("Failed to update test task: %v", err)
	}

	// 验证测试任务是否更新成功
	var updatedTask models.TestTask
	db.First(&updatedTask, task.ID)
	if updatedTask.TaskName != "Updated Test Task" {
		t.Errorf("Expected task name Updated Test Task, got %s", updatedTask.TaskName)
	}
	if updatedTask.Status != "in_progress" {
		t.Errorf("Expected status in_progress, got %s", updatedTask.Status)
	}
}

func TestTestRepository_DeleteTestTask(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}
	db.Create(task)

	// 测试删除测试任务
	err = repo.DeleteTestTask(task.ID)
	if err != nil {
		t.Errorf("Failed to delete test task: %v", err)
	}

	// 验证测试任务是否删除成功
	var deletedTask models.TestTask
	result := db.First(&deletedTask, task.ID)
	if result.Error == nil {
		t.Error("Expected error when finding deleted test task, got nil")
	}
}

func TestTestRepository_ListTestTasks(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建多个测试任务
	for i := 1; i <= 5; i++ {
		task := &models.TestTask{
			TaskName:    "Test Task " + string(rune('0'+i)),
			Status:      "pending",
			TotalTests:  0,
			PassedTests: 0,
			FailedTests: 0,
		}
		db.Create(task)
	}

	// 测试列出测试任务
	tasks, total, err := repo.ListTestTasks(1, 10, nil, "", false)
	if err != nil {
		t.Errorf("Failed to list test tasks: %v", err)
	}

	if len(tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(tasks))
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
}

func TestTestRepository_CreateTestDetail(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}
	db.Create(task)

	// 创建测试详情
	detail := &models.TestDetail{
		TestTaskID: task.ID,
		TestName:   "Test Detail",
		TestStatus: "passed",
		Duration:   0,
	}

	// 测试创建测试详情
	err = repo.CreateTestDetail(detail)
	if err != nil {
		t.Errorf("Failed to create test detail: %v", err)
	}

	// 验证测试详情是否创建成功
	var createdDetail models.TestDetail
	db.First(&createdDetail, detail.ID)
	if createdDetail.TestName != detail.TestName {
		t.Errorf("Expected test name %s, got %s", detail.TestName, createdDetail.TestName)
	}
}

func TestTestRepository_CreateTestRecord(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.TestTask{}, &models.TestDetail{}, &models.TestRecord{})

	// 创建测试仓库
	repo := NewTestRepository(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}
	db.Create(task)

	// 创建测试记录
	record := &models.TestRecord{
		TestTaskID: task.ID,
		TestName:   "Test Record",
		RecordType: "info",
		RecordFile: "Test Record Content",
		RecordTime: 0,
	}

	// 测试创建测试记录
	err = repo.CreateTestRecord(record)
	if err != nil {
		t.Errorf("Failed to create test record: %v", err)
	}

	// 验证测试记录是否创建成功
	var createdRecord models.TestRecord
	db.First(&createdRecord, record.ID)
	if createdRecord.RecordFile != record.RecordFile {
		t.Errorf("Expected record file %s, got %s", record.RecordFile, createdRecord.RecordFile)
	}
}
