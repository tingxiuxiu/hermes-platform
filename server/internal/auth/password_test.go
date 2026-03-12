package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "test123"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword returned error: %v", err)
	}
	if hashed == "" {
		t.Error("HashPassword returned empty string")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "test123"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword returned error: %v", err)
	}

	// 正确的密码应该返回 nil 错误
	err = VerifyPassword(hashed, password)
	if err != nil {
		t.Errorf("VerifyPassword returned error for correct password: %v", err)
	}

	// 错误的密码应该返回错误
	err = VerifyPassword(hashed, "wrongpassword")
	if err == nil {
		t.Error("VerifyPassword should return error for wrong password")
	}
}
