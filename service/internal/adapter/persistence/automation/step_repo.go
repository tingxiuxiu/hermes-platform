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

// StepRepo 实现 application/automation.StepRepository 端口。
type StepRepo struct {
	db *database.Pool
}

// NewStepRepo 构造步骤仓储。
func NewStepRepo(db *database.Pool) *StepRepo {
	return &StepRepo{db: db}
}

var _ appautomation.StepRepository = (*StepRepo)(nil)

const stepColumns = `id, case_uid, parent_step_id, parent_step_index, step_index,
	step_path, depth, step_name, status, start_time, end_time, duration,
	created_at, updated_at`

// Upsert 幂等写入/更新一个步骤。
//
// 幂等键：(case_uid, step_path)，对应唯一约束 `uq_execution_item_steps_path`。
// 父步骤的 parent_step_id 在这里解析：同 case_uid 下按 step_path 反查父步骤的主键。
//
// 返回的 created 表示是否为**新插入**（xmax=0 判定，标准 PG 惯用法）：
//   - INSERT 成功：xmax = 0 → created=true
//   - 命中 ON CONFLICT 走 UPDATE：xmax ≠ 0 → created=false
func (r *StepRepo) Upsert(ctx context.Context, step *automation.ExecutionStep) (bool, error) {
	var parentID *int64
	if parentPath, ok := step.StepPath().Parent(); ok {
		err := database.Executor(ctx, r.db).QueryRow(ctx,
			`SELECT id FROM execution_item_steps
			 WHERE case_uid = $1 AND step_path = $2`,
			step.CaseUID(), parentPath.String()).Scan(&parentID)
		if err != nil && !database.IsNoRows(err) {
			return false, fmt.Errorf("automation: resolve parent step: %w", err)
		}
		// 父步骤不存在时 parentID 保持 nil（孤儿，不阻塞写入）
	}

	// 用 xmax 判定新插入：INSERT 返回 xmax=0，ON CONFLICT UPDATE 返回非 0。
	// 这比「先 SELECT 再 INSERT」少一次往返，且避免了竞态。
	var created bool
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO execution_item_steps (
			case_uid, parent_step_id, parent_step_index, step_index, step_path,
			depth, step_name, status, start_time, end_time, duration
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (case_uid, step_path) DO UPDATE SET
			step_name = EXCLUDED.step_name,
			status = EXCLUDED.status,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			duration = EXCLUDED.duration,
			updated_at = now()
		RETURNING (xmax = 0) AS inserted`,
		step.CaseUID(), parentID, step.ParentStepIndex(), step.StepIndex(),
		step.StepPath().String(), step.Depth(), step.StepName(), string(step.Status()),
		step.StartTime(), step.EndTime(), step.Duration(),
	).Scan(&created)
	if err != nil {
		return false, fmt.Errorf("automation: upsert step: %w", err)
	}
	return created, nil
}

// ListByCaseUID 返回某用例的全部步骤，按 step_path 排序。
func (r *StepRepo) ListByCaseUID(ctx context.Context, caseUID uuid.UUID) ([]*automation.ExecutionStep, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+stepColumns+` FROM execution_item_steps
		 WHERE case_uid = $1 ORDER BY step_path ASC`, caseUID)
	if err != nil {
		return nil, fmt.Errorf("automation: list steps: %w", err)
	}
	defer rows.Close()

	var out []*automation.ExecutionStep
	for rows.Next() {
		step, err := scanStep(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, step)
	}
	return out, rows.Err()
}

func scanStep(row pgx.Row) (*automation.ExecutionStep, error) {
	var (
		id              int64
		caseUID         uuid.UUID
		parentStepID    *int64
		parentStepIndex *int
		stepIndex       int
		stepPath        string
		depth           int
		stepName        string
		status          string
		startTime       *time.Time
		endTime         *time.Time
		duration        *float64
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := row.Scan(&id, &caseUID, &parentStepID, &parentStepIndex, &stepIndex,
		&stepPath, &depth, &stepName, &status, &startTime, &endTime, &duration,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("automation: scan step: %w", err)
	}

	return automation.RestoreStep(automation.StepSnapshot{
		ID:              id,
		CaseUID:         caseUID,
		ParentStepID:    parentStepID,
		ParentStepIndex: parentStepIndex,
		StepIndex:       stepIndex,
		StepPath:        stepPath,
		Depth:           depth,
		StepName:        stepName,
		Status:          status,
		StartTime:       startTime,
		EndTime:         endTime,
		Duration:        duration,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}), nil
}
