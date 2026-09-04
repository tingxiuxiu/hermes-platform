package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	appsystem "github.com/hermes-platform/go-service/internal/application/system"
	"github.com/hermes-platform/go-service/internal/domain/system"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// DictRepo 实现 application/system.DictRepository 端口。
type DictRepo struct {
	db *database.Pool
}

// NewDictRepo 构造字典仓储。
func NewDictRepo(db *database.Pool) *DictRepo {
	return &DictRepo{db: db}
}

var _ appsystem.DictRepository = (*DictRepo)(nil)

const dictColumns = `id, dict_type, code, label, dict_name, description,
	category, sort_order, color, status, created_at, updated_at`

// Create 写入一条字典项。
//
// 复合唯一 (dict_type, code) 冲突时映射为领域错误 ErrDictDuplicate，
// 而不是把「唯一约束违反」的 PG 错误泄漏给上层。
func (r *DictRepo) Create(ctx context.Context, d *system.DictEntry) (int64, error) {
	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO sys_dict (
			dict_type, code, label, dict_name, description,
			category, sort_order, color, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		d.DictType(), d.Code(), d.Label(), d.DictName(),
		nullIfEmpty(d.Description()), nullIfEmpty(d.Category()),
		d.SortOrder(), nullIfEmpty(d.Color()), string(d.Status()),
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "uq_sys_dict_type_code") {
			return 0, system.ErrDictDuplicate
		}
		return 0, fmt.Errorf("system: create dict: %w", err)
	}
	return id, nil
}

// GetByID 按主键读取。
func (r *DictRepo) GetByID(ctx context.Context, id int64) (*system.DictEntry, error) {
	d, err := scanDict(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+dictColumns+` FROM sys_dict WHERE id = $1`, id))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, system.ErrDictNotFound
		}
		return nil, err
	}
	return d, nil
}

// List 按条件查询，按 sort_order 升序、id 升序。
func (r *DictRepo) List(ctx context.Context, dictType, category string, onlyActive bool) ([]*system.DictEntry, error) {
	conds := []string{"TRUE"}
	args := make([]any, 0, 3)
	next := 1
	appendArg := func(v any) string {
		args = append(args, v)
		ph := fmt.Sprintf("$%d", next)
		next++
		return ph
	}

	if dictType != "" {
		conds = append(conds, `dict_type = `+appendArg(dictType))
	}
	if category != "" {
		conds = append(conds, `category = `+appendArg(category))
	}
	if onlyActive {
		conds = append(conds, `status = 'active'`)
	}

	where := "WHERE " + conds[0]
	for _, c := range conds[1:] {
		where += " AND " + c
	}

	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+dictColumns+` FROM sys_dict `+where+
			` ORDER BY sort_order ASC, id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("system: list dicts: %w", err)
	}
	defer rows.Close()

	var out []*system.DictEntry
	for rows.Next() {
		d, err := scanDict(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Update 更新一条字典项。
func (r *DictRepo) Update(ctx context.Context, d *system.DictEntry) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE sys_dict
		   SET label = $2, dict_name = $3, description = $4, category = $5,
		       sort_order = $6, color = $7, status = $8, updated_at = now()
		 WHERE id = $1`,
		d.ID(), d.Label(), d.DictName(), nullIfEmpty(d.Description()),
		nullIfEmpty(d.Category()), d.SortOrder(), nullIfEmpty(d.Color()),
		string(d.Status()),
	)
	if err != nil {
		if isUniqueViolation(err, "uq_sys_dict_type_code") {
			return system.ErrDictDuplicate
		}
		return fmt.Errorf("system: update dict: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return system.ErrDictNotFound
	}
	return nil
}

// Delete 删除一条字典项（物理删除，与 Python 侧无软删语义一致）。
func (r *DictRepo) Delete(ctx context.Context, id int64) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx,
		`DELETE FROM sys_dict WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("system: delete dict: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return system.ErrDictNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func scanDict(row pgx.Row) (*system.DictEntry, error) {
	var (
		id          int64
		dictType    string
		code        int
		label       string
		dictName    string
		description *string
		category    *string
		sortOrder   int
		color       *string
		status      string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(&id, &dictType, &code, &label, &dictName,
		&description, &category, &sortOrder, &color, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("system: scan dict: %w", err)
	}

	return system.RestoreDictEntry(system.DictSnapshot{
		ID:          id,
		DictType:    dictType,
		Code:        code,
		Label:       label,
		DictName:    dictName,
		Description: derefStr(description),
		Category:    derefStr(category),
		SortOrder:   sortOrder,
		Color:       derefStr(color),
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}), nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// isUniqueViolation 判断是否为指定约束的唯一性冲突。
// 用 error 字符串匹配而非引入 pgconn 类型，保持仓储层对 pgx 错误类型的解耦。
func isUniqueViolation(err error, constraint string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") && strings.Contains(msg, constraint)
}
