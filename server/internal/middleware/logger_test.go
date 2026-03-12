package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogger(t *testing.T) {
	// 创建 Gin 引擎
	router := gin.Default()

	// 使用日志中间件
	router.Use(Logger())

	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

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

	// 日志中间件应该正常执行，没有错误
	// 由于日志是输出到标准输出，这里我们只需要确保中间件能够正常执行，不会导致程序崩溃
}
