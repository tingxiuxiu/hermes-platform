package identity

import (
	"context"
	"fmt"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

// CreateUserCommand 是管理员创建用户的请求。
type CreateUserCommand struct {
	Username string
	Password string
	Email    string
	Metadata identity.UserMetadata
	RoleIDs  []int64
}

// UpdateUserCommand 是更新用户基本信息的请求。
// Metadata 为增量合并语义，只有非 nil 字段生效。
type UpdateUserCommand struct {
	Username *string
	Email    *string
	Metadata identity.UserMetadata
}

// UserUseCase 编排用户管理用例（任务 T-1.9）。
//
// 依赖 cache 是为了在用户数据变更后主动失效缓存：
// 否则改完资料后，AuthN 中间件仍会在 TTL（300 秒）内读到旧用户。
// 令牌吊销负责「立即阻止访问」，缓存失效负责「立即看到新数据」，两者职责不同。
type UserUseCase struct {
	users   UserRepository
	roles   RoleRepository
	hasher  PasswordHasher
	revoker TokenRevoker
	cache   UserCache
}

// NewUserUseCase 构造用户用例。
func NewUserUseCase(
	users UserRepository,
	roles RoleRepository,
	hasher PasswordHasher,
	revoker TokenRevoker,
	cache UserCache,
) *UserUseCase {
	return &UserUseCase{
		users:   users,
		roles:   roles,
		hasher:  hasher,
		revoker: revoker,
		cache:   cache,
	}
}

// List 分页查询用户。始终排除软删除用户（由仓储保证）。
func (uc *UserUseCase) List(ctx context.Context, f UserListFilter) ([]*identity.User, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	return uc.users.List(ctx, f)
}

// GetByUsername 按用户名读取用户（首启 seed 用）。
// 不存在时返回 errors.KindNotFound 错误。
func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (*identity.User, error) {
	return uc.users.GetByUsernameOrEmail(ctx, username)
}

// Create 创建用户（不签发令牌，与注册接口区分）。
func (uc *UserUseCase) Create(ctx context.Context, cmd CreateUserCommand) (*identity.User, error) {
	if len(cmd.Password) < 6 {
		return nil, identity.ErrWeakPassword
	}
	if taken, err := uc.users.UsernameExists(ctx, cmd.Username, 0); err != nil {
		return nil, err
	} else if taken {
		return nil, identity.ErrUsernameTaken
	}
	if taken, err := uc.users.EmailExists(ctx, cmd.Email, 0); err != nil {
		return nil, err
	} else if taken {
		return nil, identity.ErrEmailTaken
	}

	hash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("identity: hash password: %w", err)
	}
	user, err := identity.NewUser(cmd.Username, cmd.Email, hash, cmd.Metadata)
	if err != nil {
		return nil, err
	}

	id, err := uc.users.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	if len(cmd.RoleIDs) > 0 {
		_ = uc.users.SetRoles(ctx, id, cmd.RoleIDs)
	}
	return uc.users.GetByID(ctx, id)
}

// Detail 查询用户详情。软删除用户视为不存在。
func (uc *UserUseCase) Detail(ctx context.Context, id int64) (*identity.User, error) {
	return uc.users.GetByID(ctx, id)
}

// Update 更新用户基本信息。
func (uc *UserUseCase) Update(ctx context.Context, id int64, cmd UpdateUserCommand) error {
	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if cmd.Username != nil && *cmd.Username != user.Username() {
		if taken, err := uc.users.UsernameExists(ctx, *cmd.Username, id); err != nil {
			return err
		} else if taken {
			return identity.ErrUsernameTaken
		}
		if err := user.Rename(*cmd.Username); err != nil {
			return err
		}
	}

	if cmd.Email != nil && *cmd.Email != user.Email() {
		if taken, err := uc.users.EmailExists(ctx, *cmd.Email, id); err != nil {
			return err
		} else if taken {
			return identity.ErrEmailTaken
		}
		if err := user.ChangeEmail(*cmd.Email); err != nil {
			return err
		}
	}

	user.UpdateMetadata(cmd.Metadata)
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}
	uc.invalidate(ctx, id)
	return nil
}

// AssignRoles 覆盖式分配角色。不存在的 role_id 静默忽略（对齐 Python）。
func (uc *UserUseCase) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	if _, err := uc.users.GetByID(ctx, userID); err != nil {
		return err
	}
	if err := uc.users.SetRoles(ctx, userID, roleIDs); err != nil {
		return err
	}
	uc.invalidate(ctx, userID)
	return nil
}

// UpdateStatus 启用/禁用账号。变更后该用户的全部令牌立即失效。
func (uc *UserUseCase) UpdateStatus(ctx context.Context, userID int64, status identity.UserStatus) error {
	if status == identity.UserStatusDeleted {
		return identity.ErrInvalidStatusForEnableDisable
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := user.SetStatus(status); err != nil {
		return err
	}
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	uc.invalidate(ctx, userID)
	// 禁用/启用都吊销令牌：避免被停用的账号仍持有有效令牌
	_, err = uc.revoker.Revoke(ctx, userID)
	return err
}

// ResetPassword 管理员重置密码。重置后该用户全部令牌立即失效。
func (uc *UserUseCase) ResetPassword(ctx context.Context, userID int64, newPassword string) error {
	if len(newPassword) < 6 {
		return identity.ErrWeakPassword
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("identity: hash password: %w", err)
	}
	if err := user.ChangePassword(hash); err != nil {
		return err
	}
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	uc.invalidate(ctx, userID)
	_, err = uc.revoker.Revoke(ctx, userID)
	return err
}

// ChangePassword 用户自行修改密码，需校验原密码。
// 成功后该用户的全部令牌立即失效（其他设备也被登出）。
func (uc *UserUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return identity.ErrWeakPassword
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	ok, err := uc.hasher.Verify(oldPassword, user.PasswordHash())
	if err != nil {
		return fmt.Errorf("identity: verify password: %w", err)
	}
	if !ok {
		return identity.ErrOldPasswordMismatch
	}

	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("identity: hash password: %w", err)
	}
	if err := user.ChangePassword(hash); err != nil {
		return err
	}
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	uc.invalidate(ctx, userID)
	_, err = uc.revoker.Revoke(ctx, userID)
	return err
}

// BatchUpdateStatus 批量启用/禁用。返回实际命中的用户数（用于响应文案）。
func (uc *UserUseCase) BatchUpdateStatus(ctx context.Context, ids []int64, status identity.UserStatus) (int64, error) {
	if status == identity.UserStatusDeleted {
		return 0, identity.ErrInvalidStatusForEnableDisable
	}
	count, err := uc.users.BatchUpdateStatus(ctx, ids, status)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		_, _ = uc.revoker.Revoke(ctx, id)
		uc.invalidate(ctx, id)
	}
	return count, nil
}

// BatchSoftDelete 批量软删除。返回实际命中的用户数。
func (uc *UserUseCase) BatchSoftDelete(ctx context.Context, ids []int64) (int64, error) {
	count, err := uc.users.BatchSoftDelete(ctx, ids)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		_, _ = uc.revoker.Revoke(ctx, id)
		uc.invalidate(ctx, id)
	}
	return count, nil
}

// invalidate 清除用户缓存。
// 缓存故障不应影响主流程，因此只忽略错误——
// 令牌吊销已经保证了「立即阻止访问」，缓存只影响「立即看到新数据」。
func (uc *UserUseCase) invalidate(ctx context.Context, userID int64) {
	if uc.cache == nil {
		return
	}
	_ = uc.cache.Delete(ctx, userID)
}
