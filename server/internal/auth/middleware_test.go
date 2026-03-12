package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"com.hermes.platform/internal/config"
)

func TestAuthMiddleware(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	InitJWT(cfg)

	// 创建 Gin 引擎
	router := gin.Default()

	// 测试路由
	router.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试 1: 没有 Authorization 头
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// 测试 2: 无效的 Authorization 头格式
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Invalid format")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// 测试 3: 无效的 token
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// 测试 4: 有效的 token
	userID := uint(1)
	email := "test@example.com"
	roles := []string{"user"}
	token, _ := GenerateToken(userID, email, roles)

	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRoleMiddleware(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	InitJWT(cfg)

	// 创建 Gin 引擎
	router := gin.Default()

	// 测试路由
	router.GET("/admin", AuthMiddleware(), RoleMiddleware("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试 1: 没有 admin 角色的用户
	userID := uint(1)
	email := "test@example.com"
	roles := []string{"user"}
	token, _ := GenerateToken(userID, email, roles)

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}

	// 测试 2: 有 admin 角色的用户
	adminToken, _ := GenerateToken(userID, email, []string{"admin"})

	req, _ = http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
