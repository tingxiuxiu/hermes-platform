package identity

import (
	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/adapter/http/dto/identity"
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	domainidentity "github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// RoleHandlers 聚合角色管理相关的处理器（6 个接口）。
type RoleHandlers struct {
	roles *appidentity.RoleUseCase
}

// NewRoleHandlers 构造角色处理器。
func NewRoleHandlers(roles *appidentity.RoleUseCase) *RoleHandlers {
	return &RoleHandlers{roles: roles}
}

// List 处理 GET /roles（管理员）。
func (h *RoleHandlers) List(c *gin.Context) {
	roles, err := h.roles.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]identity.RoleDetailData, 0, len(roles))
	for _, r := range roles {
		items = append(items, identity.ToRoleDetailData(r))
	}
	response.OKWithMessage(c, "获取角色列表成功", items)
}

// Create 处理 POST /roles（管理员）。
func (h *RoleHandlers) Create(c *gin.Context) {
	var req identity.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	role, err := h.roles.Create(c.Request.Context(), appidentity.CreateRoleCommand{
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "创建角色成功", identity.ToRoleDetailData(role))
}

// Detail 处理 GET /roles/{role_id}（管理员）。
func (h *RoleHandlers) Detail(c *gin.Context) {
	roleID, ok := pathID(c, "role_id")
	if !ok {
		return
	}

	role, err := h.roles.Detail(c.Request.Context(), roleID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "获取角色详情成功", identity.ToRoleDetailData(role))
}

// Update 处理 PUT /roles/{role_id}（管理员）。
func (h *RoleHandlers) Update(c *gin.Context) {
	roleID, ok := pathID(c, "role_id")
	if !ok {
		return
	}

	var req identity.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	role, err := h.roles.Update(c.Request.Context(), roleID, appidentity.UpdateRoleCommand{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "更新角色成功", identity.ToRoleDetailData(role))
}

// Delete 处理 DELETE /roles/{role_id}（管理员）。系统内置角色返回 403。
func (h *RoleHandlers) Delete(c *gin.Context) {
	roleID, ok := pathID(c, "role_id")
	if !ok {
		return
	}

	if err := h.roles.Delete(c.Request.Context(), roleID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "删除角色成功", nil)
}

// AssignPermissions 处理 PUT /roles/{role_id}/permissions（管理员）。
func (h *RoleHandlers) AssignPermissions(c *gin.Context) {
	roleID, ok := pathID(c, "role_id")
	if !ok {
		return
	}

	var req identity.RolePermissionAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	if err := h.roles.AssignPermissions(c.Request.Context(), roleID, req.PermissionIDs); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "权限分配成功", nil)
}

// PermissionHandlers 聚合权限管理相关的处理器（4 个接口）。
type PermissionHandlers struct {
	permissions *appidentity.PermissionUseCase
}

// NewPermissionHandlers 构造权限处理器。
func NewPermissionHandlers(permissions *appidentity.PermissionUseCase) *PermissionHandlers {
	return &PermissionHandlers{permissions: permissions}
}

// Tree 处理 GET /permissions/tree（管理员）。
func (h *PermissionHandlers) Tree(c *gin.Context) {
	nodes, err := h.permissions.Tree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]identity.PermissionTreeItem, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, identity.ToPermissionTreeItem(n))
	}
	response.OKWithMessage(c, "获取权限树成功", items)
}

// List 处理 GET /permissions（管理员）。
func (h *PermissionHandlers) List(c *gin.Context) {
	perms, err := h.permissions.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]identity.PermissionItem, 0, len(perms))
	for _, p := range perms {
		items = append(items, identity.ToPermissionItem(p))
	}
	response.OKWithMessage(c, "获取权限列表成功", items)
}

// Create 处理 POST /permissions（管理员）。
func (h *PermissionHandlers) Create(c *gin.Context) {
	var req identity.PermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	resourceType := domainidentity.ResourceTypeAPI
	if req.ResourceType != 0 {
		resourceType = domainidentity.ResourceType(req.ResourceType)
	}

	perm, err := h.permissions.Create(c.Request.Context(), appidentity.CreatePermissionCommand{
		ParentID:     req.ParentID,
		Code:         req.Code,
		Name:         req.Name,
		ResourceType: resourceType,
		Path:         req.Path,
		Method:       req.Method,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "创建权限节点成功", identity.ToPermissionItem(perm))
}

// Delete 处理 DELETE /permissions/{permission_id}（管理员）。
func (h *PermissionHandlers) Delete(c *gin.Context) {
	permissionID, ok := pathID(c, "permission_id")
	if !ok {
		return
	}

	if err := h.permissions.Delete(c.Request.Context(), permissionID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "删除权限节点成功", nil)
}
