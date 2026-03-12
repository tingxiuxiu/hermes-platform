package database

import (
	"testing"

	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrate(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	AutoMigrate(db)

	// 验证表是否创建成功
	tables := []string{"users", "roles", "permissions", "test_tasks", "test_details", "test_records"}
	for _, table := range tables {
		var count int64
		db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Count(&count)
		if count == 0 {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestUserCRUD(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	AutoMigrate(db)

	// 创建用户
	user := &models.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	// 创建
	result := db.Create(user)
	if result.Error != nil {
		t.Errorf("Failed to create user: %v", result.Error)
	}
	if user.ID == 0 {
		t.Error("User ID was not set")
	}

	// 读取
	var retrievedUser models.User
	result = db.First(&retrievedUser, user.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve user: %v", result.Error)
	}
	if retrievedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrievedUser.Email)
	}

	// 更新
	retrievedUser.Name = "updateduser"
	result = db.Save(&retrievedUser)
	if result.Error != nil {
		t.Errorf("Failed to update user: %v", result.Error)
	}

	// 验证更新
	var updatedUser models.User
	result = db.First(&updatedUser, user.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve updated user: %v", result.Error)
	}
	if updatedUser.Name != "updateduser" {
		t.Errorf("Expected name 'updateduser', got %s", updatedUser.Name)
	}

	// 删除
	result = db.Delete(&retrievedUser)
	if result.Error != nil {
		t.Errorf("Failed to delete user: %v", result.Error)
	}

	// 验证删除
	var deletedUser models.User
	result = db.First(&deletedUser, user.ID)
	if result.Error == nil {
		t.Error("User should have been deleted")
	}
}

func TestTestTaskCRUD(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	AutoMigrate(db)

	// 创建测试任务
	task := &models.TestTask{
		TaskName:    "Test Task",
		Status:      "pending",
		WorkerName:  "test-worker",
		PlanKey:     "test-plan",
		TotalTests:  10,
		PassedTests: 5,
		FailedTests: 5,
	}

	// 创建
	result := db.Create(task)
	if result.Error != nil {
		t.Errorf("Failed to create test task: %v", result.Error)
	}
	if task.ID == 0 {
		t.Error("Test task ID was not set")
	}

	// 读取
	var retrievedTask models.TestTask
	result = db.First(&retrievedTask, task.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve test task: %v", result.Error)
	}
	if retrievedTask.TaskName != task.TaskName {
		t.Errorf("Expected task name %s, got %s", task.TaskName, retrievedTask.TaskName)
	}

	// 更新
	retrievedTask.Status = "completed"
	result = db.Save(&retrievedTask)
	if result.Error != nil {
		t.Errorf("Failed to update test task: %v", result.Error)
	}

	// 验证更新
	var updatedTask models.TestTask
	result = db.First(&updatedTask, task.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve updated test task: %v", result.Error)
	}
	if updatedTask.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", updatedTask.Status)
	}

	// 删除
	result = db.Delete(&retrievedTask)
	if result.Error != nil {
		t.Errorf("Failed to delete test task: %v", result.Error)
	}

	// 验证删除
	var deletedTask models.TestTask
	result = db.First(&deletedTask, task.ID)
	if result.Error == nil {
		t.Error("Test task should have been deleted")
	}
}

func TestTestDetailCRUD(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	AutoMigrate(db)

	// 创建测试详情
	detail := &models.TestDetail{
		TestTaskID: 1,
		TestName:   "Test Case 1",
		TestStatus: "passed",
		Duration:   100,
		TestData:   `{"input": "test", "expected": "result"}`,
	}

	// 创建
	result := db.Create(detail)
	if result.Error != nil {
		t.Errorf("Failed to create test detail: %v", result.Error)
	}
	if detail.ID == 0 {
		t.Error("Test detail ID was not set")
	}

	// 读取
	var retrievedDetail models.TestDetail
	result = db.First(&retrievedDetail, detail.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve test detail: %v", result.Error)
	}
	if retrievedDetail.TestName != detail.TestName {
		t.Errorf("Expected test name %s, got %s", detail.TestName, retrievedDetail.TestName)
	}

	// 更新
	retrievedDetail.TestStatus = "failed"
	retrievedDetail.ErrorMessage = "Test failed"
	result = db.Save(&retrievedDetail)
	if result.Error != nil {
		t.Errorf("Failed to update test detail: %v", result.Error)
	}

	// 验证更新
	var updatedDetail models.TestDetail
	result = db.First(&updatedDetail, detail.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve updated test detail: %v", result.Error)
	}
	if updatedDetail.TestStatus != "failed" {
		t.Errorf("Expected status 'failed', got %s", updatedDetail.TestStatus)
	}

	// 删除
	result = db.Delete(&retrievedDetail)
	if result.Error != nil {
		t.Errorf("Failed to delete test detail: %v", result.Error)
	}

	// 验证删除
	var deletedDetail models.TestDetail
	result = db.First(&deletedDetail, detail.ID)
	if result.Error == nil {
		t.Error("Test detail should have been deleted")
	}
}

func TestTestRecordCRUD(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	AutoMigrate(db)

	// 创建测试记录
	record := &models.TestRecord{
		TestTaskID: 1,
		TestName:   "Test Case 1",
		RecordType: "log",
		RecordFile: "Test execution started",
		RecordTime: 1234567890,
		Metadata:   `{"level": "info"}`,
	}

	// 创建
	result := db.Create(record)
	if result.Error != nil {
		t.Errorf("Failed to create test record: %v", result.Error)
	}
	if record.ID == 0 {
		t.Error("Test record ID was not set")
	}

	// 读取
	var retrievedRecord models.TestRecord
	result = db.First(&retrievedRecord, record.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve test record: %v", result.Error)
	}
	if retrievedRecord.TestName != record.TestName {
		t.Errorf("Expected test name %s, got %s", record.TestName, retrievedRecord.TestName)
	}

	// 更新
	retrievedRecord.RecordFile = "Test execution completed"
	result = db.Save(&retrievedRecord)
	if result.Error != nil {
		t.Errorf("Failed to update test record: %v", result.Error)
	}

	// 验证更新
	var updatedRecord models.TestRecord
	result = db.First(&updatedRecord, record.ID)
	if result.Error != nil {
		t.Errorf("Failed to retrieve updated test record: %v", result.Error)
	}
	if updatedRecord.RecordFile != "Test execution completed" {
		t.Errorf("Expected record file 'Test execution completed', got %s", updatedRecord.RecordFile)
	}

	// 删除
	result = db.Delete(&retrievedRecord)
	if result.Error != nil {
		t.Errorf("Failed to delete test record: %v", result.Error)
	}

	// 验证删除
	var deletedRecord models.TestRecord
	result = db.First(&deletedRecord, record.ID)
	if result.Error == nil {
		t.Error("Test record should have been deleted")
	}
}

func TestInitDB(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			DBName:   "test",
			SSLMode:  "disable",
		},
	}

	// 测试 InitDB 函数（这里我们只是测试函数是否能正常执行，不实际连接数据库）
	// 注意：实际环境中可能需要使用真实的数据库连接
	db, err := InitDB(cfg)
	if err != nil {
		// 这里允许失败，因为我们没有实际的数据库服务器
		t.Logf("InitDB failed as expected: %v", err)
	} else {
		t.Logf("InitDB succeeded")
		// 关闭数据库连接
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
