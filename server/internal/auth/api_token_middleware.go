package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type TokenValidator interface {
	ValidateToken(token string) (uint, error)
}

func APITokenAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"code":    401,
				"message": "unauthorized",
				"error":   "Missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"code":    401,
				"message": "unauthorized",
				"error":   "Invalid authorization header format",
			})
			return
		}

		token := parts[1]

		userID, err := validator.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"code":    401,
				"message": "unauthorized",
				"error":   err.Error(),
			})
			return
		}

		c.Set("userID", userID)
		c.Set("token", token)

		c.Next()
	}
}
