package models

import (
	"gorm.io/gorm"
)

// TestStepDetail 测试步骤详情模型
type TestStepDetail struct {
	gorm.Model
	TestDetailID     uint   `json:"test_detail_id" gorm:"index"` // 关联到测试详情
	StepName         string `json:"step_name"` // 步骤名称
	StartTime        int64  `json:"start_time"` // 开始时间
	EndTime          int64  `json:"end_time"` // 结束时间
	Duration         int64  `json:"duration"` // 执行时长（毫秒）
	Passed           bool   `json:"passed"` // 是否通过
	Screenshot       string `json:"screenshot"` // 截图（文件链接）
	VerificationArea string `json:"verification_area"` // 验证区域（JSON 格式）
}
