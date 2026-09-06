package identity

import (
	"context"
	"fmt"

	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// SeedRepo 实现 bootstrap.Seeder，直接操作 SQL。
// 种子写入刻意绕过领域用例：它是基础设施级别的基线数据，
// 不参与业务规则校验，也不应触发领域事件。
type SeedRepo struct {
	db *database.Pool
}

// NewSeedRepo 构造种子仓储。
func NewSeedRepo(db *database.Pool) *SeedRepo {
	return &SeedRepo{db: db}
}

// UpsertRoles 幂等写入角色。已存在时刷新名称与系统标记，避免旧数据残留。
func (r *SeedRepo) UpsertRoles(ctx context.Context, roles []identity.RoleSeed) (map[string]int64, error) {
	exec := database.Executor(ctx, r.db)
	out := make(map[string]int64, len(roles))

	for _, role := range roles {
		var id int64
		err := exec.QueryRow(ctx, `
			INSERT INTO roles (code, name, description, is_system)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (code) DO UPDATE
			   SET name = EXCLUDED.name,
			       description = EXCLUDED.description,
			       is_system = EXCLUDED.is_system,
			       updated_at = now()
			RETURNING id`,
			role.Code, role.Name, nullString(role.Description), role.IsSystem,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("upsert role %q: %w", role.Code, err)
		}
		out[role.Code] = id
	}
	return out, nil
}

// UpsertPermissions 幂等写入权限节点。
func (r *SeedRepo) UpsertPermissions(ctx context.Context, perms []identity.PermissionSeed) (map[string]int64, error) {
	exec := database.Executor(ctx, r.db)
	out := make(map[string]int64, len(perms))

	for _, p := range perms {
		var id int64
		err := exec.QueryRow(ctx, `
			INSERT INTO permissions (parent_id, code, name, resource_type, path, method)
			VALUES (NULL, $1, $2, $3, $4, $5)
			ON CONFLICT (code) DO UPDATE
			   SET name = EXCLUDED.name,
			       resource_type = EXCLUDED.resource_type,
			       path = EXCLUDED.path,
			       method = EXCLUDED.method
			RETURNING id`,
			p.Code, p.Name, p.ResourceType, nullString(p.Path), nullString(p.Method),
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("upsert permission %q: %w", p.Code, err)
		}
		out[p.Code] = id
	}
	return out, nil
}

// SyncRolePermissions 覆盖式对齐角色的权限集合。
// 先删后插放在单事务里，避免出现"权限被清空但未写入"的中间态。
func (r *SeedRepo) SyncRolePermissions(ctx context.Context, roleCode string, permissionCodes []string) error {
	return r.db.WithTx(ctx, func(ctx context.Context) error {
		tx := database.Executor(ctx, r.db)

		var roleID int64
		if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE code = $1`, roleCode).Scan(&roleID); err != nil {
			return fmt.Errorf("lookup role %q: %w", roleCode, err)
		}

		if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
			return fmt.Errorf("clear role permissions: %w", err)
		}

		for _, code := range permissionCodes {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, p.id FROM permissions p WHERE p.code = $2
				ON CONFLICT DO NOTHING`, roleID, code); err != nil {
				return fmt.Errorf("bind permission %q to role %q: %w", code, roleCode, err)
			}
		}
		return nil
	})
}

// AllPermissionCodes 返回库中全部权限码。
func (r *SeedRepo) AllPermissionCodes(ctx context.Context) ([]string, error) {
	rows, err := database.Executor(ctx, r.db).
		Query(ctx, `SELECT code FROM permissions ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list permission codes: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan permission code: %w", err)
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// nullString 把空字符串转成 NULL，与 Python 侧 Optional 字段语义一致。
func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
