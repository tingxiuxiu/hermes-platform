package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockPermissionChecker struct {
	userPermissions map[uint][]string
}

func (m *mockPermissionChecker) CheckPermission(userID uint, permissionCode string) (bool, error) {
	permissions, exists := m.userPermissions[userID]
	if !exists {
		return false, nil
	}
	for _, perm := range permissions {
		if perm == permissionCode {
			return true, nil
		}
	}
	return false, nil
}

func newMockPermissionChecker() *mockPermissionChecker {
	return &mockPermissionChecker{
		userPermissions: make(map[uint][]string),
	}
}

func (m *mockPermissionChecker) setUserPermissions(userID uint, permissions []string) {
	m.userPermissions[userID] = permissions
}

func TestPermissionMiddleware_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	router := gin.New()
	router.GET("/test", PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPermissionMiddleware_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("userID", "invalid")
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNormalUserPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	normalUserID := uint(1)
	mockChecker.setUserPermissions(normalUserID, []string{"test:view"})

	router := gin.New()

	router.GET("/test/view", func(c *gin.Context) {
		c.Set("userID", normalUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "view success"})
	})

	router.POST("/test/create", func(c *gin.Context) {
		c.Set("userID", normalUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "create success"})
	})

	router.DELETE("/test/delete", func(c *gin.Context) {
		c.Set("userID", normalUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "delete success"})
	})

	t.Run("NormalUser_CanViewTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test/view", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("NormalUser_CannotCreateTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test/create", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("NormalUser_CannotDeleteTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/test/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestTesterPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	testerUserID := uint(2)
	mockChecker.setUserPermissions(testerUserID, []string{"test:view", "test:create", "test:edit"})

	router := gin.New()

	router.GET("/test/view", func(c *gin.Context) {
		c.Set("userID", testerUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "view success"})
	})

	router.POST("/test/create", func(c *gin.Context) {
		c.Set("userID", testerUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "create success"})
	})

	router.PUT("/test/edit", func(c *gin.Context) {
		c.Set("userID", testerUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:edit"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "edit success"})
	})

	router.DELETE("/test/delete", func(c *gin.Context) {
		c.Set("userID", testerUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "delete success"})
	})

	t.Run("Tester_CanViewTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test/view", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Tester_CanCreateTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test/create", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Tester_CanEditTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/test/edit", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Tester_CannotDeleteTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/test/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestAdminPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	adminUserID := uint(3)
	mockChecker.setUserPermissions(adminUserID, []string{"test:view", "test:create", "test:edit", "test:delete"})

	router := gin.New()

	router.GET("/test/view", func(c *gin.Context) {
		c.Set("userID", adminUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "view success"})
	})

	router.POST("/test/create", func(c *gin.Context) {
		c.Set("userID", adminUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "create success"})
	})

	router.PUT("/test/edit", func(c *gin.Context) {
		c.Set("userID", adminUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:edit"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "edit success"})
	})

	router.DELETE("/test/delete", func(c *gin.Context) {
		c.Set("userID", adminUserID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "delete success"})
	})

	t.Run("Admin_CanViewTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test/view", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Admin_CanCreateTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test/create", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Admin_CanEditTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/test/edit", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Admin_CanDeleteTestTask", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/test/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestPermissionMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	userID := uint(999)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestPermissionMiddleware_MultiplePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChecker := newMockPermissionChecker()
	userID := uint(1)
	mockChecker.setUserPermissions(userID, []string{"test:view", "test:create", "test:edit"})

	router := gin.New()

	router.GET("/test/view", func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "view success"})
	})

	router.POST("/test/create", func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "create success"})
	})

	router.PUT("/test/edit", func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:edit"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "edit success"})
	})

	router.DELETE("/test/delete", func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}, PermissionMiddleware(mockChecker, "test:delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "delete success"})
	})

	t.Run("HasViewPermission", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test/view", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("HasCreatePermission", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test/create", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("HasEditPermission", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/test/edit", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("MissingDeletePermission", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/test/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}
