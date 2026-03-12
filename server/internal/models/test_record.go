package models

import (
	"gorm.io/gorm"
)

// TestRecord 测试记录数据模型
type TestRecord struct {
	gorm.Model
	TestTaskID uint   `json:"test_task_id" gorm:"index"`
	TestName   string `json:"test_name"`
	RecordType string `json:"record_type" gorm:"index"` // 记录类型：log, metric, screenshot, etc.
	RecordFile string `json:"record_file"`              // 记录数据文件路径（根据类型不同，可能是文本、JSON、Base64 等）
	RecordTime int64  `json:"record_time" gorm:"index"`
	Metadata   string `json:"metadata"` // 附加元数据（JSON 格式）
}
