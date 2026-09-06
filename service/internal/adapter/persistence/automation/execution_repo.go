package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// ExecutionRepo 实现 application/automation.ExecutionRepository 端口。
type ExecutionRepo struct {
	db *database.Pool
}

// NewExecutionRepo 构造 execution 仓储。
func NewExecutionRepo(db *database.Pool) *ExecutionRepo {
	return &ExecutionRepo{db: db}
}

// 编译期断言：确保实现了端口。
var _ appautomation.ExecutionRepository = (*ExecutionRepo)(nil)

const executionColumns = `id, build_uid, job_name, job_url, project_name, software_name,
	software_version, labels, status, start_time, end_time, duration,
	planned_cases_count, pass_count, failure_count, skipped_count, pass_rate,
	last_heartbeat_at, created_at, updated_at`

// Create 写入一条执行。
//
// 幂等（D-13）：同一 build_uid 已存在时**不插入也不报错**，返回已有记录 ID 与 false。
// 返回的 ID 指向已存在的行，调用方可用它做后续操作。
func (r *ExecutionRepo) Create(ctx context.Context, e *automation.TestExecution) (int64, bool, error) {
	labels := nullLabelArray(e.Labels())

	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO test_executions (
			build_uid, job_name, job_url, project_name, software_name,
			software_version, labels, status, start_time, planned_cases_count,
			last_heartbeat_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $9)
		ON CONFLICT (build_uid) DO NOTHING
		RETURNING id`,
		e.BuildUID(), e.JobName(), e.JobURL(), e.ProjectName(),
		e.SoftwareName(), e.SoftwareVersion(), labels, string(e.Status()),
		e.StartTime(), e.PlannedCasesCount(),
	).Scan(&id)
	if err != nil {
		if database.IsNoRows(err) {
			// build_uid 已存在 → 复用已有记录
			existing, err := r.GetByBuildUID(ctx, e.BuildUID())
			if err != nil {
				return 0, false, err
			}
			return existing.ID(), false, nil
		}
		return 0, false, fmt.Errorf("automation: create execution: %w", err)
	}
	return id, true, nil
}

// GetByBuildUID 按 build_uid 读取。
func (r *ExecutionRepo) GetByBuildUID(ctx context.Context, buildUID uuid.UUID) (*automation.TestExecution, error) {
	e, err := scanExecution(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+executionColumns+` FROM test_executions WHERE build_uid = $1`, buildUID))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, automation.ErrExecutionNotFoundByUID(buildUID.String())
		}
		return nil, err
	}
	return e, nil
}

// GetByID 按主键读取。
func (r *ExecutionRepo) GetByID(ctx context.Context, id int64) (*automation.TestExecution, error) {
	e, err := scanExecution(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+executionColumns+` FROM test_executions WHERE id = $1`, id))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, automation.ErrExecutionNotFound
		}
		return nil, err
	}
	return e, nil
}

// UpdateStatus 结束一条执行并回填计数。
func (r *ExecutionRepo) UpdateStatus(ctx context.Context, e *automation.TestExecution) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE test_executions
		   SET status = $2, end_time = $3, duration = $4,
		       pass_count = $5, failure_count = $6, skipped_count = $7, pass_rate = $8,
		       labels = $9, updated_at = now()
		 WHERE id = $1`,
		e.ID(), string(e.Status()), e.EndTime(), e.Duration(),
		e.PassCount(), e.FailureCount(), e.SkippedCount(), e.PassRate(),
		nullLabelArray(e.Labels()),
	)
	if err != nil {
		return fmt.Errorf("automation: update execution status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return automation.ErrExecutionNotFound
	}
	return nil
}

// List 分页查询执行。
func (r *ExecutionRepo) List(ctx context.Context, f appautomation.ExecutionListFilter) ([]*automation.TestExecution, int64, error) {
	cond, args, next := buildExecutionFilter(f)

	var total int64
	if err := database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT count(*) FROM test_executions `+cond, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("automation: count executions: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	limitPh := fmt.Sprintf("$%d", next)
	offsetPh := fmt.Sprintf("$%d", next+1)
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+executionColumns+` FROM test_executions `+cond+
			` ORDER BY id DESC LIMIT `+limitPh+` OFFSET `+offsetPh,
		append(args, f.PageSize, (f.Page-1)*f.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("automation: list executions: %w", err)
	}
	defer rows.Close()

	var out []*automation.TestExecution
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// LatestByBuildUID 按时间倒序返回同一 build_uid 的所有执行。
func (r *ExecutionRepo) LatestByBuildUID(ctx context.Context, buildUID uuid.UUID) ([]*automation.TestExecution, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+executionColumns+` FROM test_executions
		 WHERE build_uid = $1 ORDER BY id DESC`, buildUID)
	if err != nil {
		return nil, fmt.Errorf("automation: list executions by build_uid: %w", err)
	}
	defer rows.Close()

	var out []*automation.TestExecution
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateHeartbeat 写入 last_heartbeat_at。仅 running 行受影响。
func (r *ExecutionRepo) UpdateHeartbeat(ctx context.Context, e *automation.TestExecution) error {
	at := e.LastHeartbeatAt()
	if at == nil {
		return nil
	}
	_, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE test_executions
		   SET last_heartbeat_at = $2, updated_at = now()
		 WHERE id = $1 AND status = 'running'`,
		e.ID(), *at)
	if err != nil {
		return fmt.Errorf("automation: update execution heartbeat: %w", err)
	}
	return nil
}

// ListStaleRunning 返回心跳早于 cutoff 的 running execution。
func (r *ExecutionRepo) ListStaleRunning(ctx context.Context, cutoff time.Time) ([]*automation.TestExecution, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT `+executionColumns+` FROM test_executions
		 WHERE status = 'running'
		   AND COALESCE(last_heartbeat_at, start_time) < $1
		 ORDER BY id`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("automation: list stale running executions: %w", err)
	}
	defer rows.Close()

	var out []*automation.TestExecution
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func buildExecutionFilter(f appautomation.ExecutionListFilter) (string, []any, int) {
	conds := []string{"TRUE"}
	args := make([]any, 0, 5)
	next := 1
	appendArg := func(v any) string {
		args = append(args, v)
		ph := fmt.Sprintf("$%d", next)
		next++
		return ph
	}

	if f.BuildUID != nil {
		conds = append(conds, `build_uid = `+appendArg(*f.BuildUID))
	}
	if f.JobName != "" {
		conds = append(conds, `job_name ILIKE `+appendArg("%"+f.JobName+"%"))
	}
	if f.Status != nil {
		conds = append(conds, `status = `+appendArg(string(*f.Status)))
	}
	if f.StartFrom != nil {
		conds = append(conds, `start_time >= `+appendArg(*f.StartFrom))
	}
	if f.StartTo != nil {
		conds = append(conds, `start_time <= `+appendArg(*f.StartTo))
	}

	where := "WHERE " + conds[0]
	for _, c := range conds[1:] {
		where += " AND " + c
	}
	return where, args, next
}

func scanExecution(row pgx.Row) (*automation.TestExecution, error) {
	var (
		id               int64
		buildUID         uuid.UUID
		jobName, jobURL  string
		projectName      string
		softwareName     string
		softwareVersion  string
		labelsRaw        []string
		status           string
		startTime        time.Time
		endTime          *time.Time
		duration         *float64
		planned          int
		pass, fail, skip int
		passRate         *float64
		lastHeartbeatAt  *time.Time
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := row.Scan(&id, &buildUID, &jobName, &jobURL, &projectName,
		&softwareName, &softwareVersion, &labelsRaw, &status, &startTime,
		&endTime, &duration, &planned, &pass, &fail, &skip, &passRate,
		&lastHeartbeatAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("automation: scan execution: %w", err)
	}

	e := automation.RestoreExecution(automation.ExecutionSnapshot{
		ID:                id,
		BuildUID:          buildUID,
		JobName:           jobName,
		JobURL:            jobURL,
		ProjectName:       projectName,
		SoftwareName:      softwareName,
		SoftwareVersion:   softwareVersion,
		Labels:            labelsRaw,
		Status:            status,
		StartTime:         startTime,
		EndTime:           endTime,
		Duration:          duration,
		PlannedCasesCount: planned,
		PassCount:         pass,
		FailureCount:      fail,
		SkippedCount:      skip,
		PassRate:          passRate,
		LastHeartbeatAt:   lastHeartbeatAt,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	})
	return e, nil
}
