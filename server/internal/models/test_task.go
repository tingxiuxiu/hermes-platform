package models

import (
	"errors"

	"gorm.io/gorm"
)

// TestTask 测试任务执行结果模型
type TestTask struct {
	gorm.Model
	BuildID     string       `json:"build_id" gorm:"uniqueIndex;type:varchar(255)"` // BuildID 唯一标识符，由用户上传
	TaskName    string       `json:"task_name"`
	Status      string       `json:"status" gorm:"index"` // 执行状态：pending, running, completed, failed
	StartTime   int64        `json:"start_time" gorm:"index"`
	EndTime     int64        `json:"end_time"`
	Duration    int64        `json:"duration"` // 执行时长（毫秒）
	TotalTests  int          `json:"total_tests"`
	PassedTests int          `json:"passed_tests"`
	FailedTests int          `json:"failed_tests"`
	WorkerName  string       `json:"worker_name" gorm:"index;not null"` // 执行测试的节点名称
	PlanKey     string       `json:"plan_key" gorm:"index;not null"`    // 测试所属的测试集
	TestDetails []TestDetail `json:"test_details" gorm:"foreignKey:TestTaskID"`
}

// BeforeCreate 创建前钩子，检查 BuildID 唯一性和合法性
func (t *TestTask) BeforeCreate(tx *gorm.DB) error {
	// 检查 BuildID 是否为空
	if t.BuildID == "" {
		return errors.New("BuildID cannot be empty")
	}
	// 检查 BuildID 是否为合法的 UUID 格式
	if !isValidUUID(t.BuildID) {
		return errors.New("BuildID must be a valid UUID format")
	}
	// 检查 BuildID 是否已存在
	var count int64
	tx.Model(&TestTask{}).Where("build_id = ?", t.BuildID).Count(&count)
	if count > 0 {
		return errors.New("BuildID already exists")
	}
	return nil
}

// isValidUUID 检查字符串是否为合法的 UUID 格式
func isValidUUID(uuid string) bool {
	// 简单的 UUID 格式验证
	if len(uuid) != 36 {
		return false
	}
	// 检查 UUID 的格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	for i, char := range uuid {
		switch i {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}
	return true
}
