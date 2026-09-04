package identity

// RoleSeed 描述一个预置角色。
type RoleSeed struct {
	Code        string
	Name        string
	Description string
	IsSystem    bool
}

// PermissionSeed 描述一个预置权限码。
// ResourceType：1=菜单 2=按钮/功能 3=API 接口
type PermissionSeed struct {
	Code         string
	Name         string
	ResourceType int
	Path         string
	Method       string
}

// SystemRoles 是预置角色。is_system=true 的角色禁止删除。
var SystemRoles = []RoleSeed{
	{Code: "admin", Name: "管理员", Description: "内置系统管理员，拥有全部权限", IsSystem: true},
	{Code: "viewer", Name: "只读用户", Description: "内置只读角色，仅可查看", IsSystem: true},
}

// SystemPermissions 是预置权限码，清单见 docs/go-service/03-api-contract.md §8。
var SystemPermissions = []PermissionSeed{
	{Code: "user:list", Name: "用户列表", ResourceType: 3, Path: "/users", Method: "GET"},
	{Code: "user:create", Name: "创建用户", ResourceType: 3, Path: "/users", Method: "POST"},
	{Code: "user:update", Name: "更新用户", ResourceType: 3, Path: "/users/{user_id}", Method: "PUT"},
	{Code: "user:delete", Name: "删除用户", ResourceType: 3, Path: "/users/batch-delete", Method: "POST"},
	{Code: "user:assign-role", Name: "分配角色", ResourceType: 3, Path: "/users/{user_id}/roles", Method: "PUT"},

	{Code: "role:list", Name: "角色列表", ResourceType: 3, Path: "/roles", Method: "GET"},
	{Code: "role:create", Name: "创建角色", ResourceType: 3, Path: "/roles", Method: "POST"},
	{Code: "role:update", Name: "更新角色", ResourceType: 3, Path: "/roles/{role_id}", Method: "PUT"},
	{Code: "role:delete", Name: "删除角色", ResourceType: 3, Path: "/roles/{role_id}", Method: "DELETE"},
	{Code: "role:assign-permission", Name: "分配角色权限", ResourceType: 3, Path: "/roles/{role_id}/permissions", Method: "PUT"},

	{Code: "permission:list", Name: "权限列表", ResourceType: 3, Path: "/permissions", Method: "GET"},
	{Code: "permission:create", Name: "创建权限", ResourceType: 3, Path: "/permissions", Method: "POST"},
	{Code: "permission:delete", Name: "删除权限", ResourceType: 3, Path: "/permissions/{permission_id}", Method: "DELETE"},

	{Code: "automation:read", Name: "自动化查询", ResourceType: 3, Path: "/automation/*", Method: "GET"},
	{Code: "automation:ingest", Name: "自动化上报", ResourceType: 3, Path: "/automation/executions*", Method: "POST"},
	{Code: "dashboard:read", Name: "看板查询", ResourceType: 3, Path: "/automation/dashboard/*", Method: "GET"},
	{Code: "system:dict", Name: "字典管理", ResourceType: 3, Path: "/system/dicts*", Method: ""},
}

// ViewerPermissions 是只读角色拥有的权限码子集。
var ViewerPermissions = []string{
	"user:list", "role:list", "permission:list", "automation:read", "dashboard:read",
}
