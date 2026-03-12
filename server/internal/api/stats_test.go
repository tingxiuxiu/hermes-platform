package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/database"
	"com.hermes.platform/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStatsTest(t *testing.T) (*gin.Engine, *gorm.DB, string) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test_secret_key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	database.AutoMigrate(db)

	router := gin.Default()
	RegisterRoutes(router, db)

	registerPayload := `{"name":"testuser","password":"test123","email":"test@example.com"}`
	registerReq, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)

	initPermissions(db)

	loginPayload := `{"email":"test@example.com","password":"test123"}`
	loginReq, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReq)

	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal([]byte(w.Body.String()), &loginResponse)

	return router, db, loginResponse.Data.Token
}

func createTestTaskForStats(db *gorm.DB, taskName string, status string, totalTests int, passedTests int, failedTests int) {
	task := models.TestTask{
		TaskName:    taskName,
		Status:      status,
		TotalTests:  totalTests,
		PassedTests: passedTests,
		FailedTests: failedTests,
		StartTime:   1704067200,
	}
	db.Create(&task)
}

func TestGetDashboardStats(t *testing.T) {
	router, db, token := setupStatsTest(t)

	createTestTaskForStats(db, "Task 1", "completed", 10, 8, 2)
	createTestTaskForStats(db, "Task 2", "running", 5, 3, 1)

	req, _ := http.NewRequest("GET", "/api/stats/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	successVal, ok := response["success"]
	if !ok {
		dataVal, hasData := response["data"]
		if !hasData {
			t.Errorf("Expected data field to be present")
			return
		}
		dataMap, ok := dataVal.(map[string]interface{})
		if !ok {
			t.Errorf("Expected data to be a map")
			return
		}
		if dataMap["stats"] == nil {
			t.Errorf("Expected stats to be present")
		}
		return
	}

	if successVal != true {
		t.Errorf("Expected success to be true, got %v", successVal)
	}
}

func TestGetTrendData(t *testing.T) {
	router, _, token := setupStatsTest(t)

	req, _ := http.NewRequest("GET", "/api/stats/trend?range=week", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetTrendDataWithDifferentRanges(t *testing.T) {
	router, _, token := setupStatsTest(t)

	ranges := []string{"today", "week", "month", "year"}
	for _, r := range ranges {
		req, _ := http.NewRequest("GET", "/api/stats/trend?range="+r, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d for range %s, got %d", http.StatusOK, r, w.Code)
		}
	}
}

func TestGetRunningTasks(t *testing.T) {
	router, db, token := setupStatsTest(t)

	createTestTaskForStats(db, "Running Task 1", "running", 10, 5, 2)
	createTestTaskForStats(db, "Running Task 2", "running", 8, 4, 1)
	createTestTaskForStats(db, "Completed Task", "completed", 5, 5, 0)

	req, _ := http.NewRequest("GET", "/api/stats/running-tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetRunningTasksEmpty(t *testing.T) {
	router, _, token := setupStatsTest(t)

	req, _ := http.NewRequest("GET", "/api/stats/running-tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestStatsUnauthorized(t *testing.T) {
	router, _, _ := setupStatsTest(t)

	endpoints := []string{
		"/api/stats/dashboard",
		"/api/stats/trend",
		"/api/stats/running-tasks",
	}

	for _, endpoint := range endpoints {
		req, _ := http.NewRequest("GET", endpoint, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d for %s, got %d", http.StatusUnauthorized, endpoint, w.Code)
		}
	}
}
