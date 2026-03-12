package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"com.hermes.platform/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})

	// 创建用户仓库
	repo := NewUserRepository(db)

	// 创建测试用户
	user := &models.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "test123",
	}

	// 测试创建用户
	err = repo.Create(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	// 验证用户是否创建成功
	var createdUser models.User
	db.First(&createdUser, user.ID)
	if createdUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, createdUser.Email)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})

	// 创建用户仓库
	repo := NewUserRepository(db)

	// 创建测试用户
	user := &models.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "test123",
	}
	db.Create(user)

	// 测试根据邮箱查找用户
	foundUser, err := repo.FindByEmail("test@example.com")
	if err != nil {
		t.Errorf("Failed to find user by email: %v", err)
	}

	if foundUser == nil {
		t.Error("Expected user, got nil")
	}

	if foundUser.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", foundUser.Email)
	}

	// 测试查找不存在的用户
	_, err = repo.FindByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("Expected error when finding nonexistent user, got nil")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})

	// 创建用户仓库
	repo := NewUserRepository(db)

	// 创建测试用户
	user := &models.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "test123",
	}
	db.Create(user)

	// 测试根据 ID 查找用户
	foundUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Errorf("Failed to find user by ID: %v", err)
	}

	if foundUser == nil {
		t.Error("Expected user, got nil")
	}

	if foundUser.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, foundUser.ID)
	}

	// 测试查找不存在的用户
	_, err = repo.FindByID(999)
	if err == nil {
		t.Error("Expected error when finding nonexistent user, got nil")
	}
}

func TestUserRepository_Update(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})

	// 创建用户仓库
	repo := NewUserRepository(db)

	// 创建测试用户
	user := &models.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "test123",
	}
	db.Create(user)

	// 更新用户信息
	user.Name = "updateduser"
	err = repo.Update(user)
	if err != nil {
		t.Errorf("Failed to update user: %v", err)
	}

	// 验证用户是否更新成功
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	if updatedUser.Name != "updateduser" {
		t.Errorf("Expected name updateduser, got %s", updatedUser.Name)
	}
}