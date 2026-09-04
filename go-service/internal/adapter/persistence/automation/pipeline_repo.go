package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"

	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// PipelineRepo 实现 application/automation.PipelineRepository 端口。
type PipelineRepo struct {
	db *database.Pool
}

// NewPipelineRepo 构造流水线仓储。
func NewPipelineRepo(db *database.Pool) *PipelineRepo {
	return &PipelineRepo{db: db}
}

var _ appautomation.PipelineRepository = (*PipelineRepo)(nil)

const pipelineColumns = `id, job_name, job_url, status, last_build_number,
	last_build_uid, last_build_status, last_build_timestamp, last_build_duration,
	pipeline_params, sync_at, created_at, updated_at`

// Upsert 按 job_name 幂等写入流水线。返回是否为新插入。
func (r *PipelineRepo) Upsert(ctx context.Context, p *automation.JenkinsPipeline) (bool, error) {
	params := p.PipelineParams()
	var paramsRaw []byte
	if params != nil {
		paramsRaw, _ = json.Marshal(params)
	}

	// xmax=0 判定新插入（见 StepRepo.Upsert 的说明）。
	var created bool
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO jenkins_pipelines (
			job_name, job_url, status, last_build_number, last_build_uid,
			last_build_status, last_build_timestamp, last_build_duration,
			pipeline_params, sync_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (job_name) DO UPDATE SET
			job_url = EXCLUDED.job_url,
			status = EXCLUDED.status,
			last_build_number = EXCLUDED.last_build_number,
			last_build_uid = EXCLUDED.last_build_uid,
			last_build_status = EXCLUDED.last_build_status,
			last_build_timestamp = EXCLUDED.last_build_timestamp,
			last_build_duration = EXCLUDED.last_build_duration,
			pipeline_params = EXCLUDED.pipeline_params,
			sync_at = EXCLUDED.sync_at,
			updated_at = now()
		RETURNING (xmax = 0) AS inserted`,
		p.JobName(), p.JobURL(), string(p.Status()),
		p.LastBuildNumber(), p.LastBuildUID(),
		nullOrString(string(p.LastBuildStatus())), p.LastBuildTimestamp(), p.LastBuildDuration(),
		paramsRaw, p.SyncAt(),
	).Scan(&created)
	if err != nil {
		return false, fmt.Errorf("automation: upsert pipeline: %w", err)
	}
	return created, nil
}

// GetByJobName 按任务名读取。
func (r *PipelineRepo) GetByJobName(ctx context.Context, jobName string) (*automation.JenkinsPipeline, error) {
	p, err := scanPipeline(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+pipelineColumns+` FROM jenkins_pipelines WHERE job_name = $1`, jobName))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, automation.ErrPipelineNotFoundByJobName(jobName)
		}
		return nil, err
	}
	return p, nil
}

// List 分页查询流水线（job_name 模糊 + status / last_build_status 筛选）。
func (r *PipelineRepo) List(ctx context.Context, f appautomation.PipelineListFilter) ([]*automation.JenkinsPipeline, int64, error) {
	// 注意：f 只含 job_name + status + 分页，last_build_status 未在端口字段里。
	// 这里实现与 Python 版 QueryPipelineList 保持一致的必要筛选。
	conds := []string{"TRUE"}
	args := make([]any, 0, 4)
	next := 1
	appendArg := func(v any) string {
		args = append(args, v)
		ph := fmt.Sprintf("$%d", next)
		next++
		return ph
	}

	if f.JobName != "" {
		conds = append(conds, `job_name ILIKE `+appendArg("%"+f.JobName+"%"))
	}
	if f.Status != nil {
		conds = append(conds, `status = `+appendArg(string(*f.Status)))
	}

	where := "WHERE " + conds[0]
	for _, c := range conds[1:] {
		where += " AND " + c
	}

	var total int64
	if err := database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT count(*) FROM jenkins_pipelines `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("automation: count pipelines: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	limitPh := fmt.Sprintf("$%d", next)
	offsetPh := fmt.Sprintf("$%d", next+1)
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+pipelineColumns+` FROM jenkins_pipelines `+where+
			` ORDER BY id DESC LIMIT `+limitPh+` OFFSET `+offsetPh,
		append(args, f.PageSize, (f.Page-1)*f.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("automation: list pipelines: %w", err)
	}
	defer rows.Close()

	var out []*automation.JenkinsPipeline
	for rows.Next() {
		p, err := scanPipeline(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func scanPipeline(row pgx.Row) (*automation.JenkinsPipeline, error) {
	var (
		id                int64
		jobName, jobURL   string
		status            string
		lastBuildNumber   *int
		lastBuildUID      *string
		lastBuildStatus   *string
		lastBuildTS       *time.Time
		lastBuildDuration *float64
		paramsRaw         []byte
		syncAt            *time.Time
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(&id, &jobName, &jobURL, &status, &lastBuildNumber,
		&lastBuildUID, &lastBuildStatus, &lastBuildTS, &lastBuildDuration,
		&paramsRaw, &syncAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("automation: scan pipeline: %w", err)
	}

	buildStatus := ""
	if lastBuildStatus != nil {
		buildStatus = *lastBuildStatus
	}

	// params JSONB → map
	var params map[string]any
	if len(paramsRaw) > 0 {
		_ = json.Unmarshal(paramsRaw, &params)
	}

	return automation.RestorePipeline(automation.PipelineSnapshot{
		ID:                 id,
		JobName:            jobName,
		JobURL:             jobURL,
		Status:             status,
		LastBuildNumber:    lastBuildNumber,
		LastBuildUID:       parseOptionalUUID(lastBuildUID),
		LastBuildStatus:    buildStatus,
		LastBuildTimestamp: lastBuildTS,
		LastBuildDuration:  lastBuildDuration,
		PipelineParams:     params,
		SyncAt:             syncAt,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}), nil
}

func nullOrString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// parseOptionalUUID 把可空的 UUID 字符串解析为 *uuid.UUID。
// 空串或解析失败返回 nil（LastBuildUID 可能为 NULL）。
func parseOptionalUUID(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &u
}
