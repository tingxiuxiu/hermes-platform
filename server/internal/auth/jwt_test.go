package auth

import (
	"testing"

	"com.hermes.platform/internal/config"
)

func TestGenerateToken(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24, // 24小时
		},
	}
	InitJWT(cfg)

	userID := uint(1)
	email := "test@example.com"
	roles := []string{"user"}
	token, err := GenerateToken(userID, email, roles)
	if err != nil {
		t.Errorf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken returned empty string")
	}
}

func TestParseToken(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24, // 24小时
		},
	}
	InitJWT(cfg)

	userID := uint(1)
	email := "test@example.com"
	roles := []string{"user"}
	token, err := GenerateToken(userID, email, roles)
	if err != nil {
		t.Errorf("GenerateToken returned error: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Errorf("ParseToken returned error: %v", err)
	}
	if claims == nil {
		t.Error("ParseToken returned nil claims")
		return
	}
	if claims.UserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	// 初始化 JWT 配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24, // 24小时
		},
	}
	InitJWT(cfg)

	invalidToken := "invalid.token.here"
	_, err := ParseToken(invalidToken)
	if err == nil {
		t.Error("ParseToken should return error for invalid token")
	}
}
