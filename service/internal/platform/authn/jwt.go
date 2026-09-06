package authn

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hermes-platform/go-service/internal/platform/config"
)

// TokenTypeAccess 是访问令牌的类型标记。
// 写入载荷可防止把刷新令牌当访问令牌用（二者签名算法相同）。
const TokenTypeAccess = "access"

// AccessClaims 是访问令牌的载荷。
type AccessClaims struct {
	// TokenVersion 是签发时用户的令牌版本，用于与 Redis 中的当前版本比对实现吊销。
	TokenVersion int64 `json:"tv"`
	// TokenType 固定为 TokenTypeAccess。
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// UserID 从 subject 中解析用户 ID。
func (c *AccessClaims) UserID() (int64, error) {
	if c.Subject == "" {
		return 0, ErrTokenInvalid
	}
	var id int64
	if _, err := fmt.Sscanf(c.Subject, "%d", &id); err != nil {
		return 0, ErrTokenInvalid
	}
	if id <= 0 {
		return 0, ErrTokenInvalid
	}
	return id, nil
}

// Issuer 签发并校验访问令牌。
type Issuer struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewIssuer 从配置构造 Issuer。
func NewIssuer(cfg config.AuthConfig) *Issuer {
	return &Issuer{
		secret:     []byte(cfg.SecretKey),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTokenTTL,
		refreshTTL: cfg.RefreshTokenTTL,
	}
}

// AccessTTL 返回访问令牌有效期。
func (i *Issuer) AccessTTL() time.Duration { return i.accessTTL }

// RefreshTTL 返回刷新令牌有效期。
func (i *Issuer) RefreshTTL() time.Duration { return i.refreshTTL }

// IssueAccessToken 签发访问令牌。
// 返回令牌字符串与有效期秒数（对应响应体的 expires_in）。
func (i *Issuer) IssueAccessToken(userID, tokenVersion int64) (string, int64, error) {
	now := time.Now().UTC()

	claims := AccessClaims{
		TokenVersion: tokenVersion,
		TokenType:    TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    i.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", 0, fmt.Errorf("authn: sign access token: %w", err)
	}
	return signed, int64(i.accessTTL.Seconds()), nil
}

// ParseAccessToken 校验并解析访问令牌。
// 过期与无效返回不同错误，便于调用方区分处理。
func (i *Issuer) ParseAccessToken(raw string) (*AccessClaims, error) {
	if raw == "" {
		return nil, ErrTokenInvalid
	}

	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return i.secret, nil
	}, jwt.WithIssuer(i.issuer))

	if err != nil {
		// 过期优先于其他错误归类：前端需要据此决定是否刷新
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrTokenInvalid
	}
	if _, err := claims.UserID(); err != nil {
		return nil, err
	}
	return claims, nil
}
