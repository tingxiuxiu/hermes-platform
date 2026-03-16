package models

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name        string           `json:"name" gorm:"size:100;not null"`
	Description string           `json:"description" gorm:"type:text"`
	Versions    []ProjectVersion `json:"versions" gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedBy   uint             `json:"created_by"`
	UpdatedBy   uint             `json:"updated_by"`
}

type ProjectVersion struct {
	gorm.Model
	ProjectID   uint        `json:"project_id" gorm:"not null;index"`
	Version     string      `json:"version" gorm:"size:50;not null"`
	Description string      `json:"description" gorm:"type:text"`
	Status      string      `json:"status" gorm:"size:20;default:active"` // active, archived
	TestPlans   []TestPlan  `json:"test_plans" gorm:"foreignKey:VersionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TestPlan struct {
	gorm.Model
	VersionID   uint       `json:"version_id" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	Description string     `json:"description" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:20;default:draft"` // draft, active, completed
	TestCases   []TestCase `json:"test_cases" gorm:"foreignKey:PlanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedBy   uint       `json:"created_by"`
	UpdatedBy   uint       `json:"updated_by"`
}

type TestCaseStatus string

const (
	TestCaseStatusDeveloping TestCaseStatus = "developing"
	TestCaseStatusReady      TestCaseStatus = "ready"
	TestCaseStatusSkip       TestCaseStatus = "skip"
)

type TestCase struct {
	gorm.Model
	PlanID     uint           `json:"plan_id" gorm:"not null;index"`
	Name       string         `json:"name" gorm:"size:200;not null"`
	CaseKey    string         `json:"case_key" gorm:"size:50;uniqueIndex;not null"`
	Status     TestCaseStatus `json:"status" gorm:"size:20;default:developing"`
	Remark     string         `json:"remark" gorm:"type:text"`
	PreCondition string       `json:"pre_condition" gorm:"type:text"`
	StepDescription string    `json:"step_description" gorm:"type:text"`
	ExpectedResult string     `json:"expected_result" gorm:"type:text"`
	Priority    int           `json:"priority" gorm:"default:3"` // 1=High, 2=Medium, 3=Low
	CreatedBy   uint          `json:"created_by"`
	UpdatedBy   uint          `json:"updated_by"`
}
