package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/database"
)

func TestSystemFunctionality(t *testing.T) {
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

	// 1. 测试用户注册
	t.Log("Step 1: Testing user registration")
	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, err := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	registerReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// 初始化权限数据
	initPermissions(db)

	// 2. 测试用户登录
	t.Log("Step 2: Testing user login")
	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, err := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 提取 token
	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(w.Body.String()), &loginResponse); err != nil {
		t.Errorf("Failed to parse login response: %v", err)
	}

	token := loginResponse.Data.Token

	// 3. 测试创建测试任务
	t.Log("Step 3: Testing test task creation")
	taskPayload := `{"task_name":"System Test Task","status":"pending","total_tests":3}`
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

	// 提取任务 ID
	var taskResponse struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(w.Body.String()), &taskResponse); err != nil {
		t.Errorf("Failed to parse task response: %v", err)
	}

	taskID := taskResponse.Data.ID

	// 4. 测试创建测试详情
	t.Log("Step 4: Testing test detail creation")
	detailPayload := `{"test_task_id":` + strconv.Itoa(int(taskID)) + `,"test_name":"Test Case 1","test_status":"passed","execution_time":100}`
	detailReq, err := http.NewRequest("POST", "/api/test-details", strings.NewReader(detailPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	detailReq.Header.Set("Content-Type", "application/json")
	detailReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, detailReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// 5. 测试创建测试记录
	t.Log("Step 5: Testing test record creation")
	recordPayload := `{"test_task_id":` + strconv.Itoa(int(taskID)) + `,"test_name":"Test Case 1","record_type":"log","record_data":"Test passed successfully"}`
	recordReq, err := http.NewRequest("POST", "/api/test-records", strings.NewReader(recordPayload))
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	recordReq.Header.Set("Content-Type", "application/json")
	recordReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, recordReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// 6. 测试获取测试任务详情
	t.Log("Step 6: Testing test task retrieval")
	taskDetailReq, err := http.NewRequest("GET", "/api/test-tasks/"+strconv.Itoa(int(taskID)), nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	taskDetailReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, taskDetailReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 7. 测试列出测试任务
	t.Log("Step 7: Testing test task listing")
	taskListReq, err := http.NewRequest("GET", "/api/test-tasks", nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	taskListReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, taskListReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 8. 测试获取个人资料
	t.Log("Step 8: Testing profile retrieval")
	profileReq, err := http.NewRequest("GET", "/api/auth/profile", nil)
	if err != nil {
		t.Errorf("NewRequest returned error: %v", err)
	}
	profileReq.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, profileReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	t.Log("System functionality test completed successfully")
}
