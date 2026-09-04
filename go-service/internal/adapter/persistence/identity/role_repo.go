package identity

import (
	"context"
	"fmt"
	"time"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// RoleRepo 实现 application/identity.RoleRepository 端口。
type RoleRepo struct {
	db *database.Pool
}

// NewRoleRepo 构造角色仓储。
func NewRoleRepo(db *database.Pool) *RoleRepo {
	return &RoleRepo{db: db}
}

var _ appidentity.RoleRepository = (*RoleRepo)(nil)

const roleColumns = `id, code, name, description, is_system, created_at, updated_at`

// Create 写入角色并返回生成的 ID。
func (r *RoleRepo) Create(ctx context.Context, role *identity.Role) (int64, error) {
	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO roles (code, name, description, is_system)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		role.Code(), role.Name(), nullIfEmpty(role.Description()), role.IsSystem(),
	).Scan(&id)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return 0, identity.ErrRoleCodeTakenWithCode(role.Code())
		}
		return 0, fmt.Errorf("identity: create role: %w", err)
	}
	return id, nil
}

// GetByID 读取角色及其权限。
func (r *RoleRepo) GetByID(ctx context.Context, id int64) (*identity.Role, error) {
	return r.getOne(ctx, `WHERE id = $1`, id)
}

// GetByCode 按标识读取角色。
func (r *RoleRepo) GetByCode(ctx context.Context, code string) (*identity.Role, error) {
	return r.getOne(ctx, `WHERE code = $1`, code)
}

func (r *RoleRepo) getOne(ctx context.Context, where string, args ...any) (*identity.Role, error) {
	role, err := scanRole(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+roleColumns+` FROM roles `+where, args...))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, identity.ErrRoleNotFound
		}
		return nil, err
	}

	perms, err := r.permissionsOf(ctx, role.ID())
	if err != nil {
		return nil, err
	}
	role.SetPermissions(perms)
	return role, nil
}

// CodeExists 检查角色标识是否被占用。
func (r *RoleRepo) CodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	return existsBy(r.db, ctx, `roles`, `code = $1`, code, excludeID)
}

// List 返回全部角色（含权限），按 id ASC 排序。
// 一次性加载全部权限后回填，避免 N+1。
func (r *RoleRepo) List(ctx context.Context) ([]*identity.Role, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+roleColumns+` FROM roles ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("identity: list roles: %w", err)
	}
	defer rows.Close()

	var roles []*identity.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("identity: iterate roles: %w", err)
	}
	if len(roles) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID())
	}
	byRole, err := r.permissionsOfRoles(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		role.SetPermissions(byRole[role.ID()])
	}
	return roles, nil
}

// ListByIDs 批量读取角色（不含权限），用于创建用户时绑定初始角色。
func (r *RoleRepo) ListByIDs(ctx context.Context, ids []int64) ([]*identity.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+roleColumns+` FROM roles WHERE id = ANY($1) ORDER BY id ASC`, ids)
	if err != nil {
		return nil, fmt.Errorf("identity: list roles by ids: %w", err)
	}
	defer rows.Close()

	var roles []*identity.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// Update 写回角色名称与描述。
func (r *RoleRepo) Update(ctx context.Context, role *identity.Role) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE roles
		   SET name = $2, description = $3, updated_at = now()
		 WHERE id = $1`,
		role.ID(), role.Name(), nullIfEmpty(role.Description()))
	if err != nil {
		return fmt.Errorf("identity: update role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return identity.ErrRoleNotFound
	}
	return nil
}

// Delete 删除角色。
// 关联的 user_roles / role_permissions 由数据库外键 ON DELETE CASCADE 清理，
// 因此这里不需要（也不应该）手工删除——交给数据库保证原子性。
func (r *RoleRepo) Delete(ctx context.Context, id int64) error {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("identity: delete role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return identity.ErrRoleNotFound
	}
	return nil
}

// SetPermissions 覆盖式设置角色权限，单事务先删后插。
func (r *RoleRepo) SetPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	return r.db.WithTx(ctx, func(ctx context.Context) error {
		tx := database.Executor(ctx, r.db)

		if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
			return fmt.Errorf("identity: clear role permissions: %w", err)
		}
		for _, pid := range permissionIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, p.id FROM permissions p WHERE p.id = $2
				ON CONFLICT DO NOTHING`, roleID, pid); err != nil {
				return fmt.Errorf("identity: bind permission %d: %w", pid, err)
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func (r *RoleRepo) permissionsOf(ctx context.Context, roleID int64) ([]*identity.Permission, error) {
	byRole, err := r.permissionsOfRoles(ctx, []int64{roleID})
	if err != nil {
		return nil, err
	}
	return byRole[roleID], nil
}

// permissionsOfRoles 批量加载多个角色的权限，避免 N+1。
func (r *RoleRepo) permissionsOfRoles(ctx context.Context, roleIDs []int64) (map[int64][]*identity.Permission, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT rp.role_id, p.id, p.parent_id, p.code, p.name,
		       p.resource_type, p.path, p.method, p.created_at
		  FROM role_permissions rp
		  JOIN permissions p ON p.id = rp.permission_id
		 WHERE rp.role_id = ANY($1)
		 ORDER BY p.id ASC`, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("identity: load role permissions: %w", err)
	}
	defer rows.Close()

	out := make(map[int64][]*identity.Permission, len(roleIDs))
	for rows.Next() {
		var roleID, id int64
		var parentID *int64
		var code, name string
		var resourceType int16
		var path, method *string
		var createdAt time.Time

		if err := rows.Scan(&roleID, &id, &parentID, &code, &name,
			&resourceType, &path, &method, &createdAt); err != nil {
			return nil, fmt.Errorf("identity: scan role permission: %w", err)
		}

		p := ""
		if path != nil {
			p = *path
		}
		m := ""
		if method != nil {
			m = *method
		}
		out[roleID] = append(out[roleID], identity.RestorePermission(identity.PermissionSnapshot{
			ID:           id,
			ParentID:     parentID,
			Code:         code,
			Name:         name,
			ResourceType: resourceType,
			Path:         p,
			Method:       m,
			CreatedAt:    createdAt,
		}))
	}
	return out, rows.Err()
}

func scanRole(row interface{ Scan(...any) error }) (*identity.Role, error) {
	var (
		id          int64
		code, name  string
		description *string
		isSystem    bool
		createdAt   time.Time
		updatedAt   time.Time
	)
	if err := row.Scan(&id, &code, &name, &description, &isSystem, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("identity: scan role: %w", err)
	}
	desc := ""
	if description != nil {
		desc = *description
	}
	return identity.RestoreRole(identity.RoleSnapshot{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: desc,
		IsSystem:    isSystem,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}), nil
}
