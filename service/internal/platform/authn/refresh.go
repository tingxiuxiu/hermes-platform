package authn

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hermes-platform/go-service/internal/platform/redis"
)

const refreshKeyPrefix = "refresh:"

// RefreshStore 管理刷新令牌。
//
// 刷新令牌是服务端可精确吊销的不透明随机串（ADR-0005）：
//   - 字符串本身不携带任何信息，无法被伪造；
//   - Redis 中只存 sha256 摘要，数据库泄漏也无法还原可用令牌；
//   - 每次刷新都**轮换**：旧令牌立即失效，签发新令牌。
type RefreshStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// refreshPayload 是刷新令牌在 Redis 中的值。
type refreshPayload struct {
	UserID       int64 `json:"uid"`
	TokenVersion int64 `json:"tv"`
}

// NewRefreshStore 构造刷新令牌存储。
func NewRefreshStore(rdb *redis.Client, ttl time.Duration) *RefreshStore {
	return &RefreshStore{rdb: rdb, ttl: ttl}
}

// Issue 签发一个新的刷新令牌。返回原始令牌串与有效期秒数。
// 同一用户可以同时持有多个刷新令牌（多端登录），互不干扰。
func (s *RefreshStore) Issue(ctx context.Context, userID, tokenVersion int64) (string, int64, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", 0, fmt.Errorf("authn: generate refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)

	payload, err := json.Marshal(refreshPayload{UserID: userID, TokenVersion: tokenVersion})
	if err != nil {
		return "", 0, fmt.Errorf("authn: marshal refresh payload: %w", err)
	}

	if err := s.rdb.Set(ctx, refreshKeyPrefix+redis.HashKey(raw), payload, s.ttl).Err(); err != nil {
		return "", 0, fmt.Errorf("authn: store refresh token: %w", err)
	}
	return raw, int64(s.ttl.Seconds()), nil
}

// consumeScript 原子地「取出并删除」刷新令牌。
// 必须用 Lua 保证原子性：否则两个并发刷新请求可能同时读到同一个旧令牌，
// 导致轮换失效。
var consumeScript = `
local v = redis.call('GET', KEYS[1])
if v then
  redis.call('DEL', KEYS[1])
end
return v
`

// Consume 消费（取出并删除）一个刷新令牌，返回其中的用户 ID 与令牌版本。
// 令牌不存在、已过期或已被消费过时返回 ErrRefreshTokenInvalid。
func (s *RefreshStore) Consume(ctx context.Context, raw string) (int64, int64, error) {
	if raw == "" {
		return 0, 0, ErrRefreshTokenInvalid
	}

	val, err := s.rdb.Eval(ctx, consumeScript, []string{refreshKeyPrefix + redis.HashKey(raw)}).Result()
	if err != nil {
		if redis.IsNil(err) {
			return 0, 0, ErrRefreshTokenInvalid
		}
		return 0, 0, fmt.Errorf("authn: consume refresh token: %w", err)
	}
	if val == nil {
		return 0, 0, ErrRefreshTokenInvalid
	}

	str, ok := val.(string)
	if !ok {
		return 0, 0, ErrRefreshTokenInvalid
	}

	var p refreshPayload
	if err := json.Unmarshal([]byte(str), &p); err != nil {
		return 0, 0, ErrRefreshTokenInvalid
	}
	return p.UserID, p.TokenVersion, nil
}
