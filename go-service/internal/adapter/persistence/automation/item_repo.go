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

// ItemRepo 实现 application/automation.ItemRepository 端口。
type ItemRepo struct {
	db *database.Pool
}

// NewItemRepo 构造用例仓储。
func NewItemRepo(db *database.Pool) *ItemRepo {
	return &ItemRepo{db: db}
}

var _ appautomation.ItemRepository = (*ItemRepo)(nil)

const itemColumns = `id, build_uid, case_uid, case_key, case_name, labels,
	attempt_number, is_latest, status, start_time, end_time, duration,
	error_message, error_traceback, created_at, updated_at`

// Create 写入一条用例，返回数据库生成的 ID。
func (r *ItemRepo) Create(ctx context.Context, item *automation.ExecutionItem) (int64, error) {
	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO execution_items (
			build_uid, case_uid, case_key, case_name, labels,
			attempt_number, is_latest, status, start_time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		item.BuildUID(), item.CaseUID(), item.CaseKey(), item.CaseName(),
		nullLabelArray(item.Labels()), item.AttemptNumber(), item.IsLatest(),
		string(item.Status()), item.StartTime(),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("automation: create item: %w", err)
	}
	return id, nil
}

// GetByCaseUID 按 case_uid 读取。
func (r *ItemRepo) GetByCaseUID(ctx context.Context, caseUID uuid.UUID) (*automation.ExecutionItem, error) {
	item, err := scanItem(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+itemColumns+` FROM execution_items WHERE case_uid = $1`, caseUID))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, automation.ErrItemNotFoundByUID(caseUID.String())
		}
		return nil, err
	}
	return item, nil
}

// GetByID 按主键读取。
func (r *ItemRepo) GetByID(ctx context.Context, id int64) (*automation.ExecutionItem, error) {
	item, err := scanItem(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+itemColumns+` FROM execution_items WHERE id = $1`, id))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, automation.ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

// AttemptCount 返回某 build_uid + case_key 下已有的 attempt 数量。
func (r *ItemRepo) AttemptCount(ctx context.Context, buildUID uuid.UUID, caseKey string) (int, error) {
	var n int
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		SELECT count(*) FROM execution_items
		 WHERE build_uid = $1 AND case_key = $2`,
		buildUID, caseKey).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("automation: count attempts: %w", err)
	}
	return n, nil
}

// NextAttemptNumber 原子地取「下一个 attempt 序号」。
//
// 用 pg_advisory_xact_lock 保证同一 (build_uid, case_key) 的序号串行分配：
// 多个 pytest 进程可能同时为同一用例上报新 attempt，若并发插入会撞唯一约束
// `uq_execution_items_attempt (build_uid, case_key, attempt_number)`。
// advisory lock 在事务提交时自动释放，配合调用方的事务边界工作。
func (r *ItemRepo) NextAttemptNumber(ctx context.Context, buildUID uuid.UUID, caseKey string) (int, error) {
	// advisory lock 键：build_uid 的 int64 哈希 + case_key。用稳定哈希避免跨进程不一致。
	lockKey := hashString(buildUID.String() + "\x00" + caseKey)

	// 先取锁：advisory lock 在事务提交/回滚时自动释放。
	if _, err := database.Executor(ctx, r.db).Exec(ctx,
		`SELECT pg_advisory_xact_lock($1)`, lockKey); err != nil {
		return 0, fmt.Errorf("automation: acquire attempt lock: %w", err)
	}

	var next int
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		SELECT count(*) + 1 FROM execution_items
		 WHERE build_uid = $1 AND case_key = $2`,
		buildUID, caseKey).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("automation: next attempt number: %w", err)
	}
	if next < 1 {
		next = 1
	}
	return next, nil
}

// CreateWithNextAttempt 在**单个事务**内分配 attempt 序号并插入用例。
//
// 并发的正确实现：pg_advisory_xact_lock 必须在与 INSERT 相同的事务内持有，
// 锁才会在事务提交前一直生效。若锁与插入分离（NextAttemptNumber + Create
// 两次独立调用），锁在 Exec 返回即释放，并发仍会撞唯一约束
// `uq_execution_items_attempt`。
func (r *ItemRepo) CreateWithNextAttempt(ctx context.Context, buildUID uuid.UUID, item *automation.ExecutionItem) (int64, error) {
	lockKey := hashString(buildUID.String() + "\x00" + item.CaseKey())

	var id int64
	err := r.db.WithTx(ctx, func(ctx context.Context) error {
		exec := database.Executor(ctx, r.db)

		// 1. 取 advisory lock（事务内，提交前一直持有）
		if _, err := exec.Exec(ctx,
			`SELECT pg_advisory_xact_lock($1)`, lockKey); err != nil {
			return fmt.Errorf("automation: acquire attempt lock: %w", err)
		}

		// 2. 计算下一个 attempt 序号
		var next int
		if err := exec.QueryRow(ctx, `
			SELECT count(*) + 1 FROM execution_items
			 WHERE build_uid = $1 AND case_key = $2`,
			buildUID, item.CaseKey()).Scan(&next); err != nil {
			return fmt.Errorf("automation: next attempt number: %w", err)
		}
		if next < 1 {
			next = 1
		}
		item.SetAttemptNumber(next)

		// 3. INSERT（锁仍在事务内，并发安全）
		err := exec.QueryRow(ctx, `
			INSERT INTO execution_items (
				build_uid, case_uid, case_key, case_name, labels,
				attempt_number, is_latest, status, start_time
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			item.BuildUID(), item.CaseUID(), item.CaseKey(), item.CaseName(),
			nullLabelArray(item.Labels()), next, item.IsLatest(),
			string(item.Status()), item.StartTime(),
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("automation: create item in tx: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update 更新一条用例。
func (r *ItemRepo) Update(ctx context.Context, item *automation.ExecutionItem) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE execution_items
		   SET status = $2, end_time = $3, duration = $4,
		       error_message = $5, error_traceback = $6, labels = $7,
		       start_time = $8, updated_at = now()
		 WHERE id = $1`,
		item.ID(), string(item.Status()), item.EndTime(), item.Duration(),
		nullIfEmpty(item.ErrorMessage()), nullIfEmpty(item.ErrorTraceback()),
		nullLabelArray(item.Labels()), item.StartTime(),
	)
	if err != nil {
		return fmt.Errorf("automation: update item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return automation.ErrItemNotFound
	}
	return nil
}

// ListByBuildUID 返回某执行的全部用例，按 attempt_number 升序。
func (r *ItemRepo) ListByBuildUID(ctx context.Context, buildUID uuid.UUID) ([]*automation.ExecutionItem, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+itemColumns+` FROM execution_items
		 WHERE build_uid = $1 ORDER BY attempt_number ASC, id ASC`, buildUID)
	if err != nil {
		return nil, fmt.Errorf("automation: list items: %w", err)
	}
	defer rows.Close()

	var out []*automation.ExecutionItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListByBuildUIDAndCaseKey 返回某执行内某用例的全部 attempt，按 attempt_number 升序。
func (r *ItemRepo) ListByBuildUIDAndCaseKey(ctx context.Context, buildUID uuid.UUID, caseKey string) ([]*automation.ExecutionItem, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+itemColumns+` FROM execution_items
		 WHERE build_uid = $1 AND case_key = $2 ORDER BY attempt_number ASC`,
		buildUID, caseKey)
	if err != nil {
		return nil, fmt.Errorf("automation: list attempts: %w", err)
	}
	defer rows.Close()

	var out []*automation.ExecutionItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListHistoryByCaseKey 返回某用例跨 execution 的历史结果。
// 每个 execution 只取最终 attempt（is_latest = true），按时间倒序。
func (r *ItemRepo) ListHistoryByCaseKey(ctx context.Context, caseKey string) ([]*automation.ExecutionItem, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT `+itemColumns+` FROM execution_items
		 WHERE case_key = $1 AND is_latest = TRUE
		 ORDER BY start_time DESC, id DESC`, caseKey)
	if err != nil {
		return nil, fmt.Errorf("automation: list case history: %w", err)
	}
	defer rows.Close()

	var out []*automation.ExecutionItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListByBuildUIDLatest 返回某执行下每个 case_key 的**最新 attempt**。
//
// ⚠️ 不能依赖 is_latest 列：缺陷 D-02 下所有 attempt 都是 is_latest=true
// （新建不翻转旧行），用它过滤会重复统计重试的旧 attempt。
// 改用窗口函数按 (build_uid, case_key) 取 attempt_number 最大的行。
func (r *ItemRepo) ListByBuildUIDLatest(ctx context.Context, buildUID uuid.UUID) ([]*automation.ExecutionItem, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT `+itemColumns+` FROM (
			SELECT *, ROW_NUMBER() OVER (
				PARTITION BY case_key ORDER BY attempt_number DESC, id DESC
			) AS rn
			FROM execution_items
			WHERE build_uid = $1
		) ranked WHERE rn = 1`, buildUID)
	if err != nil {
		return nil, fmt.Errorf("automation: list latest items: %w", err)
	}
	defer rows.Close()

	var out []*automation.ExecutionItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// CaseUIDOfItem 把 item_id 解析为 case_uid（步骤上报接口用 item_id 定位用例）。
func (r *ItemRepo) CaseUIDOfItem(ctx context.Context, itemID int64) (uuid.UUID, error) {
	var caseUID uuid.UUID
	err := database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT case_uid FROM execution_items WHERE id = $1`, itemID).Scan(&caseUID)
	if err != nil {
		if database.IsNoRows(err) {
			return uuid.Nil, automation.ErrItemNotFound
		}
		return uuid.Nil, fmt.Errorf("automation: resolve item case_uid: %w", err)
	}
	return caseUID, nil
}

func scanItem(row pgx.Row) (*automation.ExecutionItem, error) {
	var (
		id             int64
		buildUID       uuid.UUID
		caseUID        uuid.UUID
		caseKey        string
		caseName       string
		labelsRaw      []string
		attemptNumber  int
		isLatest       bool
		status         string
		startTime      *time.Time
		endTime        *time.Time
		duration       *float64
		errorMsg       *string
		errorTraceback *string
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := row.Scan(&id, &buildUID, &caseUID, &caseKey, &caseName, &labelsRaw,
		&attemptNumber, &isLatest, &status, &startTime, &endTime, &duration,
		&errorMsg, &errorTraceback, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("automation: scan item: %w", err)
	}

	msg := ""
	if errorMsg != nil {
		msg = *errorMsg
	}
	tb := ""
	if errorTraceback != nil {
		tb = *errorTraceback
	}

	item := automation.RestoreItem(automation.ItemSnapshot{
		ID:             id,
		BuildUID:       buildUID,
		CaseUID:        caseUID,
		CaseKey:        caseKey,
		CaseName:       caseName,
		Labels:         labelsRaw,
		AttemptNumber:  attemptNumber,
		IsLatest:       isLatest,
		Status:         status,
		StartTime:      startTime,
		EndTime:        endTime,
		Duration:       duration,
		ErrorMessage:   msg,
		ErrorTraceback: tb,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	})
	return item, nil
}
