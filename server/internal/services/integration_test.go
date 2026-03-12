package services

import (
	"testing"

	"com.hermes.platform/internal/database"
	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthService_Integration(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// 创建认证服务
	authService := NewAuthService(userRepo, roleRepo)

	// 测试注册功能
	user, err := authService.Register("testuser", "test@example.com", "test123")
	if err != nil {
		t.Errorf("Failed to register user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set, got 0")
	}

	// 测试登录功能
	token, err := authService.Login("test@example.com", "test123")
	if err != nil {
		t.Errorf("Failed to login: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be set, got empty string")
	}

	// 测试修改密码功能
	err = authService.ChangePassword(user.ID, "test123", "newpassword123")
	if err != nil {
		t.Errorf("Failed to change password: %v", err)
	}

	// 测试使用新密码登录
	tokenNew, err := authService.Login("test@example.com", "newpassword123")
	if err != nil {
		t.Errorf("Failed to login with new password: %v", err)
	}

	if tokenNew == "" {
		t.Error("Expected token to be set, got empty string")
	}
}

func TestTestService_Integration(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建仓库
	testRepo := repository.NewTestRepository(db)

	// 创建测试服务
	testService := NewTestService(testRepo)

	// 测试创建测试任务
	task := &models.TestTask{
		BuildID:     "12345678-1234-1234-1234-123456789012",
		TaskName:    "Test Task",
		Status:      "pending",
		StartTime:   1620000000,
		TotalTests:  5,
		PassedTests: 0,
		FailedTests: 0,
		WorkerName:  "test-worker",
		PlanKey:     "test-plan",
	}

	err = testService.CreateTestTask(task)
	if err != nil {
		t.Errorf("Failed to create test task: %v", err)
	}

	if task.ID == 0 {
		t.Error("Expected task ID to be set, got 0")
	}

	// 测试获取测试任务
	getTask, err := testService.GetTestTaskByID(task.ID)
	if err != nil {
		t.Errorf("Failed to get test task: %v", err)
	}

	if getTask.ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, getTask.ID)
	}

	// 测试列出测试任务
	tasks, total, err := testService.ListTestTasks(1, 10, nil, "", false)
	if err != nil {
		t.Errorf("Failed to list test tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
}
