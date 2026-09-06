package dashboard

import (
	"context"
	"fmt"
	"time"

	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	"github.com/hermes-platform/go-service/internal/domain/dashboard"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// ExecutionSourceReader 实现 application/dashboard.ExecutionReader 端口，
// 直接读 automation 的源表（test_executions / execution_items / jenkins_pipelines）。
//
// 这是 dashboard 读侧的**唯一**源数据入口，复用 SQL 而非经过 automation 应用层，
// 避免读侧依赖写侧用例（保持 CQRS 的读写分离）。
type ExecutionSourceReader struct {
	db *database.Pool
}

// NewExecutionSourceReader 构造源数据读取器。
func NewExecutionSourceReader(db *database.Pool) *ExecutionSourceReader {
	return &ExecutionSourceReader{db: db}
}

var _ appdashboard.ExecutionReader = (*ExecutionSourceReader)(nil)

// ListExecutionsInRange 返回某时间窗口内的 execution 元数据（含终止态与 running）。
func (r *ExecutionSourceReader) ListExecutionsInRange(ctx context.Context, since time.Time) ([]dashboard.TerminalExecution, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT e.id, COALESCE(p.id, 0), e.job_name, COALESCE(e.job_url,''), e.status, e.start_time, e.duration
		  FROM test_executions e
		  LEFT JOIN jenkins_pipelines p ON p.job_name = e.job_name
		 WHERE e.start_time >= $1
		 ORDER BY e.start_time DESC`, since)
	if err != nil {
		return nil, fmt.Errorf("dashboard source: list executions: %w", err)
	}
	defer rows.Close()

	var out []dashboard.TerminalExecution
	for rows.Next() {
		var e dashboard.TerminalExecution
		var jobURL string
		if err := rows.Scan(&e.ExecutionID, &e.JobID, &e.JobName, &jobURL,
			&e.Status, &e.StartedAt, &e.Duration); err != nil {
			return nil, fmt.Errorf("dashboard source: scan execution: %w", err)
		}
		e.JobURL = jobURL
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListTerminalCasesInRange 返回某时间窗口内的终止态用例。
//
// execution_items 用 build_uid 关联 test_executions（无 execution_id 列）。
// 用 execution 的 start_time 归入窗口：用例自己的 start_time 可能为 NULL
// （插件只上报 case_uid，不一定带开始时间），用父 execution 的时间更可靠。
//
// ⚠️ 不依赖 is_latest 列（缺陷 D-02 下全 true），改用窗口函数取每
// (execution, case_key) 的 attempt_number 最大行，避免重复统计重试的旧 attempt。
func (r *ExecutionSourceReader) ListTerminalCasesInRange(ctx context.Context, since time.Time) ([]dashboard.TerminalCase, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT e.id, i.case_key, i.status, e.start_time
		  FROM (
			SELECT build_uid, case_key, status,
			       ROW_NUMBER() OVER (
			           PARTITION BY build_uid, case_key ORDER BY attempt_number DESC, id DESC
			       ) AS rn
			  FROM execution_items
			 WHERE status <> 'running'
		  ) i
		  JOIN test_executions e ON e.build_uid = i.build_uid
		 WHERE e.start_time >= $1 AND i.rn = 1`, since)
	if err != nil {
		return nil, fmt.Errorf("dashboard source: list terminal cases: %w", err)
	}
	defer rows.Close()

	var out []dashboard.TerminalCase
	for rows.Next() {
		var c dashboard.TerminalCase
		if err := rows.Scan(&c.ExecutionID, &c.CaseKey, &c.Status, &c.StartedAt); err != nil {
			return nil, fmt.Errorf("dashboard source: scan terminal case: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetExecutionInput 返回单条 execution 的完整投影输入。
func (r *ExecutionSourceReader) GetExecutionInput(ctx context.Context, executionID int64) (*dashboard.ExecutionInput, error) {
	// execution 主数据 + 从 jenkins_pipelines 解析 job_id
	var in dashboard.ExecutionInput
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		SELECT e.id, e.build_uid::text, COALESCE(p.id, 0), e.job_name, COALESCE(e.job_url,''), e.status, e.start_time, e.duration,
		       e.planned_cases_count
		  FROM test_executions e
		  LEFT JOIN jenkins_pipelines p ON p.job_name = e.job_name
		 WHERE e.id = $1`, executionID,
	).Scan(&in.ExecutionID, &in.BuildUID, &in.JobID, &in.JobName, &in.JobURL, &in.Status,
		&in.StartedAt, &in.Duration, &in.PreCasesCount)
	if err != nil {
		if database.IsNoRows(err) {
			return nil, dashboard.ErrExecutionNotFound
		}
		return nil, fmt.Errorf("dashboard source: get execution: %w", err)
	}

	// 该 execution 的全部用例（通过 build_uid 关联）
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT i.id, i.case_key, i.case_name, i.attempt_number, i.status,
		       i.start_time, i.end_time, i.duration, i.error_message
		  FROM execution_items i
		  JOIN test_executions e ON e.build_uid = i.build_uid
		 WHERE e.id = $1
		 ORDER BY i.id ASC`, executionID)
	if err != nil {
		return nil, fmt.Errorf("dashboard source: list execution cases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c dashboard.CaseInput
		var errorMsg *string
		if err := rows.Scan(&c.ItemID, &c.CaseKey, &c.CaseName, &c.AttemptNumber,
			&c.Status, &c.StartedAt, &c.EndedAt, &c.Duration, &errorMsg); err != nil {
			return nil, fmt.Errorf("dashboard source: scan case: %w", err)
		}
		if errorMsg != nil {
			c.ErrorMessage = *errorMsg
		}
		in.Cases = append(in.Cases, c)
	}
	return &in, rows.Err()
}

// RunningExecutionsCount 返回当前 running 状态的 execution 数。
func (r *ExecutionSourceReader) RunningExecutionsCount(ctx context.Context) (int, error) {
	var n int
	err := database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT count(*) FROM test_executions WHERE status = 'running'`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("dashboard source: count running executions: %w", err)
	}
	return n, nil
}

// ListRunningExecutionIDs 返回当前 running 状态的 execution ID 列表。
func (r *ExecutionSourceReader) ListRunningExecutionIDs(ctx context.Context) ([]int64, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT id FROM test_executions WHERE status = 'running' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("dashboard source: list running execution ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("dashboard source: scan running execution id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ActivePipelinesCount 返回启用中的流水线数。
func (r *ExecutionSourceReader) ActivePipelinesCount(ctx context.Context) (int, error) {
	var n int
	err := database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT count(*) FROM jenkins_pipelines WHERE status = 'active'`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("dashboard source: count active pipelines: %w", err)
	}
	return n, nil
}
