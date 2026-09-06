// Package identity 承载 identity 上下文的应用层：用例编排与出站端口定义（ADR-0004）。
//
// 关键纪律：
//   - 用例返回**领域对象**或领域 DTO，绝不返回 HTTP 响应模型；
//   - 本包只 import domain 与 platform 的无状态基础包，禁止 import 任何 adapter；
//   - Redis / JWT / 密码哈希等基础设施一律以接口形式声明，便于单测替换。
package identity

import (
	"context"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

// ---------------------------------------------------------------------------
// 持久化端口
// ---------------------------------------------------------------------------

// UserListFilter 是用户列表查询条件，对应 GET /users 的查询参数。
type UserListFilter struct {
	Username string
	Email    string
	Status   *identity.UserStatus
	RoleID   *int64
	Page     int
	PageSize int
}

// UserRepository 是用户聚合的持久化端口。
type UserRepository interface {
	// Create 写入新用户，返回数据库生成的 ID。
	Create(ctx context.Context, u *identity.User) (int64, error)

	// GetByID 按主键读取。**必须排除软删除用户**。
	GetByID(ctx context.Context, id int64) (*identity.User, error)

	// GetByUsernameOrEmail 按用户名或邮箱读取，用于登录。
	// 返回 identity.ErrUserNotFound 表示不存在。
	GetByUsernameOrEmail(ctx context.Context, account string) (*identity.User, error)

	// UsernameExists / EmailExists 用于冲突校验。
	// excludeID > 0 时排除该用户自身（更新场景）。
	UsernameExists(ctx context.Context, username string, excludeID int64) (bool, error)
	EmailExists(ctx context.Context, email string, excludeID int64) (bool, error)

	// List 分页查询，**必须排除软删除用户**，按 id DESC 排序。
	List(ctx context.Context, f UserListFilter) ([]*identity.User, int64, error)

	// Update 写回用户的基本信息与状态。
	Update(ctx context.Context, u *identity.User) error

	// SetRoles 覆盖式设置用户角色。不存在的 role_id 静默忽略。
	SetRoles(ctx context.Context, userID int64, roleIDs []int64) error

	// BatchUpdateStatus 批量变更状态，返回实际命中的行数。
	// 必须跳过软删除用户与已删除用户。
	BatchUpdateStatus(ctx context.Context, ids []int64, status identity.UserStatus) (int64, error)

	// BatchSoftDelete 批量软删除，返回实际命中的行数。
	BatchSoftDelete(ctx context.Context, ids []int64) (int64, error)
}

// RoleRepository 是角色聚合的持久化端口。
type RoleRepository interface {
	Create(ctx context.Context, r *identity.Role) (int64, error)
	GetByID(ctx context.Context, id int64) (*identity.Role, error)
	GetByCode(ctx context.Context, code string) (*identity.Role, error)
	CodeExists(ctx context.Context, code string, excludeID int64) (bool, error)

	// List 返回全部角色（含关联权限），按 id ASC 排序。
	List(ctx context.Context) ([]*identity.Role, error)

	// ListByIDs 批量读取，用于创建用户时绑定初始角色。
	ListByIDs(ctx context.Context, ids []int64) ([]*identity.Role, error)

	Update(ctx context.Context, r *identity.Role) error
	Delete(ctx context.Context, id int64) error

	// SetPermissions 覆盖式设置角色权限，单事务。
	SetPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error
}

// PermissionRepository 是权限实体的持久化端口。
type PermissionRepository interface {
	Create(ctx context.Context, p *identity.Permission) (int64, error)
	GetByID(ctx context.Context, id int64) (*identity.Permission, error)
	GetByCode(ctx context.Context, code string) (*identity.Permission, error)
	CodeExists(ctx context.Context, code string, excludeID int64) (bool, error)
	List(ctx context.Context) ([]*identity.Permission, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*identity.Permission, error)

	// ListByRoleIDs 批量查询多个角色持有的权限码，用于细粒度鉴权。
	ListByRoleIDs(ctx context.Context, roleIDs []int64) ([]string, error)

	Delete(ctx context.Context, id int64) error
}

// UserCache 缓存「当前登录用户」的查询结果，TTL 300 秒（对齐 Python）。
//
// 语义约定（实现方必须遵守）：
//   - 未命中、Redis 不可用、数据损坏一律返回 (nil, nil)，不把缓存故障传导为请求失败；
//   - 缓存值**不得包含 password_hash**（安全约束）。
type UserCache interface {
	Get(ctx context.Context, userID int64) (*identity.User, error)
	Set(ctx context.Context, u *identity.User) error
	Delete(ctx context.Context, userID int64) error
}

// ---------------------------------------------------------------------------
// 基础设施端口
// ---------------------------------------------------------------------------

// PasswordHasher 是密码哈希端口，由 platform/password 实现。
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, encoded string) (bool, error)
}

// TokenIssuer 是访问令牌签发端口，由 platform/authn 实现。
type TokenIssuer interface {
	IssueAccessToken(userID, tokenVersion int64) (string, int64, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

// RefreshTokenStore 是刷新令牌存储端口，由 platform/authn 实现。
type RefreshTokenStore interface {
	Issue(ctx context.Context, userID, tokenVersion int64) (string, int64, error)
	Consume(ctx context.Context, raw string) (int64, int64, error)
}

// TokenRevoker 是令牌吊销端口，由 platform/authn 实现。
type TokenRevoker interface {
	Version(ctx context.Context, userID int64) (int64, error)
	Revoke(ctx context.Context, userID int64) (int64, error)
}

// LoginGuard 是登录失败锁定端口，由 platform/authn 实现。
type LoginGuard interface {
	Check(ctx context.Context, username, clientIP string) error
	RecordFailure(ctx context.Context, username, clientIP string) error
	Clear(ctx context.Context, username, clientIP string)
}

// Clock 是时间源端口，由 platform/clock 实现。
type Clock interface {
	Now() time.Time
}
