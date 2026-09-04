package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/database"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// UserRepo 实现 application/identity.UserRepository 端口。
type UserRepo struct {
	db *database.Pool
}

// NewUserRepo 构造用户仓储。
func NewUserRepo(db *database.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// 编译期断言：确保实现了端口，避免接口漂移只在运行时暴露。
var _ appidentity.UserRepository = (*UserRepo)(nil)

const userColumns = `id, username, email, password_hash, status, metadata,
	last_login_at, last_login_ip, created_at, updated_at`

// Create 写入新用户并返回生成的 ID。
// 用 RETURNING 显式取回 id，避免额外一次查询（缺陷 D-26 的通用防御）。
func (r *UserRepo) Create(ctx context.Context, u *identity.User) (int64, error) {
	meta, err := json.Marshal(u.Metadata())
	if err != nil {
		return 0, fmt.Errorf("identity: marshal user metadata: %w", err)
	}

	var id int64
	err = database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, status, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		u.Username(), u.Email(), u.PasswordHash(), int16(u.Status()), meta,
	).Scan(&id)
	if err != nil {
		if database.IsUniqueViolation(err) {
			// 由调用方决定是用户名还是邮箱冲突，这里只报告通用冲突
			return 0, identity.ErrUserConflict
		}
		return 0, fmt.Errorf("identity: create user: %w", err)
	}
	return id, nil
}

// GetByID 按主键读取用户及其角色。**排除软删除用户**（status <> 2）。
//
// 这个过滤是必需的：详情、更新、分配角色等接口都基于 GetByID，
// 若返回已删除用户，就会出现「软删除后仍能查看/编辑」的越权窗口。
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*identity.User, error) {
	return r.getOne(ctx, `WHERE u.id = $1 AND u.status <> 2`, id)
}

// GetByUsernameOrEmail 按用户名或邮箱读取（登录用）。
//
// 刻意**不过滤**软删除：Login 用例需要拿到这条记录，
// 才能把「已删除账号登录」显式转成 invalid credentials（401），
// 与「账号不存在」返回完全相同的响应，避免账号枚举。
// 若在这里过滤掉，已删除用户会走「用户不存在」分支——结果仍是 401，
// 但用例就失去了区分能力，将来做风控埋点会缺信息。
func (r *UserRepo) GetByUsernameOrEmail(ctx context.Context, account string) (*identity.User, error) {
	return r.getOne(ctx, `WHERE u.username = $1 OR u.email = $1`, account)
}

func (r *UserRepo) getOne(ctx context.Context, where string, args ...any) (*identity.User, error) {
	user, err := scanUser(database.Executor(ctx, r.db).QueryRow(ctx,
		`SELECT `+userColumns+` FROM users u `+where, args...))
	if err != nil {
		if database.IsNoRows(err) {
			return nil, identity.ErrUserNotFound
		}
		return nil, err
	}

	if err := r.attachRoles(ctx, []*identity.User{user}); err != nil {
		return nil, err
	}
	return user, nil
}

// UsernameExists 检查用户名是否被占用。excludeID > 0 时排除该用户自身。
func (r *UserRepo) UsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	return r.exists(ctx, `username = $1`, username, excludeID)
}

// EmailExists 检查邮箱是否被占用。
func (r *UserRepo) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	return r.exists(ctx, `email = $1`, email, excludeID)
}

func (r *UserRepo) exists(ctx context.Context, cond string, value string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE status <> 2 AND ` + cond
	args := []any{value}
	if excludeID > 0 {
		query += ` AND id <> $2`
		args = append(args, excludeID)
	}
	query += `)`

	var exists bool
	if err := database.Executor(ctx, r.db).QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("identity: check exists: %w", err)
	}
	return exists, nil
}

// List 分页查询用户。始终排除软删除用户，按 id DESC 排序。
func (r *UserRepo) List(ctx context.Context, f appidentity.UserListFilter) ([]*identity.User, int64, error) {
	cond, args, next := buildUserFilter(f)

	var total int64
	countQuery := `SELECT count(*) FROM users u ` + cond
	if err := database.Executor(ctx, r.db).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("identity: count users: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	limitArg := next
	offsetArg := next + 1
	listQuery := fmt.Sprintf(
		`SELECT %s FROM users u %s ORDER BY u.id DESC LIMIT $%d OFFSET $%d`,
		userColumns, cond, limitArg, offsetArg,
	)
	listArgs := append(args, f.PageSize, (f.Page-1)*f.PageSize)

	rows, err := database.Executor(ctx, r.db).Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("identity: list users: %w", err)
	}
	defer rows.Close()

	var users []*identity.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("identity: iterate users: %w", err)
	}

	// 一次性加载全部用户的角色，避免 N+1
	if err := r.attachRoles(ctx, users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// buildUserFilter 拼装动态查询条件。
// 分页参数单独处理，因此这里返回下一个可用的占位符序号。
func buildUserFilter(f appidentity.UserListFilter) (string, []any, int) {
	conds := []string{`u.status <> 2`} // 永远排除软删除
	args := make([]any, 0, 4)
	next := 1

	appendArg := func(v any) string {
		args = append(args, v)
		ph := fmt.Sprintf("$%d", next)
		next++
		return ph
	}

	if f.Username != "" {
		conds = append(conds, `u.username ILIKE `+appendArg("%"+f.Username+"%"))
	}
	if f.Email != "" {
		conds = append(conds, `u.email ILIKE `+appendArg("%"+f.Email+"%"))
	}
	if f.Status != nil {
		conds = append(conds, `u.status = `+appendArg(int16(*f.Status)))
	}
	if f.RoleID != nil {
		// 通过 user_roles 关联筛选，可能产生重复行，用 EXISTS 避免
		conds = append(conds, fmt.Sprintf(
			`EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = %s)`,
			appendArg(*f.RoleID)))
	}

	where := "WHERE " + conds[0]
	for _, c := range conds[1:] {
		where += " AND " + c
	}
	return where, args, next
}

// Update 写回用户的基本信息、状态与密码哈希。
// 用 RETURNING 取回 updated_at，避免额外查询（缺陷 D-26）。
func (r *UserRepo) Update(ctx context.Context, u *identity.User) error {
	meta, err := json.Marshal(u.Metadata())
	if err != nil {
		return fmt.Errorf("identity: marshal user metadata: %w", err)
	}

	var updatedAt any
	err = database.Executor(ctx, r.db).QueryRow(ctx, `
		UPDATE users
		   SET username = $2,
		       email = $3,
		       password_hash = $4,
		       status = $5,
		       metadata = $6,
		       last_login_at = $7,
		       last_login_ip = $8,
		       updated_at = now()
		 WHERE id = $1
		RETURNING updated_at`,
		u.ID(), u.Username(), u.Email(), u.PasswordHash(), int16(u.Status()),
		meta, u.LastLoginAt(), nullIfEmpty(u.LastLoginIP()),
	).Scan(&updatedAt)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return identity.ErrUserConflict
		}
		if database.IsNoRows(err) {
			return identity.ErrUserNotFound
		}
		return fmt.Errorf("identity: update user: %w", err)
	}
	return nil
}

// SetRoles 覆盖式设置用户角色，单事务先删后插。
// 不存在的 role_id 静默忽略（对齐 Python：只插入 roles 表里能查到的）。
func (r *UserRepo) SetRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.db.WithTx(ctx, func(ctx context.Context) error {
		tx := database.Executor(ctx, r.db)

		if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
			return fmt.Errorf("identity: clear user roles: %w", err)
		}
		for _, roleID := range roleIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO user_roles (user_id, role_id)
				SELECT $1, r.id FROM roles r WHERE r.id = $2
				ON CONFLICT DO NOTHING`, userID, roleID); err != nil {
				return fmt.Errorf("identity: bind role %d: %w", roleID, err)
			}
		}
		return nil
	})
}

// BatchUpdateStatus 批量变更状态，返回实际命中的行数。
// 已删除（status=2）的用户被跳过，不会被"复活"。
func (r *UserRepo) BatchUpdateStatus(ctx context.Context, ids []int64, status identity.UserStatus) (int64, error) {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE users
		   SET status = $2, updated_at = now()
		 WHERE id = ANY($1) AND status <> 2`,
		ids, int16(status))
	if err != nil {
		return 0, fmt.Errorf("identity: batch update status: %w", err)
	}
	return tag.RowsAffected(), nil
}

// BatchSoftDelete 批量软删除，返回实际命中的行数。
func (r *UserRepo) BatchSoftDelete(ctx context.Context, ids []int64) (int64, error) {
	tag, err := database.Executor(ctx, r.db).Exec(ctx, `
		UPDATE users
		   SET status = 2, updated_at = now()
		 WHERE id = ANY($1) AND status <> 2`,
		ids)
	if err != nil {
		return 0, fmt.Errorf("identity: batch soft delete: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// attachRoles 批量加载角色后回填到用户实体，避免逐用户查询造成 N+1。
func (r *UserRepo) attachRoles(ctx context.Context, users []*identity.User) error {
	if len(users) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(users))
	index := make(map[int64]*identity.User, len(users))
	for _, u := range users {
		ids = append(ids, u.ID())
		index[u.ID()] = u
	}

	rows, err := database.Executor(ctx, r.db).Query(ctx, `
		SELECT ur.user_id, r.id, r.code, r.name, r.description,
		       r.is_system, r.created_at, r.updated_at
		  FROM user_roles ur
		  JOIN roles r ON r.id = ur.role_id
		 WHERE ur.user_id = ANY($1)
		 ORDER BY r.id ASC`, ids)
	if err != nil {
		return fmt.Errorf("identity: load user roles: %w", err)
	}
	defer rows.Close()

	byUser := make(map[int64][]*identity.Role, len(users))
	for rows.Next() {
		var userID int64
		var code, name string
		var id int64
		var description *string
		var isSystem bool
		var createdAt, updatedAt any

		if err := rows.Scan(&userID, &id, &code, &name, &description,
			&isSystem, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("identity: scan user role: %w", err)
		}
		desc := ""
		if description != nil {
			desc = *description
		}
		byUser[userID] = append(byUser[userID], identity.RestoreRole(identity.RoleSnapshot{
			ID:          id,
			Code:        code,
			Name:        name,
			Description: desc,
			IsSystem:    isSystem,
		}))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("identity: iterate user roles: %w", err)
	}

	for _, u := range users {
		u.SetRoles(byUser[u.ID()])
	}
	return nil
}

// scanUser 从一行扫描出用户实体。
// 接受 pgx.Row 接口，因此 QueryRow 与 Rows 两种来源都能用。
func scanUser(row pgx.Row) (*identity.User, error) {
	var (
		id           int64
		username     string
		email        string
		passwordHash string
		status       int16
		metaRaw      []byte
		lastLoginAt  *time.Time
		lastLoginIP  *string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(&id, &username, &email, &passwordHash, &status,
		&metaRaw, &lastLoginAt, &lastLoginIP, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("identity: scan user: %w", err)
	}

	var meta identity.UserMetadata
	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			// 脏数据不应让整个请求 500，退化为默认值
			meta = identity.UserMetadata{}
		}
	}

	parsedStatus, err := identity.ParseUserStatus(status)
	if err != nil {
		return nil, errors.Wrap(errors.KindInternal, "invalid user status in database", err)
	}

	ip := ""
	if lastLoginIP != nil {
		ip = *lastLoginIP
	}

	return identity.RestoreUser(identity.UserSnapshot{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       int16(parsedStatus),
		Metadata:     meta,
		LastLoginAt:  lastLoginAt,
		LastLoginIP:  ip,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}), nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
