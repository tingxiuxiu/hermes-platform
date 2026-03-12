package models

import (
	"gorm.io/gorm"
)

// TestDetail 测试结果详情模型
type TestDetail struct {
	gorm.Model
	TestTaskID    uint   `json:"test_task_id" gorm:"index"`
	TestName      string `json:"test_name"`                // case key
	TestStatus    string `json:"test_status" gorm:"index"` // 测试状态：passed, failed, skipped
	ErrorMessage  string `json:"error_message"`
	TestStartTime int64  `json:"test_start_time"` // 开始时间
	TestEndTime   int64  `json:"test_end_time"`   // 结束时间
	Duration      int64  `json:"duration"`        // 执行时长（毫秒）
	TestData      string `json:"test_data"`       // 测试数据（JSON 格式）
}
