package identity

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/adapter/http/dto/identity"
	"github.com/hermes-platform/go-service/internal/adapter/http/middleware"
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	domainidentity "github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// UserHandlers 聚合用户管理相关的处理器（11 个接口）。
type UserHandlers struct {
	users *appidentity.UserUseCase
	roles *appidentity.RoleUseCase
}

// NewUserHandlers 构造用户处理器。
// 依赖 RoleUseCase 只为 role-options 接口提供角色列表。
func NewUserHandlers(users *appidentity.UserUseCase, roles *appidentity.RoleUseCase) *UserHandlers {
	return &UserHandlers{users: users, roles: roles}
}

// List 处理 GET /users（管理员）。
func (h *UserHandlers) List(c *gin.Context) {
	var q identity.UserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	filter, err := toUserFilter(q)
	if err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	users, total, err := h.users.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]identity.UserItem, 0, len(users))
	for _, u := range users {
		items = append(items, identity.ToUserItem(u))
	}

	response.OKWithMessage(c, "获取用户列表成功", identity.UserListData{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Items:    items,
	})
}

// Create 处理 POST /users（管理员）。
func (h *UserHandlers) Create(c *gin.Context) {
	var req identity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	user, err := h.users.Create(c.Request.Context(), appidentity.CreateUserCommand{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Metadata: toDomainMetadata(req.Metadata),
		RoleIDs:  req.RoleIDs,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "创建用户成功", identity.ToUserItem(user))
}

// RoleOptions 处理 GET /users/role-options。
//
// 缺陷 D-10 修正：Python 版这里错误地返回了**用户列表**，
// 与 RoleOptionResponse（角色列表）的声明矛盾，接口实际上不可用。
// Go 版按接口语义返回角色列表。
func (h *UserHandlers) RoleOptions(c *gin.Context) {
	roles, err := h.roles.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]identity.RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, identity.ToRoleItem(r))
	}
	response.OKWithMessage(c, "获取角色选择列表成功", items)
}

// Detail 处理 GET /users/{user_id}。
func (h *UserHandlers) Detail(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}

	user, err := h.users.Detail(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "获取用户详情成功", identity.ToUserDetailData(user))
}

// Update 处理 PUT /users/{user_id}（管理员）。
func (h *UserHandlers) Update(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}

	var req identity.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	err := h.users.Update(c.Request.Context(), userID, appidentity.UpdateUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Metadata: toDomainMetadata(req.Metadata),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "更新用户信息成功", nil)
}

// AssignRoles 处理 PUT /users/{user_id}/roles（管理员）。
func (h *UserHandlers) AssignRoles(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}

	var req identity.UserRoleAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	if err := h.users.AssignRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "分配用户角色成功", nil)
}

// UpdateStatus 处理 PUT /users/{user_id}/status（管理员）。
func (h *UserHandlers) UpdateStatus(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}

	var req identity.UserStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	if req.Status == nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}
	status, err := domainidentity.ParseUserStatus(int16(*req.Status))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.users.UpdateStatus(c.Request.Context(), userID, status); err != nil {
		response.Error(c, err)
		return
	}

	// 文案与 Python 一致：「用户账号已成功启用」/「用户账号已成功禁用」
	action := "禁用"
	if status == domainidentity.UserStatusActive {
		action = "启用"
	}
	response.OKWithMessage(c, fmt.Sprintf("用户账号已成功%s", action), nil)
}

// ResetPassword 处理 POST /users/{user_id}/reset-password（管理员）。
func (h *UserHandlers) ResetPassword(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}

	var req identity.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	if err := h.users.ResetPassword(c.Request.Context(), userID, req.NewPassword); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "重置密码成功", nil)
}

// ChangeOwnPassword 处理 POST /users/me/password。
//
// 注意：这里用的是 AuthN 注入的当前用户 ID，而不是路径参数。
// 用例内部会**重新从数据库加载**用户以拿到密码哈希——
// 中间件注入的对象可能来自缓存，其密码哈希是空的。
func (h *UserHandlers) ChangeOwnPassword(c *gin.Context) {
	user, ok := middleware.MustUser(c)
	if !ok {
		return
	}

	var req identity.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	err := h.users.ChangePassword(c.Request.Context(), user.ID(), req.OldPassword, req.NewPassword)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "修改密码成功", nil)
}

// BatchUpdateStatus 处理 POST /users/batch-status（管理员）。
func (h *UserHandlers) BatchUpdateStatus(c *gin.Context) {
	var req identity.BatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	if req.Status == nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}
	status, err := domainidentity.ParseUserStatus(int16(*req.Status))
	if err != nil {
		response.Error(c, err)
		return
	}

	count, err := h.users.BatchUpdateStatus(c.Request.Context(), req.UserIDs, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	// 数量取**实际命中**的行数，而非请求传入的长度（对齐 Python 行为）
	response.OKWithMessage(c, fmt.Sprintf("批量更新 %d 个用户状态成功", count), nil)
}

// BatchDelete 处理 POST /users/batch-delete（管理员，软删除）。
func (h *UserHandlers) BatchDelete(c *gin.Context) {
	var req identity.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	count, err := h.users.BatchSoftDelete(c.Request.Context(), req.UserIDs)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, fmt.Sprintf("批量删除 %d 个用户成功", count), nil)
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

func toUserFilter(q identity.UserListQuery) (appidentity.UserListFilter, error) {
	f := appidentity.UserListFilter{
		Username: q.Username,
		Email:    q.Email,
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 10
	}

	if q.Status != nil {
		status, err := domainidentity.ParseUserStatus(int16(*q.Status))
		if err != nil {
			return f, err
		}
		f.Status = &status
	}
	if q.RoleID != nil {
		f.RoleID = q.RoleID
	}
	return f, nil
}
