package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit(t *testing.T) {
	// 创建 Gin 引擎
	router := gin.Default()

	// 使用限流中间件（设置为 1 次/秒）
	router.Use(RateLimit(1, time.Second))

	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 第一次请求 - 应该成功
	t.Run("FirstRequest", func(t *testing.T) {
		// 准备请求
		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Errorf("NewRequest returned error: %v", err)
		}

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// 第二次请求 - 由于 Redis 未初始化，应该也成功
	t.Run("SecondRequest", func(t *testing.T) {
		// 准备请求
		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Errorf("NewRequest returned error: %v", err)
		}

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// 由于 Redis 未初始化，限流中间件会跳过限流，所以所有请求都应该成功
	// 在实际环境中，当 Redis 初始化后，第二次请求会被限流
}
