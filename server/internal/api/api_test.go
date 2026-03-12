package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/database"
	"com.hermes.platform/internal/models"
)

// initPermissions 初始化测试权限数据
func initPermissions(db *gorm.DB) {
	// 创建测试权限
	permissions := []models.Permission{
		{Code: "test_task:create", Name: "创建测试任务"},
		{Code: "test_task:view", Name: "查看测试任务"},
		{Code: "test_task:edit", Name: "编辑测试任务"},
		{Code: "test_task:delete", Name: "删除测试任务"},
		{Code: "test_detail:create", Name: "创建测试详情"},
		{Code: "test_detail:view", Name: "查看测试详情"},
		{Code: "test_detail:edit", Name: "编辑测试详情"},
		{Code: "test_detail:delete", Name: "删除测试详情"},
		{Code: "test_record:create", Name: "创建测试记录"},
		{Code: "test_record:view", Name: "查看测试记录"},
		{Code: "test_record:edit", Name: "编辑测试记录"},
		{Code: "test_record:delete", Name: "删除测试记录"},
	}

	for _, perm := range permissions {
		db.Create(&perm)
	}

	// 创建测试角色
	role := models.Role{
		Name:        "admin",
		Description: "管理员角色",
	}
	db.Create(&role)

	// 为角色分配权限
	var allPermissions []models.Permission
	db.Find(&allPermissions)
	for _, perm := range allPermissions {
		db.Model(&role).Association("Permissions").Append(&perm)
	}

	// 为用户分配角色
	var user models.User
	db.Where("email = ?", "test@example.com").First(&user)
	db.Model(&user).Association("Roles").Append(&role)
}

func TestAuthRegister(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 准备注册请求
	payload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	req, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(payload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestAuthLogin(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 先注册用户
	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	// 测试登录
	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, err := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthChangePassword(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 先注册用户
	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	// 登录获取 token
	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, err := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	// 提取 token
	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(w.Body.String()), &loginResponse); err != nil {
		t.Errorf("Failed to parse login response: %v", err)
	}

	// 测试修改密码
	changePasswordPayload := `{"old_password":"test123","new_password":"newpassword123"}`
	changePasswordReq, err := http.NewRequest("POST", "/api/auth/change-password", strings.NewReader(changePasswordPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	changePasswordReq.Header.Set("Content-Type", "application/json")
	changePasswordReq.Header.Set("Authorization", "Bearer "+loginResponse.Data.Token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, changePasswordReq)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthGetProfile(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 先注册用户
	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	// 登录获取 token
	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, err := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	// 提取 token
	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(w.Body.String()), &loginResponse); err != nil {
		t.Errorf("Failed to parse login response: %v", err)
	}

	// 测试获取个人资料
	profileReq, err := http.NewRequest("GET", "/api/auth/profile", nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	profileReq.Header.Set("Authorization", "Bearer "+loginResponse.Data.Token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, profileReq)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 测试健康检查
	healthReq, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, healthReq)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTestTasks(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 创建 Gin 引擎
	router := gin.Default()
	RegisterRoutes(router, db)

	// 先注册用户
	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	// 初始化权限数据
	initPermissions(db)

	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, err := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(w.Body.String()), &loginResponse); err != nil {
		t.Errorf("Failed to parse login response: %v", err)
	}

	token := loginResponse.Data.Token

	// 测试创建测试任务
	taskPayload := `{"task_name":"Test Task","status":"pending","total_tests":5}`
	taskReq, err := http.NewRequest("POST", "/api/test-tasks", strings.NewReader(taskPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	taskReq.Header.Set("Content-Type", "application/json")
	taskReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, taskReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// 测试列出测试任务
	listReq, err := http.NewRequest("GET", "/api/test-tasks", nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	listReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, listReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}


