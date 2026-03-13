package api

import (
	"com.hermes.platform/internal/crypto"
	"com.hermes.platform/internal/services"
	"errors"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password"`
	EncryptedPwd  string `json:"encrypted_pwd"`
}

func (r *LoginRequest) Validate() error {
	if r.Password == "" && r.EncryptedPwd == "" {
		return errors.New("either password or encrypted_pwd is required")
	}
	if r.Password != "" && len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

// RSAPubKeyResponse RSA 公钥响应
type RSAPubKeyResponse struct {
	PublicKey string `json:"public_key"`
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest 更新个人资料请求结构
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

// ForgotPasswordRequest 忘记密码请求结构
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Created(c, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		BadRequest(c, err.Error())
		return
	}

	password := req.Password

	if req.EncryptedPwd != "" {
		decryptedPwd, err := crypto.Decrypt(req.EncryptedPwd)
		if err != nil {
			Unauthorized(c, "Failed to decrypt password")
			return
		}
		password = decryptedPwd
	}

	token, err := h.authService.Login(req.Email, password)
	if err != nil {
		Unauthorized(c, err.Error())
		return
	}

	Success(c, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

// GetRSAPubKey 获取 RSA 公钥
func (h *AuthHandler) GetRSAPubKey(c *gin.Context) {
	pubKey := crypto.GetPublicKeyPEM()
	if pubKey == "" {
		InternalServerError(c, "Failed to get RSA public key")
		return
	}

	Success(c, RSAPubKeyResponse{
		PublicKey: pubKey,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 从上下文中获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	err := h.authService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Password changed successfully"})
}

// GetProfile 获取用户个人资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从上下文中获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	// 提取用户角色
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	Success(c, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"roles": roles,
		},
	})
}

// UpdateProfile 更新用户个人资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 从上下文中获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	err := h.authService.UpdateUser(userID.(uint), req.Name)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Profile updated successfully"})
}

// ForgotPassword 忘记密码
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 检查用户是否存在
	_, err := h.authService.GetUserByEmail(req.Email)
	if err != nil {
		// 即使用户不存在，也返回成功消息，避免信息泄露
		Success(c, gin.H{"message": "Password reset email has been sent"})
		return
	}

	// 这里应该实现邮件发送逻辑
	// 由于是演示系统，我们只返回成功消息
	Success(c, gin.H{"message": "Password reset email has been sent"})
}
