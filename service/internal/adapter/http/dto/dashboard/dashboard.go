// Package dashboard 定义 dashboard 上下文的 HTTP 响应 DTO。
//
// 字段严格取自 Python dashboard/constants.py，含命名漂移修正：
//   - started_at ← start_time（D-14）
//   - item_id ← case_id（D-15）
package dashboard

import (
	"time"

	"github.com/hermes-platform/go-service/internal/domain/dashboard"
)

// SummaryItem 对应 constants.DashboardSummaryItem。
type SummaryItem struct {
	ActiveJobsCount        int       `json:"active_jobs_count"`
	TotalExecutionsCount   int       `json:"total_executions_count"`
	RunningExecutionsCount int       `json:"running_executions_count"`
	SuccessCases7d         int       `json:"success_cases_7d"`
	FailureCases7d         int       `json:"failure_cases_7d"`
	SkippedCases7d         int       `json:"skipped_cases_7d"`
	PassRate7d             float64   `json:"pass_rate_7d"`
	AvgExecutionDuration7d *float64  `json:"avg_execution_duration_7d"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// TrendItem 对应 constants.DashboardTrendItem。
type TrendItem struct {
	StatDate       time.Time `json:"stat_date"`
	ExecutionTotal int       `json:"execution_total"`
	SuccessCases   int       `json:"success_cases"`
	FailureCases   int       `json:"failure_cases"`
	SkippedCases   int       `json:"skipped_cases"`
	RunningCases   int       `json:"running_cases"`
}

// RunningExecutionItem 对应 constants.DashboardRunningExecutionItem。
// 字段 started_at 来自快照表的 start_time（D-14 显式映射）。
type RunningExecutionItem struct {
	ExecutionID         int64     `json:"execution_id"`
	BuildUID            string    `json:"build_uid"`
	JobID               int64     `json:"job_id"`
	JobName             string    `json:"job_name"`
	JobURL              string    `json:"job_url"`
	Status              string    `json:"status"`
	StartedAt           time.Time `json:"started_at"`
	Duration            *float64  `json:"duration"`
	PreCasesCount       int       `json:"pre_cases_count"`
	CompletedCasesCount int       `json:"completed_cases_count"`
	SuccessCasesCount   int       `json:"success_cases_count"`
	FailureCasesCount   int       `json:"failure_cases_count"`
	SkippedCasesCount   int       `json:"skipped_cases_count"`
	RunningCasesCount   int       `json:"running_cases_count"`
	ProgressPercent     float64   `json:"progress_percent"`
	RunningCaseNames    []string  `json:"running_case_names"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// RunningCaseItem 对应 constants.DashboardRunningCaseItem。
// 字段 item_id 来自快照表的 case_id（D-15 显式映射）。
type RunningCaseItem struct {
	ExecutionID   int64      `json:"execution_id"`
	ItemID        int64      `json:"item_id"`
	CaseKey       string     `json:"case_key"`
	CaseName      string     `json:"case_name"`
	AttemptNumber int        `json:"attempt_number"`
	Status        string     `json:"status"`
	StartAt       *time.Time `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
	Duration      *float64   `json:"duration"`
	ErrorMessage  *string    `json:"error_message"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// OverviewData 对应 constants.DashboardOverviewData。
type OverviewData struct {
	Summary           SummaryItem            `json:"summary"`
	Trends            []TrendItem            `json:"trends"`
	RunningExecutions []RunningExecutionItem `json:"running_executions"`
}

// RunningCaseListData 对应 constants.DashboardRunningCaseListData。
type RunningCaseListData struct {
	ExecutionID int64             `json:"execution_id"`
	Items       []RunningCaseItem `json:"items"`
}

// ---------------------------------------------------------------------------
// 领域 → DTO 映射
// ---------------------------------------------------------------------------

func ToSummaryItem(s dashboard.SummarySnapshot) SummaryItem {
	return SummaryItem{
		ActiveJobsCount:        s.ActiveJobsCount,
		TotalExecutionsCount:   s.TotalExecutionsCount,
		RunningExecutionsCount: s.RunningExecutionsCount,
		SuccessCases7d:         s.SuccessCases7d,
		FailureCases7d:         s.FailureCases7d,
		SkippedCases7d:         s.SkippedCases7d,
		PassRate7d:             s.PassRate7d,
		AvgExecutionDuration7d: s.AvgExecutionDuration7d,
		UpdatedAt:              s.UpdatedAt,
	}
}

func ToTrendItem(t dashboard.TrendSnapshot) TrendItem {
	return TrendItem{
		StatDate:       t.StatDate,
		ExecutionTotal: t.ExecutionTotal,
		SuccessCases:   t.SuccessCases,
		FailureCases:   t.FailureCases,
		SkippedCases:   t.SkippedCases,
		RunningCases:   t.RunningCases,
	}
}

func ToRunningExecutionItem(e dashboard.ExecutionSnapshot) RunningExecutionItem {
	names := e.RunningCaseNames
	if names == nil {
		names = []string{}
	}
	return RunningExecutionItem{
		ExecutionID:         e.ExecutionID,
		BuildUID:            e.BuildUID,
		JobID:               e.JobID,
		JobName:             e.JobName,
		JobURL:              e.JobURL,
		Status:              e.Status,
		StartedAt:           e.StartedAt,
		Duration:            e.Duration,
		PreCasesCount:       e.PreCasesCount,
		CompletedCasesCount: e.CompletedCasesCount,
		SuccessCasesCount:   e.SuccessCasesCount,
		FailureCasesCount:   e.FailureCasesCount,
		SkippedCasesCount:   e.SkippedCasesCount,
		RunningCasesCount:   e.RunningCasesCount,
		ProgressPercent:     e.ProgressPercent,
		RunningCaseNames:    names,
		UpdatedAt:           e.UpdatedAt,
	}
}

func ToRunningCaseItem(c dashboard.CaseSnapshot) RunningCaseItem {
	return RunningCaseItem{
		ExecutionID:   c.ExecutionID,
		ItemID:        c.ItemID,
		CaseKey:       c.CaseKey,
		CaseName:      c.CaseName,
		AttemptNumber: c.AttemptNumber,
		Status:        c.Status,
		StartAt:       c.StartedAt,
		EndAt:         c.EndedAt,
		Duration:      c.Duration,
		ErrorMessage:  nullIfEmpty(c.ErrorMessage),
		UpdatedAt:     c.UpdatedAt,
	}
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}
