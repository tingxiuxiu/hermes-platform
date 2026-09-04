// Package automation 定义 automation 上下文的 HTTP 请求/响应 DTO。
//
// 字段与 Python 侧一一对应，不得擅自重命名或省略（ADR-0001）。
// 分页字段是 **rows**（与 identity 的 items 不同，这是契约的一部分，不得统一）。
package automation

import (
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/domain/automation"
)

// ---------------------------------------------------------------------------
// 请求 DTO（对齐 Python schemas.py）
// ---------------------------------------------------------------------------

// CreateExecutionRequest 对应 CreateExecution。
type CreateExecutionRequest struct {
	BuildUID          uuid.UUID `json:"build_uid" binding:"required"`
	JobName           string    `json:"job_name" binding:"required,max=255"`
	JobURL            string    `json:"job_url" binding:"required,max=255"`
	ProjectName       string    `json:"project_name" binding:"required,max=255"`
	SoftwareName      string    `json:"software_name" binding:"required,max=255"`
	SoftwareVersion   string    `json:"software_version" binding:"required,max=255"`
	Labels            []string  `json:"labels"`
	StartTime         time.Time `json:"start_time" binding:"required"`
	PlannedCasesCount *int      `json:"planned_cases_count"`
}

// UpdateExecutionRequest 对应 UpdateExecution。
type UpdateExecutionRequest struct {
	Status   *automation.ExecutionStatus `json:"status"`
	EndTime  *time.Time                  `json:"end_time"`
	Duration *float64                    `json:"duration"`
}

// CreateItemRequest 对应 CreateExecutionItem。
// case_key 在 Python 是 max_length=512，但 DB 列是 VARCHAR(128)（缺陷 D-14）。
// Go 侧以 DB 约束为准（128），超长会被领域实体拒绝——避免写入后读不出来。
type CreateItemRequest struct {
	BuildUID  string     `json:"build_uid" binding:"required"`
	CaseKey   string     `json:"case_key" binding:"required"`
	CaseName  string     `json:"case_name" binding:"required"`
	CaseUID   uuid.UUID  `json:"case_uid" binding:"required"`
	Labels    []string   `json:"labels"`
	StartTime *time.Time `json:"start_time"`
}

// UpdateItemRequest 对应 UpdateExecutionItem。
type UpdateItemRequest struct {
	Status         automation.CaseStatus   `json:"status" binding:"required"`
	EndTime        *time.Time              `json:"end_time"`
	Duration       *float64                `json:"duration"`
	ErrorMessage   string                  `json:"error_message"`
	ErrorTraceback string                  `json:"error_traceback"`
	Attachments    []AttachmentItemRequest `json:"attachments"`
}

// AttachmentItemRequest 对应 schemas.AttachmentItem。
type AttachmentItemRequest struct {
	AttachmentType string  `json:"attachment_type" binding:"required"`
	FileName       string  `json:"file_name"`
	URL            string  `json:"url" binding:"required"`
	MimeType       *string `json:"mime_type"`
}

// CreateStepRequest 对应 CreateStep（批量上报时 list 里的一项）。
type CreateStepRequest struct {
	CaseUID         uuid.UUID               `json:"case_uid" binding:"required"`
	StepIndex       *int                    `json:"step_index"`
	ParentStepIndex *int                    `json:"parent_step_index"`
	StepPath        string                  `json:"step_path" binding:"required,max=255"`
	StepName        string                  `json:"step_name" binding:"required,max=255"`
	Status          automation.StepStatus   `json:"status" binding:"required"`
	StartTime       *time.Time              `json:"start_time"`
	EndTime         *time.Time              `json:"end_time"`
	Duration        *float64                `json:"duration"`
	Attachments     []AttachmentItemRequest `json:"attachments"`
}

// UpsertStepRequest 是 live 主路径：POST /automation/items/{case_uid}/steps。
// case_uid 在路径上；attachments 即使传入也忽略。
type UpsertStepRequest struct {
	StepPath    string                  `json:"step_path" binding:"required,max=255"`
	StepName    string                  `json:"step_name" binding:"required,max=255"`
	Status      automation.StepStatus   `json:"status" binding:"required"`
	StartTime   *time.Time              `json:"start_time"`
	EndTime     *time.Time              `json:"end_time"`
	Duration    *float64                `json:"duration"`
	Attachments []AttachmentItemRequest `json:"attachments"`
}

// ---------------------------------------------------------------------------
// 响应 DTO（对齐 Python constants.py）
// ---------------------------------------------------------------------------

// ExecutionRow 对应 constants.ExecutionRow。
type ExecutionRow struct {
	ID                int64      `json:"id"`
	BuildUID          string     `json:"build_uid"`
	JobName           string     `json:"job_name"`
	JobURL            *string    `json:"job_url"`
	ProjectName       string     `json:"project_name"`
	SoftwareName      string     `json:"software_name"`
	SoftwareVersion   string     `json:"software_version"`
	Labels            []string   `json:"labels"`
	Status            string     `json:"status"`
	StartTime         time.Time  `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	Duration          *float64   `json:"duration"`
	PlannedCasesCount *int       `json:"planned_cases_count"`
	PassCount         *int       `json:"pass_count"`
	FailureCount      *int       `json:"failure_count"`
	SkippedCount      *int       `json:"skipped_count"`
	PassRate          *float64   `json:"pass_rate"`
	LastHeartbeatAt   *time.Time `json:"last_heartbeat_at,omitempty"`
}

// HeartbeatData 是心跳接口的响应。
type HeartbeatData struct {
	LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
}

// ExecutionItemRow 对应 constants.ExecutionItemRow。
type ExecutionItemRow struct {
	ID             int64               `json:"id"`
	BuildUID       string              `json:"build_uid"`
	CaseUID        string              `json:"case_uid"`
	CaseKey        string              `json:"case_key"`
	CaseName       string              `json:"case_name"`
	Labels         []string            `json:"labels"`
	AttemptNumber  int                 `json:"attempt_number"`
	IsLatest       bool                `json:"is_latest"`
	Status         string              `json:"status"`
	StartTime      *time.Time          `json:"start_time"`
	EndTime        *time.Time          `json:"end_time"`
	Duration       *float64            `json:"duration"`
	ErrorMessage   *string             `json:"error_message"`
	ErrorTraceback *string             `json:"error_traceback"`
	Attachments    []AttachmentItemRow `json:"attachments"`
	Steps          []StepRow           `json:"steps"`
}

// StepRow 对应 constants.StepRow。
type StepRow struct {
	ID              int64               `json:"id"`
	CaseUID         string              `json:"case_uid"`
	ParentStepID    *int64              `json:"parent_step_id"`
	ParentStepIndex *int                `json:"parent_step_index"`
	StepIndex       int                 `json:"step_index"`
	StepPath        string              `json:"step_path"`
	Depth           int                 `json:"depth"`
	StepName        string              `json:"step_name"`
	Status          string              `json:"status"`
	StartTime       *time.Time          `json:"start_time"`
	EndTime         *time.Time          `json:"end_time"`
	Duration        *float64            `json:"duration"`
	SubSteps        []StepRow           `json:"sub_steps"`
	Attachments     []AttachmentItemRow `json:"attachments"`
}

// AttachmentItemRow 对应 constants.AttachmentItem。
type AttachmentItemRow struct {
	AttachmentType string  `json:"attachment_type"`
	FileName       string  `json:"file_name"`
	URL            string  `json:"url"`
	MimeType       *string `json:"mime_type"`
}

// PipelineRow 对应 constants.PipelineRow。
type PipelineRow struct {
	ID                 int64          `json:"id"`
	JobName            string         `json:"job_name"`
	JobURL             string         `json:"job_url"`
	Status             string         `json:"status"`
	LastBuildNumber    *int           `json:"last_build_number"`
	LastBuildUID       *string        `json:"last_build_uid"`
	LastBuildStatus    *string        `json:"last_build_status"`
	LastBuildTimestamp *time.Time     `json:"last_build_timestamp"`
	LastBuildDuration  *float64       `json:"last_build_duration"`
	PipelineParams     map[string]any `json:"pipeline_params"`
	SyncAt             *time.Time     `json:"sync_at"`
}

// ---------------------------------------------------------------------------
// 分页封装（对齐 constants 的 *ListData）
// ---------------------------------------------------------------------------

type ExecutionListData struct {
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Rows     []ExecutionRow `json:"rows"`
}

type PipelineListData struct {
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Rows     []PipelineRow `json:"rows"`
}

type ExecutionItemListData struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Rows     []ExecutionItemRow `json:"rows"`
}

// ItemSummaryRow 是会话列表/live 摘要：不带步骤树与附件。
type ItemSummaryRow struct {
	ID            int64      `json:"id"`
	CaseUID       string     `json:"case_uid"`
	CaseKey       string     `json:"case_key"`
	CaseName      string     `json:"case_name"`
	AttemptNumber int        `json:"attempt_number"`
	Status        string     `json:"status"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Duration      *float64   `json:"duration"`
}

// ItemSummaryListData 是 GET /executions/{build_uid}/items 的分页摘要。
type ItemSummaryListData struct {
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Rows     []ItemSummaryRow `json:"rows"`
}

// LiveCurrentItem 是 live 快照里当前 running 用例（含步骤树）。
type LiveCurrentItem struct {
	CaseUID  string    `json:"case_uid"`
	CaseKey  string    `json:"case_key"`
	CaseName string    `json:"case_name"`
	Status   string    `json:"status"`
	Steps    []StepRow `json:"steps"`
}

// LiveSnapshotData 是 GET /executions/{build_uid}/live 的响应。
type LiveSnapshotData struct {
	Execution   ExecutionRow     `json:"execution"`
	Items       []ItemSummaryRow `json:"items"`
	CurrentItem *LiveCurrentItem `json:"current_item"`
}
