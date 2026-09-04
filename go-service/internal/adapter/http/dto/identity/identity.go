// Package identity 定义 identity 上下文的 HTTP 请求/响应 DTO。
//
// 字段与 Python 侧一一对应，不得擅自重命名或省略：
//   - 请求 DTO ← service/src/app/auth/schemas.py
//   - 响应 DTO ← service/src/app/auth/constants.py
//
// 这是 ADR-0001 的硬性要求：契约以 Python 代码为基准。
package identity

import (
	"time"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

// ---------------------------------------------------------------------------
// 复用的小结构
// ---------------------------------------------------------------------------

// RoleItem 对应 Python constants.RoleItem。
type RoleItem struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UserMetadata 对应 Python schemas.UserMetadata。
type UserMetadata struct {
	Nickname   *string `json:"nickname"`
	Phone      *string `json:"phone"`
	Avatar     *string `json:"avatar"`
	Department *string `json:"department"`
}

// UserItem 对应 Python constants.UserItem。
type UserItem struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       *string    `json:"email"`
	Status      int        `json:"status"`
	Roles       []RoleItem `json:"roles"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   *time.Time `json:"created_at"`
}

// UserDetailData 对应 Python constants.UserDetailData（继承 UserItem）。
type UserDetailData struct {
	ID          int64         `json:"id"`
	Username    string        `json:"username"`
	Email       *string       `json:"email"`
	Status      int           `json:"status"`
	Roles       []RoleItem    `json:"roles"`
	LastLoginAt *time.Time    `json:"last_login_at"`
	CreatedAt   *time.Time    `json:"created_at"`
	LastLoginIP *string       `json:"last_login_ip"`
	Metadata    *UserMetadata `json:"metadata"`
	UpdatedAt   *time.Time    `json:"updated_at"`
}

// PermissionItem 对应 Python constants.PermissionItem。
type PermissionItem struct {
	ID           int64      `json:"id"`
	ParentID     *int64     `json:"parent_id"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	ResourceType int        `json:"resource_type"`
	Path         *string    `json:"path"`
	Method       *string    `json:"method"`
	CreatedAt    *time.Time `json:"created_at"`
}

// PermissionTreeItem 对应 Python constants.PermissionTreeItem。
type PermissionTreeItem struct {
	ID           int64                `json:"id"`
	ParentID     *int64               `json:"parent_id"`
	Code         string               `json:"code"`
	Name         string               `json:"name"`
	ResourceType int                  `json:"resource_type"`
	Path         *string              `json:"path"`
	Method       *string              `json:"method"`
	CreatedAt    *time.Time           `json:"created_at"`
	Children     []PermissionTreeItem `json:"children"`
}

// RoleDetailData 对应 Python constants.RoleDetailData。
type RoleDetailData struct {
	ID          int64            `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	IsSystem    bool             `json:"is_system"`
	Permissions []PermissionItem `json:"permissions"`
	CreatedAt   *time.Time       `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// 认证接口
// ---------------------------------------------------------------------------

// LoginRequest 对应 POST /login 的请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 对应 POST /register 与 POST /users 的请求体。
type RegisterRequest struct {
	Username string        `json:"username" binding:"required,min=3,max=64"`
	Password string        `json:"password" binding:"required,min=6"`
	Email    string        `json:"email" binding:"required"`
	Metadata *UserMetadata `json:"metadata"`
	RoleIDs  []int64       `json:"role_ids"`
}

// RefreshRequest 对应 POST /login/refresh 的请求体。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginData 对应 Python constants.UserLoginData。
// 在保留 Python 四个原字段的基础上追加 refresh 相关字段（ADR-0005）。
type LoginData struct {
	AccessToken      string    `json:"access_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int64     `json:"expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresIn int64     `json:"refresh_expires_in"`
	User             *UserItem `json:"user"`
}

// ---------------------------------------------------------------------------
// 用户管理接口
// ---------------------------------------------------------------------------

// UserListData 对应 Python constants.UserListData。
// 注意分页字段是 **items**（与 automation 的 rows 不同，不得统一）。
type UserListData struct {
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Items    []UserItem `json:"items"`
}

// UserUpdateRequest 对应 PUT /users/{user_id} 的请求体。
type UserUpdateRequest struct {
	UserID   int64         `json:"user_id" binding:"required"`
	Username *string       `json:"username"`
	Email    *string       `json:"email"`
	Metadata *UserMetadata `json:"metadata"`
}

// UserRoleAssignRequest 对应 PUT /users/{user_id}/roles。
type UserRoleAssignRequest struct {
	UserID  int64   `json:"user_id" binding:"required"`
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}

// UserStatusUpdateRequest 对应 PUT /users/{user_id}/status。
//
// Status 用**指针**是必需的：go-playground/validator 的 required 标签把数值 0
// 当作「未提供」（零值即空），而 status=0（启用）恰恰是最常用的合法取值。
// 改成指针后 required 只判断是否为 nil，语义才正确。
type UserStatusUpdateRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
	Status *int  `json:"status" binding:"required,oneof=0 1"`
}

// ResetPasswordRequest 对应 POST /users/{user_id}/reset-password。
type ResetPasswordRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePasswordRequest 对应 POST /users/me/password。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// BatchStatusRequest 对应 POST /users/batch-status。
// Status 用指针的原因同 UserStatusUpdateRequest：0 是合法取值。
type BatchStatusRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required,min=1"`
	Status  *int    `json:"status" binding:"required,oneof=0 1"`
}

// BatchDeleteRequest 对应 POST /users/batch-delete。
type BatchDeleteRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required,min=1"`
}

// UserListQuery 是 GET /users 的查询参数。
type UserListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Username string `form:"username"`
	Email    string `form:"email"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
	RoleID   *int64 `form:"role_id"`
}

// ---------------------------------------------------------------------------
// 角色与权限接口
// ---------------------------------------------------------------------------

// RoleCreateRequest 对应 POST /roles。
type RoleCreateRequest struct {
	Code          string  `json:"code" binding:"required,min=2,max=64"`
	Name          string  `json:"name" binding:"required,min=2,max=64"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

// RoleUpdateRequest 对应 PUT /roles/{role_id}。
type RoleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// RolePermissionAssignRequest 对应 PUT /roles/{role_id}/permissions。
type RolePermissionAssignRequest struct {
	RoleID        int64   `json:"role_id" binding:"required"`
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}

// PermissionCreateRequest 对应 POST /permissions。
type PermissionCreateRequest struct {
	ParentID     *int64 `json:"parent_id"`
	Code         string `json:"code" binding:"required,min=2,max=128"`
	Name         string `json:"name" binding:"required,min=2,max=64"`
	ResourceType int    `json:"resource_type" binding:"omitempty,oneof=1 2 3"`
	Path         string `json:"path"`
	Method       string `json:"method"`
}

// ---------------------------------------------------------------------------
// 领域对象 → DTO 的映射
// ---------------------------------------------------------------------------

// ToRoleItem 把领域角色转成响应 DTO。
func ToRoleItem(r *identity.Role) RoleItem {
	return RoleItem{
		ID:          r.ID(),
		Code:        r.Code(),
		Name:        r.Name(),
		Description: nullIfEmpty(r.Description()),
	}
}

// ToUserItem 把领域用户转成列表/详情共用的基础 DTO。
func ToUserItem(u *identity.User) UserItem {
	return UserItem{
		ID:          u.ID(),
		Username:    u.Username(),
		Email:       nullIfEmpty(u.Email()),
		Status:      int(u.Status()),
		Roles:       toRoleItems(u.Roles()),
		LastLoginAt: u.LastLoginAt(),
		CreatedAt:   timePtr(u.CreatedAt()),
	}
}

// ToUserDetailData 把领域用户转成详情 DTO。
func ToUserDetailData(u *identity.User) UserDetailData {
	return UserDetailData{
		ID:          u.ID(),
		Username:    u.Username(),
		Email:       nullIfEmpty(u.Email()),
		Status:      int(u.Status()),
		Roles:       toRoleItems(u.Roles()),
		LastLoginAt: u.LastLoginAt(),
		CreatedAt:   timePtr(u.CreatedAt()),
		LastLoginIP: nullIfEmpty(u.LastLoginIP()),
		Metadata:    toMetadata(u.Metadata()),
		UpdatedAt:   timePtr(u.UpdatedAt()),
	}
}

// ToPermissionItem 把领域权限转成响应 DTO。
func ToPermissionItem(p *identity.Permission) PermissionItem {
	return PermissionItem{
		ID:           p.ID(),
		ParentID:     p.ParentID(),
		Code:         p.Code(),
		Name:         p.Name(),
		ResourceType: int(p.ResourceType()),
		Path:         nullIfEmpty(p.Path()),
		Method:       nullIfEmpty(p.Method()),
		CreatedAt:    timePtr(p.CreatedAt()),
	}
}

// ToPermissionTreeItem 把领域权限树节点转成响应 DTO。
func ToPermissionTreeItem(n identity.PermissionNode) PermissionTreeItem {
	children := make([]PermissionTreeItem, 0, len(n.Children))
	for _, c := range n.Children {
		children = append(children, ToPermissionTreeItem(c))
	}
	item := ToPermissionItem(n.Permission)
	return PermissionTreeItem{
		ID:           item.ID,
		ParentID:     item.ParentID,
		Code:         item.Code,
		Name:         item.Name,
		ResourceType: item.ResourceType,
		Path:         item.Path,
		Method:       item.Method,
		CreatedAt:    item.CreatedAt,
		Children:     children,
	}
}

// ToRoleDetailData 把领域角色转成含权限的详情 DTO。
func ToRoleDetailData(r *identity.Role) RoleDetailData {
	perms := make([]PermissionItem, 0, len(r.Permissions()))
	for _, p := range r.Permissions() {
		perms = append(perms, ToPermissionItem(p))
	}
	return RoleDetailData{
		ID:          r.ID(),
		Code:        r.Code(),
		Name:        r.Name(),
		Description: nullIfEmpty(r.Description()),
		IsSystem:    r.IsSystem(),
		Permissions: perms,
		CreatedAt:   timePtr(r.CreatedAt()),
		UpdatedAt:   timePtr(r.UpdatedAt()),
	}
}

func toRoleItems(roles []*identity.Role) []RoleItem {
	items := make([]RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, ToRoleItem(r))
	}
	return items
}

func toMetadata(m identity.UserMetadata) *UserMetadata {
	return &UserMetadata{
		Nickname:   m.Nickname,
		Phone:      m.Phone,
		Avatar:     m.Avatar,
		Department: m.Department,
	}
}

// nullIfEmpty 把空字符串转成 JSON null。
// Python 侧 Optional 字段缺省为 None，前端按 null 处理，不能输出 ""。
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	v := t
	return &v
}
