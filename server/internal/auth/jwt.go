package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"com.hermes.platform/internal/config"
)

// 定义 JWT 配置
var (
	jwtSecret     []byte
	jwtExpiration time.Duration
)

// InitJWT 初始化 JWT 配置
func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWT.Secret)
	jwtExpiration = time.Duration(cfg.JWT.Expiration) * time.Hour
}

// JWTClaims 自定义 JWT 声明结构
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID uint, email string, roles []string) (string, error) {
	// 创建声明
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析 JWT 令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
