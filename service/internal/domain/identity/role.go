package identity

import (
	"strings"
	"time"
)

// Role 是角色实体。
type Role struct {
	id          int64
	code        string
	name        string
	description string
	isSystem    bool
	createdAt   time.Time
	updatedAt   time.Time
	permissions []*Permission
}

// RoleSnapshot 用于从数据库重建角色。
type RoleSnapshot struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []*Permission
}

// NewRole 构造新角色。自建角色一律 isSystem=false：
// 系统内置角色只能通过种子数据创建，不允许运行期新增。
func NewRole(code, name, description string) (*Role, error) {
	r := &Role{}
	if err := r.setCode(code); err != nil {
		return nil, err
	}
	if err := r.setName(name); err != nil {
		return nil, err
	}
	r.description = description
	r.isSystem = false
	return r, nil
}

func RestoreRole(s RoleSnapshot) *Role {
	return &Role{
		id:          s.ID,
		code:        s.Code,
		name:        s.Name,
		description: s.Description,
		isSystem:    s.IsSystem,
		createdAt:   s.CreatedAt,
		updatedAt:   s.UpdatedAt,
		permissions: s.Permissions,
	}
}

func (r *Role) ID() int64                  { return r.id }
func (r *Role) Code() string               { return r.code }
func (r *Role) Name() string               { return r.name }
func (r *Role) Description() string        { return r.description }
func (r *Role) IsSystem() bool             { return r.isSystem }
func (r *Role) CreatedAt() time.Time       { return r.createdAt }
func (r *Role) UpdatedAt() time.Time       { return r.updatedAt }
func (r *Role) Permissions() []*Permission { return r.permissions }

// Rename 修改角色名称。
func (r *Role) Rename(name string) error { return r.setName(name) }

// ChangeDescription 修改角色描述。空字符串表示清空，与 Python 的 `description is not None` 语义一致。
func (r *Role) ChangeDescription(description string) { r.description = description }

// SetPermissions 覆盖式设置权限集合。
func (r *Role) SetPermissions(perms []*Permission) { r.permissions = perms }

// CanDelete 判断是否允许删除。内置系统角色禁止删除。
func (r *Role) CanDelete() bool { return !r.isSystem }

func (r *Role) setCode(code string) error {
	code = strings.TrimSpace(code)
	if textLen(code) < 2 || textLen(code) > 64 {
		return ErrInvalidRoleCode
	}
	r.code = code
	return nil
}

func (r *Role) setName(name string) error {
	name = strings.TrimSpace(name)
	if textLen(name) < 2 || textLen(name) > 64 {
		return ErrInvalidRoleName
	}
	r.name = name
	return nil
}
