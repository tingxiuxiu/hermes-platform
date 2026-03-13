package api

import (
	"strconv"

	"com.hermes.platform/internal/services"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	tokenService services.APITokenService
}

func NewTokenHandler(tokenService services.APITokenService) *TokenHandler {
	return &TokenHandler{tokenService: tokenService}
}

type CreateTokenRequest struct {
	Name string `json:"name" binding:"required"`
}

type TokenResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	IsRevoked bool   `json:"is_revoked"`
	CreatedAt string `json:"created_at"`
}

type TokenListItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	IsRevoked bool   `json:"is_revoked"`
	CreatedAt string `json:"created_at"`
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

func (h *TokenHandler) CreateToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	token, err := h.tokenService.CreateToken(userID.(uint), req.Name)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, TokenResponse{
		ID:        token.ID,
		Name:      token.Name,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt.Format("2006-01-02 15:04:05"),
		IsRevoked: token.IsRevoked,
		CreatedAt: token.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (h *TokenHandler) ListTokens(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	tokens, err := h.tokenService.GetTokensByUser(userID.(uint))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	var tokenList []TokenListItem
	for _, token := range tokens {
		tokenList = append(tokenList, TokenListItem{
			ID:        token.ID,
			Name:      token.Name,
			Token:     maskToken(token.Token),
			ExpiresAt: token.ExpiresAt.Format("2006-01-02 15:04:05"),
			IsRevoked: token.IsRevoked,
			CreatedAt: token.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	Success(c, gin.H{"tokens": tokenList})
}

func (h *TokenHandler) DeleteToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "User not authenticated")
		return
	}

	tokenIDStr := c.Param("id")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		BadRequest(c, "Invalid token ID")
		return
	}

	err = h.tokenService.RevokeToken(uint(tokenID), userID.(uint))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Token revoked successfully"})
}
