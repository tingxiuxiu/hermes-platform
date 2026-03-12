package utils

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

func TestCheckPassword(t *testing.T) {
	password := "test123"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword returned error: %v", err)
	}

	// 正确的密码应该返回 true
	if !CheckPassword(password, hashed) {
		t.Error("CheckPassword returned false for correct password")
	}

	// 错误的密码应该返回 false
	if CheckPassword("wrongpassword", hashed) {
		t.Error("CheckPassword returned true for wrong password")
	}
}
