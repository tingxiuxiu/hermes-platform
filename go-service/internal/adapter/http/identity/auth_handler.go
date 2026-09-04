// Package identity 承载 identity 上下文的 HTTP 处理器。
//
// 职责边界：只做「HTTP ↔ 用例」的翻译——
// 绑定参数、调用用例、把领域结果转成 DTO 并写入统一响应信封。
// 任何业务规则都不应出现在本包。
package identity

import (
	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/adapter/http/dto/identity"
	"github.com/hermes-platform/go-service/internal/adapter/http/middleware"
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	domainidentity "github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// dto 是本包对「响应/请求 DTO 包」的别名。
// DTO 包与当前包同名（都是 identity），不加别名会让代码难以阅读。
type (
	LoginData       = identity.LoginData
	LoginRequest    = identity.LoginRequest
	RegisterRequest = identity.RegisterRequest
	RefreshRequest  = identity.RefreshRequest
	UserItem        = identity.UserItem
)

// AuthHandlers 聚合认证相关的处理器。
type AuthHandlers struct {
	auth *appidentity.AuthUseCase
}

// NewAuthHandlers 构造认证处理器。
func NewAuthHandlers(auth *appidentity.AuthUseCase) *AuthHandlers {
	return &AuthHandlers{auth: auth}
}

// LoginForm 处理 POST /login/access-token。
//
// 这是 OAuth2 表单登录，供 Swagger 的 Authorize 按钮使用，
// 与 POST /login 的唯一区别是请求体格式（form vs json）。
func (h *AuthHandlers) LoginForm(c *gin.Context) {
	var form struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}
	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, 422, "用户名与密码不能为空")
		return
	}
	h.login(c, form.Username, form.Password)
}

// Login 处理 POST /login（JSON 登录）。
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}
	h.login(c, req.Username, req.Password)
}

// Register 处理 POST /register。注册成功后直接返回令牌（对齐 Python 的 register_and_login）。
func (h *AuthHandlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	result, err := h.auth.Register(c.Request.Context(), appidentity.RegisterCommand{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Metadata: toDomainMetadata(req.Metadata),
		RoleIDs:  req.RoleIDs,
		ClientIP: clientIP(c),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "登录成功", toLoginData(result))
}

// Logout 处理 POST /logout。递增令牌版本号使该用户全部令牌立即失效。
func (h *AuthHandlers) Logout(c *gin.Context) {
	user, ok := middleware.MustUser(c)
	if !ok {
		return
	}
	if err := h.auth.Logout(c.Request.Context(), user.ID()); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "登出成功", nil)
}

// Refresh 处理 POST /login/refresh。用刷新令牌换取新的一对令牌并轮换。
func (h *AuthHandlers) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "刷新成功", toLoginData(result))
}

func (h *AuthHandlers) login(c *gin.Context, username, password string) {
	result, err := h.auth.Login(c.Request.Context(), appidentity.LoginCommand{
		Username: username,
		Password: password,
		ClientIP: clientIP(c),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithMessage(c, "登录成功", toLoginData(result))
}

// clientIP 取客户端 IP。优先信任反向代理透传的 X-Forwarded-For（nginx 已配置）。
func clientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	return c.ClientIP()
}

func toLoginData(r *appidentity.AuthResult) LoginData {
	if r == nil {
		return LoginData{}
	}
	data := LoginData{
		AccessToken:      r.AccessToken,
		TokenType:        r.TokenType,
		ExpiresIn:        r.ExpiresIn,
		RefreshToken:     r.RefreshToken,
		RefreshExpiresIn: r.RefreshExpiresIn,
	}
	if r.User != nil {
		item := identity.ToUserItem(r.User)
		data.User = &item
	}
	return data
}

// toDomainMetadata 把请求 DTO 的扩展信息转成领域值对象。
// 传 nil 时返回零值，表示「不设置任何扩展字段」。
func toDomainMetadata(m *identity.UserMetadata) domainidentity.UserMetadata {
	if m == nil {
		return domainidentity.UserMetadata{}
	}
	return domainidentity.UserMetadata{
		Nickname:   m.Nickname,
		Phone:      m.Phone,
		Avatar:     m.Avatar,
		Department: m.Department,
	}
}
