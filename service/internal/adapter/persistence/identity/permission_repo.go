package identity

import (
	"context"
	"fmt"
	"time"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// PermissionRepo 实现 application/identity.PermissionRepository 端口。
type PermissionRepo struct {
	db *database.Pool
}

// NewPermissionRepo 构造权限仓储。
func NewPermissionRepo(db *database.Pool) *PermissionRepo {
	return &PermissionRepo{db: db}
}

var _ appidentity.PermissionRepository = (*PermissionRepo)(nil)

const permissionColumns = `id, parent_id, code, name, resource_type, path, method, created_at`

// Create 写入权限节点。
func (r *PermissionRepo) Create(ctx context.Context, p *identity.Permission) (int64, error) {
	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO permissions (parent_id, code, name, resource_type, path, method)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		p.ParentID(), p.Code(), p.Name(), int16(p.ResourceType()),
		nullIfEmpty(p.Path()), nullIfEmpty(p.Method()),
	).Scan(&id)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return 0, identity.ErrPermissionCodeTakenWithCode(p.Code())
		}
		if database.IsForeignKeyViolation(err) {
			return 0, identity.ErrPermissionNotFound
		}
		return 0, fmt.Errorf("identity: create permission: %w", err)
	}
	return id, nil
}

func (r *PermissionRepo) GetByID(ctx context.Context, id int64) (*identity.Permission, error) {
	return r.getOne(ctx, `WHERE id = $1`, id)
}

func (r *PermissionRepo) GetByCode(ctx context.Context, code string) (*identity.Permission, error) {
	return r.getOne(ctx, `WHERE code = $1`, code)
}

func (r *PermissionRepo) getOne(ctx context.Context, where string, args ...any) (*identity.Permission, error) {
	p, err := scanPermission(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+permissionColumns+` FROM permissions `+where, args...))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, identity.ErrPermissionNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PermissionRepo) CodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	return existsBy(r.db, ctx, `permissions`, `code = $1`, code, excludeID)
}

// List 返回全部权限，按 id ASC 排序。
// 调用方（PermissionUseCase.Tree）依赖这个顺序保证同层节点的稳定输出。
func (r *PermissionRepo) List(ctx context.Context) ([]*identity.Permission, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+permissionColumns+` FROM permissions ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("identity: list permissions: %w", err)
	}
	defer rows.Close()

	var perms []*identity.Permission
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("identity: iterate permissions: %w", err)
	}
	return perms, nil
}

func (r *PermissionRepo) ListByIDs(ctx context.Context, ids []int64) ([]*identity.Permission, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+permissionColumns+` FROM permissions WHERE id = ANY($1) ORDER BY id ASC`, ids)
	if err != nil {
		return nil, fmt.Errorf("identity: list permissions by ids: %w", err)
	}
	defer rows.Close()

	var perms []*identity.Permission
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

// ListByRoleIDs 批量返回多个角色持有的权限码，供细粒度鉴权使用。
func (r *PermissionRepo) ListByRoleIDs(ctx context.Context, roleIDs []int64) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT DISTINCT p.code
		  FROM role_permissions rp
		  JOIN permissions p ON p.id = rp.permission_id
		 WHERE rp.role_id = ANY($1)
		 ORDER BY p.code ASC`, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("identity: list permission codes by roles: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("identity: scan permission code: %w", err)
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// Delete 删除权限节点。关联的 role_permissions 由外键 CASCADE 清理。
func (r *PermissionRepo) Delete(ctx context.Context, id int64) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx,
		`DELETE FROM permissions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("identity: delete permission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return identity.ErrPermissionNotFound
	}
	return nil
}

func scanPermission(row interface{ Scan(...any) error }) (*identity.Permission, error) {
	var (
		id           int64
		parentID     *int64
		code, name   string
		resourceType int16
		path, method *string
		createdAt    time.Time
	)
	if err := row.Scan(&id, &parentID, &code, &name, &resourceType,
		&path, &method, &createdAt); err != nil {
		return nil, fmt.Errorf("identity: scan permission: %w", err)
	}

	p := ""
	if path != nil {
		p = *path
	}
	m := ""
	if method != nil {
		m = *method
	}
	return identity.RestorePermission(identity.PermissionSnapshot{
		ID:           id,
		ParentID:     parentID,
		Code:         code,
		Name:         name,
		ResourceType: resourceType,
		Path:         p,
		Method:       m,
		CreatedAt:    createdAt,
	}), nil
}

// existsBy 是角色与权限共用的「标识是否被占用」查询。
func existsBy(db *database.Pool, ctx context.Context, table, cond string, value string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM ` + table + ` WHERE ` + cond
	args := []any{value}
	if excludeID > 0 {
		query += ` AND id <> $2`
		args = append(args, excludeID)
	}
	query += `)`

	var exists bool
	if err := database.Executor(ctx, db).QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("identity: check %s exists: %w", table, err)
	}
	return exists, nil
}
