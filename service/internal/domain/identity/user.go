package identity

import (
	"regexp"
	"strings"
	"time"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// UserStatus 用户状态。
// 取值与数据库 CHECK 约束 chk_users_status (0,1,2) 一致。
//
// 注意：Python models.py 的注释写的是 0/-1/-2，与 CHECK 约束矛盾（缺陷 D-01），
// 这里以 0/1/2 为准。
type UserStatus int16

const (
	UserStatusActive   UserStatus = 0 // 正常
	UserStatusDisabled UserStatus = 1 // 禁用
	UserStatusDeleted  UserStatus = 2 // 已删除（软删）
)

// Valid 判断状态值是否合法。
func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusDisabled, UserStatusDeleted:
		return true
	default:
		return false
	}
}

// ParseUserStatus 解析状态值，非法值返回校验错误。
func ParseUserStatus(v int16) (UserStatus, error) {
	s := UserStatus(v)
	if !s.Valid() {
		return 0, errors.Errorf(errors.KindValidation, "非法的用户状态: %d", v)
	}
	return s, nil
}

// UserMetadata 用户扩展信息，对应 users.metadata 这个 JSONB 列。
// json tag 与 Python schemas.UserMetadata 完全一致。
type UserMetadata struct {
	Nickname   *string `json:"nickname"`
	Phone      *string `json:"phone"`
	Avatar     *string `json:"avatar"`
	Department *string `json:"department"`
}

// User 是 identity 上下文的聚合根。
//
// passwordHash 刻意**不导出**：响应体、日志、结构体转储都不可能意外泄漏哈希值。
// 持久化层通过 RestoreUser 重建实体，通过 PasswordHash() 读取。
type User struct {
	id           int64
	username     string
	email        string
	passwordHash string
	status       UserStatus
	metadata     UserMetadata
	lastLoginAt  *time.Time
	lastLoginIP  string
	createdAt    time.Time
	updatedAt    time.Time
	roles        []*Role
}

// NewUser 构造新用户（业务入口，注册/创建用户时调用）。
// 密码哈希由调用方（用例层）通过 platform/password 生成后传入，
// domain 层不关心具体哈希算法。
func NewUser(username, email, passwordHash string, meta UserMetadata) (*User, error) {
	u := &User{}
	if err := u.setUsername(username); err != nil {
		return nil, err
	}
	if err := u.setEmail(email); err != nil {
		return nil, err
	}
	if passwordHash == "" {
		return nil, errors.New(errors.KindInternal, "password hash must not be empty")
	}
	u.passwordHash = passwordHash
	u.status = UserStatusActive
	u.metadata = meta
	u.roles = nil
	return u, nil
}

// RestoreUser 从数据库重建实体。仅供持久化适配器使用。
type UserSnapshot struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Status       int16
	Metadata     UserMetadata
	LastLoginAt  *time.Time
	LastLoginIP  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Roles        []*Role
}

func RestoreUser(s UserSnapshot) *User {
	status := UserStatus(s.Status)
	if !status.Valid() {
		status = UserStatusDisabled
	}
	return &User{
		id:           s.ID,
		username:     s.Username,
		email:        s.Email,
		passwordHash: s.PasswordHash,
		status:       status,
		metadata:     s.Metadata,
		lastLoginAt:  s.LastLoginAt,
		lastLoginIP:  s.LastLoginIP,
		createdAt:    s.CreatedAt,
		updatedAt:    s.UpdatedAt,
		roles:        s.Roles,
	}
}

// ---- 读取器 ----

func (u *User) ID() int64               { return u.id }
func (u *User) Username() string        { return u.username }
func (u *User) Email() string           { return u.email }
func (u *User) PasswordHash() string    { return u.passwordHash }
func (u *User) Status() UserStatus      { return u.status }
func (u *User) Metadata() UserMetadata  { return u.metadata }
func (u *User) LastLoginAt() *time.Time { return u.lastLoginAt }
func (u *User) LastLoginIP() string     { return u.lastLoginIP }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) Roles() []*Role          { return u.roles }

// ---- 行为 ----

// IsDeleted 判断是否已软删除。
// 所有查询都必须排除软删除用户（Python 侧的 `status != 2` 条件）。
func (u *User) IsDeleted() bool { return u.status == UserStatusDeleted }

// IsActive 判断是否可正常登录。
func (u *User) IsActive() bool { return u.status == UserStatusActive }

// AdminRoleCodes 是拥有管理员特权的角色标识集合。
// 与 Python dependencies.get_current_active_admin 完全一致。
var AdminRoleCodes = map[string]struct{}{
	"admin":         {},
	"superadmin":    {},
	"administrator": {},
}

// IsAdmin 判断是否为管理员。
// 规则复刻 Python：用户名为 admin，或拥有 admin/superadmin/administrator 任一角色。
func (u *User) IsAdmin() bool {
	if u.username == "admin" {
		return true
	}
	for _, r := range u.roles {
		if _, ok := AdminRoleCodes[strings.ToLower(r.Code())]; ok {
			return true
		}
	}
	return false
}

// HasPermissionCode 判断是否拥有指定权限码。
// 用于细粒度鉴权（本轮默认不启用，见 03-api-contract.md §8）。
func (u *User) HasPermissionCode(code string) bool {
	for _, r := range u.roles {
		for _, p := range r.Permissions() {
			if p.Code() == code {
				return true
			}
		}
	}
	return false
}

// SetRoles 覆盖式设置角色集合。分配角色接口使用。
func (u *User) SetRoles(roles []*Role) { u.roles = roles }

// ChangePassword 设置新的密码哈希。
func (u *User) ChangePassword(hash string) error {
	if hash == "" {
		return errors.New(errors.KindInternal, "password hash must not be empty")
	}
	u.passwordHash = hash
	return nil
}

// Rename 修改用户名，含格式校验。
func (u *User) Rename(username string) error { return u.setUsername(username) }

// ChangeEmail 修改邮箱，含格式校验。
func (u *User) ChangeEmail(email string) error { return u.setEmail(email) }

// UpdateMetadata 增量合并扩展信息。
// 只对非 nil 字段生效，对齐 Python 的 exclude_unset 语义。
func (u *User) UpdateMetadata(patch UserMetadata) {
	if patch.Nickname != nil {
		u.metadata.Nickname = patch.Nickname
	}
	if patch.Phone != nil {
		u.metadata.Phone = patch.Phone
	}
	if patch.Avatar != nil {
		u.metadata.Avatar = patch.Avatar
	}
	if patch.Department != nil {
		u.metadata.Department = patch.Department
	}
}

// SetStatus 变更账号状态。软删除不可逆：已删除的账号不允许再改回正常。
func (u *User) SetStatus(s UserStatus) error {
	if !s.Valid() {
		return errors.Errorf(errors.KindValidation, "非法的用户状态: %d", int16(s))
	}
	if u.status == UserStatusDeleted {
		return errors.New(errors.KindConflict, "已删除的用户无法变更状态")
	}
	u.status = s
	return nil
}

// SoftDelete 软删除账号。
func (u *User) SoftDelete() error { return u.SetStatus(UserStatusDeleted) }

// RecordLogin 记录登录时间与来源 IP。
func (u *User) RecordLogin(at time.Time, ip string) {
	u.lastLoginAt = &at
	u.lastLoginIP = ip
}

// ---- 内部校验 ----

func (u *User) setUsername(username string) error {
	username = strings.TrimSpace(username)
	if textLen(username) < 3 || textLen(username) > 64 {
		return ErrInvalidUsername
	}
	u.username = username
	return nil
}

// emailPattern 是邮箱格式的宽松校验。
// 刻意不做完整 RFC 5322 校验：Python 侧用 Pydantic EmailStr，
// 过于严格的自研正则反而会拒绝合法地址。这里只拦截明显非法的输入。
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func (u *User) setEmail(email string) error {
	email = strings.TrimSpace(email)
	if !ValidEmail(email) {
		return ErrInvalidEmail
	}
	u.email = email
	return nil
}

// ValidEmail 与 setEmail 使用同一套宽松规则（对齐 Python EmailStr 的下限）。
func ValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return email != "" && emailPattern.MatchString(email)
}
