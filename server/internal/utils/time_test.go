package utils

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	expected := "2023-12-25 10:30:45"
	result := FormatTime(testTime)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestParseTime(t *testing.T) {
	timeStr := "2023-12-25 10:30:45"
	expected := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	result, err := ParseTime(timeStr)
	if err != nil {
		t.Errorf("ParseTime returned error: %v", err)
	}
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGetCurrentTime(t *testing.T) {
	result := GetCurrentTime()
	if result.IsZero() {
		t.Error("GetCurrentTime returned zero time")
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	result := GetCurrentTimestamp()
	if result <= 0 {
		t.Error("GetCurrentTimestamp returned non-positive value")
	}
}

func TestAddDays(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	days := 7
	expected := time.Date(2024, 1, 1, 10, 30, 45, 0, time.UTC)
	result := AddDays(testTime, days)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestIsExpired(t *testing.T) {
	// 测试过期时间
	expiredTime := time.Now().Add(-24 * time.Hour)
	if !IsExpired(expiredTime) {
		t.Error("IsExpired should return true for expired time")
	}

	// 测试未过期时间
	futureTime := time.Now().Add(24 * time.Hour)
	if IsExpired(futureTime) {
		t.Error("IsExpired should return false for future time")
	}
}
