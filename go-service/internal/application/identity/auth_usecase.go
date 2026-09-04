package identity

import (
	"context"
	"fmt"

	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/authn"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// AuthResult 是认证类用例的统一返回。
// 刻意不包含任何 HTTP 概念（状态码、响应信封），由 adapter/http 层组装。
type AuthResult struct {
	AccessToken      string
	TokenType        string
	ExpiresIn        int64 // 秒
	RefreshToken     string
	RefreshExpiresIn int64 // 秒
	User             *identity.User
}

// RegisterCommand 是注册请求。
type RegisterCommand struct {
	Username string
	Password string
	Email    string
	Metadata identity.UserMetadata
	RoleIDs  []int64
	ClientIP string
}

// LoginCommand 是登录请求。Username 允许填用户名或邮箱（对齐 Python 的 or_ 查询）。
type LoginCommand struct {
	Username string
	Password string
	ClientIP string
}

// AuthUseCase 编排认证相关用例（任务 T-1.8）。
type AuthUseCase struct {
	users   UserRepository
	roles   RoleRepository
	hasher  PasswordHasher
	issuer  TokenIssuer
	refresh RefreshTokenStore
	revoker TokenRevoker
	guard   LoginGuard
	clock   Clock
}

// NewAuthUseCase 构造认证用例。
func NewAuthUseCase(
	users UserRepository,
	roles RoleRepository,
	hasher PasswordHasher,
	issuer TokenIssuer,
	refresh RefreshTokenStore,
	revoker TokenRevoker,
	guard LoginGuard,
	clock Clock,
) *AuthUseCase {
	return &AuthUseCase{
		users:   users,
		roles:   roles,
		hasher:  hasher,
		issuer:  issuer,
		refresh: refresh,
		revoker: revoker,
		guard:   guard,
		clock:   clock,
	}
}

// Register 注册新用户并直接签发令牌（对齐 Python 的 register_and_login）。
//
// 刻意在同一个调用里完成注册与签发：Python 版把二者拆成「先注册再登录」，
// 会出现跨请求的竞态（注册事务未提交就去登录）。合并后不存在这个窗口。
func (uc *AuthUseCase) Register(ctx context.Context, cmd RegisterCommand) (*AuthResult, error) {
	if len(cmd.Password) < 6 {
		return nil, identity.ErrWeakPassword
	}

	// 用户名与邮箱分别校验，返回可区分的错误
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
		// 不存在的 role_id 静默忽略，与 Python 行为一致
		_ = uc.users.SetRoles(ctx, id, cmd.RoleIDs)
	}

	created, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	created.RecordLogin(uc.clock.Now(), cmd.ClientIP)
	if err := uc.users.Update(ctx, created); err != nil {
		return nil, err
	}

	return uc.issue(ctx, created)
}

// Login 校验凭据并签发令牌。
//
// 安全要点：用户不存在与密码错误返回**同一个**错误，
// 避免攻击者通过响应差异枚举有效账号。
func (uc *AuthUseCase) Login(ctx context.Context, cmd LoginCommand) (*AuthResult, error) {
	if err := uc.guard.Check(ctx, cmd.Username, cmd.ClientIP); err != nil {
		return nil, err
	}

	user, err := uc.users.GetByUsernameOrEmail(ctx, cmd.Username)
	if err != nil && !errors.Is(err, errors.KindNotFound) {
		return nil, err
	}
	if user == nil {
		// 用户不存在也要走一次失败计数，否则锁定机制可被绕过
		_ = uc.guard.RecordFailure(ctx, cmd.Username, cmd.ClientIP)
		return nil, identity.ErrInvalidCredentials
	}

	ok, err := uc.hasher.Verify(cmd.Password, user.PasswordHash())
	if err != nil {
		return nil, fmt.Errorf("identity: verify password: %w", err)
	}
	if !ok {
		_ = uc.guard.RecordFailure(ctx, cmd.Username, cmd.ClientIP)
		return nil, identity.ErrInvalidCredentials
	}

	if user.IsDeleted() {
		return nil, identity.ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, identity.ErrUserInactive
	}

	uc.guard.Clear(ctx, cmd.Username, cmd.ClientIP)

	user.RecordLogin(uc.clock.Now(), cmd.ClientIP)
	if err := uc.users.Update(ctx, user); err != nil {
		return nil, err
	}

	return uc.issue(ctx, user)
}

// Refresh 用刷新令牌换取新的一对令牌，并轮换刷新令牌。
//
// 轮换后旧刷新令牌立即失效；若检测到令牌版本落后（用户已被登出/改密/禁用），
// 拒绝刷新，强制重新登录。
func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	userID, tokenVersion, err := uc.refresh.Consume(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	current, err := uc.revoker.Version(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tokenVersion < current {
		return nil, authn.ErrTokenRevoked
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, identity.ErrUserInactive
	}

	return uc.issue(ctx, user)
}

// Logout 登出：递增令牌版本号，使该用户全部已签发令牌立即失效。
func (uc *AuthUseCase) Logout(ctx context.Context, userID int64) error {
	if _, err := uc.revoker.Revoke(ctx, userID); err != nil {
		return err
	}
	return nil
}

// issue 签发一对令牌。令牌版本取服务端当前值，保证后续可精确吊销。
func (uc *AuthUseCase) issue(ctx context.Context, user *identity.User) (*AuthResult, error) {
	version, err := uc.revoker.Version(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	access, expiresIn, err := uc.issuer.IssueAccessToken(user.ID(), version)
	if err != nil {
		return nil, err
	}

	refresh, refreshExpiresIn, err := uc.refresh.Issue(ctx, user.ID(), version)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:      access,
		TokenType:        "bearer",
		ExpiresIn:        expiresIn,
		RefreshToken:     refresh,
		RefreshExpiresIn: refreshExpiresIn,
		User:             user,
	}, nil
}
