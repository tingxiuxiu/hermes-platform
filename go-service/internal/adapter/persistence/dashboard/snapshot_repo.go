package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	"github.com/hermes-platform/go-service/internal/domain/dashboard"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// SnapshotRepo 实现 application/dashboard.SnapshotRepository 端口。
type SnapshotRepo struct {
	db *database.Pool
}

// NewSnapshotRepo 构造快照仓储。
func NewSnapshotRepo(db *database.Pool) *SnapshotRepo {
	return &SnapshotRepo{db: db}
}

var _ appdashboard.SnapshotRepository = (*SnapshotRepo)(nil)

// UpsertSummary 幂等写入全局摘要。
func (r *SnapshotRepo) UpsertSummary(ctx context.Context, s dashboard.SummarySnapshot) error {
	_, err := database.Executor(ctx, r.db).Exec(ctx, `
		INSERT INTO automation_dashboard_summary_snapshots (
			snapshot_key, active_jobs_count, total_executions_count, running_executions_count,
			success_cases_7d, failure_cases_7d, skipped_cases_7d, pass_rate_7d,
			avg_execution_duration_7d, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())
		ON CONFLICT (snapshot_key) DO UPDATE SET
			active_jobs_count = EXCLUDED.active_jobs_count,
			total_executions_count = EXCLUDED.total_executions_count,
			running_executions_count = EXCLUDED.running_executions_count,
			success_cases_7d = EXCLUDED.success_cases_7d,
			failure_cases_7d = EXCLUDED.failure_cases_7d,
			skipped_cases_7d = EXCLUDED.skipped_cases_7d,
			pass_rate_7d = EXCLUDED.pass_rate_7d,
			avg_execution_duration_7d = EXCLUDED.avg_execution_duration_7d,
			updated_at = now()`,
		s.SnapshotKey, s.ActiveJobsCount, s.TotalExecutionsCount, s.RunningExecutionsCount,
		s.SuccessCases7d, s.FailureCases7d, s.SkippedCases7d, s.PassRate7d,
		s.AvgExecutionDuration7d,
	)
	if err != nil {
		return fmt.Errorf("dashboard: upsert summary: %w", err)
	}
	return nil
}

// UpsertTrend 幂等写入一条日趋势。
func (r *SnapshotRepo) UpsertTrend(ctx context.Context, t dashboard.TrendSnapshot) error {
	_, err := database.Executor(ctx, r.db).Exec(ctx, `
		INSERT INTO automation_dashboard_daily_trends (
			stat_date, execution_total, success_cases, failure_cases, skipped_cases, running_cases
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (stat_date) DO UPDATE SET
			execution_total = EXCLUDED.execution_total,
			success_cases = EXCLUDED.success_cases,
			failure_cases = EXCLUDED.failure_cases,
			skipped_cases = EXCLUDED.skipped_cases,
			running_cases = EXCLUDED.running_cases`,
		dateOnly(t.StatDate), t.ExecutionTotal, t.SuccessCases,
		t.FailureCases, t.SkippedCases, t.RunningCases,
	)
	if err != nil {
		return fmt.Errorf("dashboard: upsert trend: %w", err)
	}
	return nil
}

// ReplaceExecutionSnapshots 单事务重建某 execution 的快照。
func (r *SnapshotRepo) ReplaceExecutionSnapshots(
	ctx context.Context,
	exec dashboard.ExecutionSnapshot,
	cases []dashboard.CaseSnapshot,
) error {
	return r.db.WithTx(ctx, func(ctx context.Context) error {
		tx := database.Executor(ctx, r.db)

		// 先删该 execution 的 case 快照（全量重建，避免残留失效行）
		if _, err := tx.Exec(ctx, `
			DELETE FROM automation_dashboard_execution_item_snapshots
			 WHERE execution_id = $1`, exec.ExecutionID); err != nil {
			return fmt.Errorf("dashboard: clear case snapshots: %w", err)
		}

		// execution 快照 upsert
		names, _ := json.Marshal(exec.RunningCaseNames)
		if _, err := tx.Exec(ctx, `
			INSERT INTO automation_dashboard_execution_snapshots (
				execution_id, build_uid, job_id, job_name, job_url, status, start_time, duration,
				pre_cases_count, completed_cases_count, success_cases_count,
				failure_cases_count, skipped_cases_count, running_cases_count,
				progress_percent, running_case_names, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,now())
			ON CONFLICT (execution_id) DO UPDATE SET
				build_uid = EXCLUDED.build_uid,
				job_id = EXCLUDED.job_id,
				job_name = EXCLUDED.job_name,
				job_url = EXCLUDED.job_url,
				status = EXCLUDED.status,
				start_time = EXCLUDED.start_time,
				duration = EXCLUDED.duration,
				pre_cases_count = EXCLUDED.pre_cases_count,
				completed_cases_count = EXCLUDED.completed_cases_count,
				success_cases_count = EXCLUDED.success_cases_count,
				failure_cases_count = EXCLUDED.failure_cases_count,
				skipped_cases_count = EXCLUDED.skipped_cases_count,
				running_cases_count = EXCLUDED.running_cases_count,
				progress_percent = EXCLUDED.progress_percent,
				running_case_names = EXCLUDED.running_case_names,
				updated_at = now()`,
			exec.ExecutionID, nullIfEmpty(exec.BuildUID), exec.JobID, exec.JobName, exec.JobURL, exec.Status,
			exec.StartedAt, exec.Duration, exec.PreCasesCount, exec.CompletedCasesCount,
			exec.SuccessCasesCount, exec.FailureCasesCount, exec.SkippedCasesCount,
			exec.RunningCasesCount, exec.ProgressPercent, names,
		); err != nil {
			return fmt.Errorf("dashboard: upsert execution snapshot: %w", err)
		}

		// case 快照批量插入
		for _, c := range cases {
			if _, err := tx.Exec(ctx, `
				INSERT INTO automation_dashboard_execution_item_snapshots (
					execution_id, case_id, case_key, case_name, attempt_number,
					status, start_time, end_time, duration, error_message, updated_at
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now())
				ON CONFLICT (execution_id, case_key) DO UPDATE SET
					case_id = EXCLUDED.case_id,
					case_name = EXCLUDED.case_name,
					attempt_number = EXCLUDED.attempt_number,
					status = EXCLUDED.status,
					start_time = EXCLUDED.start_time,
					end_time = EXCLUDED.end_time,
					duration = EXCLUDED.duration,
					error_message = EXCLUDED.error_message,
					updated_at = now()`,
				c.ExecutionID, c.ItemID, c.CaseKey, c.CaseName, c.AttemptNumber,
				c.Status, c.StartedAt, c.EndedAt, c.Duration,
				nullIfEmpty(c.ErrorMessage),
			); err != nil {
				return fmt.Errorf("dashboard: upsert case snapshot: %w", err)
			}
		}
		return nil
	})
}

// GetSummary 读取全局摘要。不存在返回 nil, nil。
func (r *SnapshotRepo) GetSummary(ctx context.Context) (*dashboard.SummarySnapshot, error) {
	s, err := scanSummary(database.Executor(ctx, r.db).QueryRow(ctx, `
		SELECT snapshot_key, active_jobs_count, total_executions_count, running_executions_count,
		       success_cases_7d, failure_cases_7d, skipped_cases_7d, pass_rate_7d,
		       avg_execution_duration_7d, updated_at
		  FROM automation_dashboard_summary_snapshots
		 WHERE snapshot_key = $1`, dashboard.SnapshotKey))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

// GetTrends 返回最近 N 天的日趋势（升序）。
func (r *SnapshotRepo) GetTrends(ctx context.Context, since time.Time) ([]dashboard.TrendSnapshot, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT stat_date, execution_total, success_cases, failure_cases, skipped_cases, running_cases
		  FROM automation_dashboard_daily_trends
		 WHERE stat_date >= $1 ORDER BY stat_date ASC`, dateOnly(since))
	if err != nil {
		return nil, fmt.Errorf("dashboard: get trends: %w", err)
	}
	defer rows.Close()

	var out []dashboard.TrendSnapshot
	for rows.Next() {
		t, err := scanTrend(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

// ListRunningExecutions 返回 running 状态的 execution 快照。
func (r *SnapshotRepo) ListRunningExecutions(ctx context.Context) ([]dashboard.ExecutionSnapshot, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT execution_id, build_uid::text, job_id, job_name, job_url, status, start_time, duration,
		       pre_cases_count, completed_cases_count, success_cases_count,
		       failure_cases_count, skipped_cases_count, running_cases_count,
		       progress_percent, running_case_names, updated_at
		  FROM automation_dashboard_execution_snapshots
		 WHERE status = 'running'
		 ORDER BY start_time DESC`)
	if err != nil {
		return nil, fmt.Errorf("dashboard: list running executions: %w", err)
	}
	defer rows.Close()

	var out []dashboard.ExecutionSnapshot
	for rows.Next() {
		e, err := scanExecutionSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// ListCasesByExecution 返回某 execution 的全部 case 快照。
func (r *SnapshotRepo) ListCasesByExecution(ctx context.Context, executionID int64) ([]dashboard.CaseSnapshot, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT id, execution_id, case_id, case_key, case_name, attempt_number,
		       status, start_time, end_time, duration, error_message, updated_at
		  FROM automation_dashboard_execution_item_snapshots
		 WHERE execution_id = $1
		 ORDER BY case_id ASC`, executionID)
	if err != nil {
		return nil, fmt.Errorf("dashboard: list cases by execution: %w", err)
	}
	defer rows.Close()

	var out []dashboard.CaseSnapshot
	for rows.Next() {
		c, err := scanCaseSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanSummary(row pgx.Row) (*dashboard.SummarySnapshot, error) {
	var (
		key         string
		active      int
		total       int
		running     int
		success     int
		failure     int
		skipped     int
		passRate    float64
		avgDuration *float64
		updatedAt   time.Time
	)
	if err := row.Scan(&key, &active, &total, &running, &success, &failure,
		&skipped, &passRate, &avgDuration, &updatedAt); err != nil {
		return nil, fmt.Errorf("dashboard: scan summary: %w", err)
	}
	return &dashboard.SummarySnapshot{
		SnapshotKey:            key,
		ActiveJobsCount:        active,
		TotalExecutionsCount:   total,
		RunningExecutionsCount: running,
		SuccessCases7d:         success,
		FailureCases7d:         failure,
		SkippedCases7d:         skipped,
		PassRate7d:             passRate,
		AvgExecutionDuration7d: avgDuration,
		UpdatedAt:              updatedAt,
	}, nil
}

func scanTrend(row pgx.Row) (*dashboard.TrendSnapshot, error) {
	var (
		statDate  time.Time
		execTotal int
		success   int
		failure   int
		skipped   int
		running   int
	)
	if err := row.Scan(&statDate, &execTotal, &success, &failure, &skipped, &running); err != nil {
		return nil, fmt.Errorf("dashboard: scan trend: %w", err)
	}
	return &dashboard.TrendSnapshot{
		StatDate:       statDate,
		ExecutionTotal: execTotal,
		SuccessCases:   success,
		FailureCases:   failure,
		SkippedCases:   skipped,
		RunningCases:   running,
	}, nil
}

func scanExecutionSnapshot(row pgx.Row) (*dashboard.ExecutionSnapshot, error) {
	var (
		executionID     int64
		buildUID        *string
		jobID           int64
		jobName, jobURL string
		status          string
		startedAt       time.Time
		duration        *float64
		preCases        int
		completed       int
		success         int
		failure         int
		skipped         int
		running         int
		progress        float64
		namesRaw        []byte
		updatedAt       time.Time
	)
	if err := row.Scan(&executionID, &buildUID, &jobID, &jobName, &jobURL, &status,
		&startedAt, &duration, &preCases, &completed, &success, &failure,
		&skipped, &running, &progress, &namesRaw, &updatedAt); err != nil {
		return nil, fmt.Errorf("dashboard: scan execution snapshot: %w", err)
	}

	var names []string
	if len(namesRaw) > 0 {
		_ = json.Unmarshal(namesRaw, &names)
	}

	uid := ""
	if buildUID != nil {
		uid = *buildUID
	}

	return &dashboard.ExecutionSnapshot{
		ExecutionID:         executionID,
		BuildUID:            uid,
		JobID:               jobID,
		JobName:             jobName,
		JobURL:              jobURL,
		Status:              status,
		StartedAt:           startedAt,
		Duration:            duration,
		PreCasesCount:       preCases,
		CompletedCasesCount: completed,
		SuccessCasesCount:   success,
		FailureCasesCount:   failure,
		SkippedCasesCount:   skipped,
		RunningCasesCount:   running,
		ProgressPercent:     progress,
		RunningCaseNames:    names,
		UpdatedAt:           updatedAt,
	}, nil
}

func scanCaseSnapshot(row pgx.Row) (*dashboard.CaseSnapshot, error) {
	var (
		id            int64
		executionID   int64
		itemID        int64
		caseKey       string
		caseName      string
		attemptNumber int
		status        string
		startedAt     *time.Time
		endedAt       *time.Time
		duration      *float64
		errorMsg      *string
		updatedAt     time.Time
	)
	if err := row.Scan(&id, &executionID, &itemID, &caseKey, &caseName,
		&attemptNumber, &status, &startedAt, &endedAt, &duration,
		&errorMsg, &updatedAt); err != nil {
		return nil, fmt.Errorf("dashboard: scan case snapshot: %w", err)
	}

	msg := ""
	if errorMsg != nil {
		msg = *errorMsg
	}

	return &dashboard.CaseSnapshot{
		ID:            id,
		ExecutionID:   executionID,
		ItemID:        itemID,
		CaseKey:       caseKey,
		CaseName:      caseName,
		AttemptNumber: attemptNumber,
		Status:        status,
		StartedAt:     startedAt,
		EndedAt:       endedAt,
		Duration:      duration,
		ErrorMessage:  msg,
		UpdatedAt:     updatedAt,
	}, nil
}
