package identity

import (
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
)

// Handlers 聚合 identity 上下文的三组处理器。
//
// 刻意用**具名字段**而不是匿名嵌入：
// UserHandlers / RoleHandlers / PermissionHandlers 都有 List、Create、Update 等
// 同名方法，嵌入会导致选择器歧义（编译期 ambiguous selector）。
// 具名字段让路由表里的调用点一眼能看出归属哪个资源。
type Handlers struct {
	Users       *UserHandlers
	Roles       *RoleHandlers
	Permissions *PermissionHandlers
}

// NewHandlers 一次性构造 identity 的三组处理器。
func NewHandlers(
	users *appidentity.UserUseCase,
	roles *appidentity.RoleUseCase,
	permissions *appidentity.PermissionUseCase,
) *Handlers {
	return &Handlers{
		// role-options 接口需要角色列表，因此 UserHandlers 依赖 RoleUseCase
		Users:       NewUserHandlers(users, roles),
		Roles:       NewRoleHandlers(roles),
		Permissions: NewPermissionHandlers(permissions),
	}
}
