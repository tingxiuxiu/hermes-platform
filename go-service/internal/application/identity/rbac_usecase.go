package identity

import (
	"context"
	"fmt"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

// RoleUseCase 编排角色管理用例（任务 T-1.10）。
type RoleUseCase struct {
	roles       RoleRepository
	permissions PermissionRepository
}

// NewRoleUseCase 构造角色用例。
func NewRoleUseCase(roles RoleRepository, permissions PermissionRepository) *RoleUseCase {
	return &RoleUseCase{roles: roles, permissions: permissions}
}

// CreateRoleCommand 是创建角色的请求。
type CreateRoleCommand struct {
	Code          string
	Name          string
	Description   string
	PermissionIDs []int64
}

// UpdateRoleCommand 是更新角色的请求。
// Description 为 nil 表示不改；指向空字符串表示清空。
type UpdateRoleCommand struct {
	Name        *string
	Description *string
}

// List 返回全部角色及其权限，按 id ASC 排序。
func (uc *RoleUseCase) List(ctx context.Context) ([]*identity.Role, error) {
	return uc.roles.List(ctx)
}

// Create 创建角色。自建角色一律 isSystem=false。
func (uc *RoleUseCase) Create(ctx context.Context, cmd CreateRoleCommand) (*identity.Role, error) {
	if taken, err := uc.roles.CodeExists(ctx, cmd.Code, 0); err != nil {
		return nil, err
	} else if taken {
		return nil, identity.ErrRoleCodeTakenWithCode(cmd.Code)
	}

	role, err := identity.NewRole(cmd.Code, cmd.Name, cmd.Description)
	if err != nil {
		return nil, err
	}

	id, err := uc.roles.Create(ctx, role)
	if err != nil {
		return nil, err
	}
	if len(cmd.PermissionIDs) > 0 {
		if err := uc.roles.SetPermissions(ctx, id, cmd.PermissionIDs); err != nil {
			return nil, err
		}
	}
	return uc.roles.GetByID(ctx, id)
}

// Detail 查询角色详情（含权限列表）。
func (uc *RoleUseCase) Detail(ctx context.Context, id int64) (*identity.Role, error) {
	return uc.roles.GetByID(ctx, id)
}

// Update 更新角色名称与描述。
func (uc *RoleUseCase) Update(ctx context.Context, id int64, cmd UpdateRoleCommand) (*identity.Role, error) {
	role, err := uc.roles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cmd.Name != nil {
		if err := role.Rename(*cmd.Name); err != nil {
			return nil, err
		}
	}
	if cmd.Description != nil {
		role.ChangeDescription(*cmd.Description)
	}
	if err := uc.roles.Update(ctx, role); err != nil {
		return nil, err
	}
	return uc.roles.GetByID(ctx, id)
}

// Delete 删除角色。内置系统角色（is_system=true）禁止删除。
func (uc *RoleUseCase) Delete(ctx context.Context, id int64) error {
	role, err := uc.roles.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !role.CanDelete() {
		return identity.ErrSystemRoleImmutableWithCode(role.Code())
	}
	return uc.roles.Delete(ctx, id)
}

// AssignPermissions 覆盖式分配角色权限。
func (uc *RoleUseCase) AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	if _, err := uc.roles.GetByID(ctx, roleID); err != nil {
		return err
	}
	return uc.roles.SetPermissions(ctx, roleID, permissionIDs)
}

// PermissionUseCase 编排权限管理用例（任务 T-1.10）。
type PermissionUseCase struct {
	permissions PermissionRepository
}

// NewPermissionUseCase 构造权限用例。
func NewPermissionUseCase(permissions PermissionRepository) *PermissionUseCase {
	return &PermissionUseCase{permissions: permissions}
}

// CreatePermissionCommand 是创建权限节点的请求。
type CreatePermissionCommand struct {
	ParentID     *int64
	Code         string
	Name         string
	ResourceType identity.ResourceType
	Path         string
	Method       string
}

// List 返回平铺的权限列表，按 id ASC 排序。
func (uc *PermissionUseCase) List(ctx context.Context) ([]*identity.Permission, error) {
	return uc.permissions.List(ctx)
}

// Tree 返回树形权限结构。
// 父节点缺失或自引用的节点会作为根节点返回，避免节点在树中丢失。
func (uc *PermissionUseCase) Tree(ctx context.Context) ([]identity.PermissionNode, error) {
	perms, err := uc.permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	return identity.BuildPermissionTree(perms), nil
}

// Create 创建权限节点。
func (uc *PermissionUseCase) Create(ctx context.Context, cmd CreatePermissionCommand) (*identity.Permission, error) {
	if taken, err := uc.permissions.CodeExists(ctx, cmd.Code, 0); err != nil {
		return nil, err
	} else if taken {
		return nil, identity.ErrPermissionCodeTakenWithCode(cmd.Code)
	}

	perm, err := identity.NewPermission(cmd.Code, cmd.Name, cmd.ResourceType, cmd.Path, cmd.Method)
	if err != nil {
		return nil, err
	}
	if cmd.ParentID != nil {
		// 父节点必须存在，否则存进去会变成孤儿节点
		if _, err := uc.permissions.GetByID(ctx, *cmd.ParentID); err != nil {
			return nil, err
		}
		perm.SetParent(cmd.ParentID)
	}

	id, err := uc.permissions.Create(ctx, perm)
	if err != nil {
		return nil, err
	}
	created, err := uc.permissions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Delete 删除权限节点。
// 关联的 role_permissions 由数据库外键 ON DELETE CASCADE 清理。
func (uc *PermissionUseCase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.permissions.GetByID(ctx, id); err != nil {
		return err
	}
	return uc.permissions.Delete(ctx, id)
}

// ListByRoleIDs 返回多个角色持有的全部权限码，供细粒度鉴权中间件使用。
func (uc *PermissionUseCase) ListByRoleIDs(ctx context.Context, roleIDs []int64) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	codes, err := uc.permissions.ListByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("identity: list permission codes: %w", err)
	}
	return codes, nil
}
