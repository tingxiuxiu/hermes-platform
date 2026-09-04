// Package identity 实现 identity 上下文的 Redis 缓存。
package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/redis"
)

const userCacheKeyPrefix = "user:cache:%d"

// UserCache 缓存「当前登录用户」的查询结果，TTL 与 Python 版一致（300 秒）。
//
// 安全约束：缓存中**绝不包含 password_hash**。
// 该缓存只服务于请求鉴权（get_current_user 等价物），不需要密码哈希；
// 一旦把哈希写进 Redis，就等于把凭证多复制了一份到另一个可被 dump 的存储里。
type UserCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewUserCache 构造用户缓存。
func NewUserCache(rdb *redis.Client, ttl time.Duration) *UserCache {
	if ttl <= 0 {
		ttl = 300 * time.Second
	}
	return &UserCache{rdb: rdb, ttl: ttl}
}

var _ appidentity.UserCache = (*UserCache)(nil)

// cachedUser 是缓存值的结构。刻意不包含密码哈希。
type cachedUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int16  `json:"status"`
	Roles    []struct {
		ID          int64   `json:"id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	} `json:"roles"`
}

// Get 读取缓存。未命中或数据损坏时返回 nil, nil——
// 缓存不是数据源，任何异常都应降级为「未命中」而不是让请求失败。
func (c *UserCache) Get(ctx context.Context, userID int64) (*identity.User, error) {
	key := fmt.Sprintf(userCacheKeyPrefix, userID)

	raw, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if redis.IsNil(err) {
			return nil, nil
		}
		return nil, nil
	}
	if raw == "" {
		return nil, nil
	}

	var cached cachedUser
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		return nil, nil
	}

	roleSnapshots := make([]identity.RoleSnapshot, 0, len(cached.Roles))
	for _, r := range cached.Roles {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		roleSnapshots = append(roleSnapshots, identity.RoleSnapshot{
			ID:          r.ID,
			Code:        r.Code,
			Name:        r.Name,
			Description: desc,
		})
	}

	roles := make([]*identity.Role, 0, len(roleSnapshots))
	for i := range roleSnapshots {
		role := identity.RestoreRole(roleSnapshots[i])
		roles = append(roles, role)
	}

	return identity.RestoreUser(identity.UserSnapshot{
		ID:       cached.ID,
		Username: cached.Username,
		Email:    cached.Email,
		// PasswordHash 故意留空：缓存不承载凭证
		Status: cached.Status,
		Roles:  roles,
	}), nil
}

// Set 写入缓存。写失败只记录不影响主流程（与 Python 的行为一致）。
func (c *UserCache) Set(ctx context.Context, u *identity.User) error {
	if u == nil {
		return nil
	}

	roles := make([]struct {
		ID          int64   `json:"id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}, 0, len(u.Roles()))

	for _, r := range u.Roles() {
		var desc *string
		if r.Description() != "" {
			v := r.Description()
			desc = &v
		}
		roles = append(roles, struct {
			ID          int64   `json:"id"`
			Code        string  `json:"code"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
		}{ID: r.ID(), Code: r.Code(), Name: r.Name(), Description: desc})
	}

	payload, err := json.Marshal(cachedUser{
		ID:       u.ID(),
		Username: u.Username(),
		Email:    u.Email(),
		Status:   int16(u.Status()),
		Roles:    roles,
	})
	if err != nil {
		return fmt.Errorf("cache: marshal user: %w", err)
	}

	key := fmt.Sprintf(userCacheKeyPrefix, u.ID())
	if err := c.rdb.Set(ctx, key, payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache: set user: %w", err)
	}
	return nil
}

// Delete 清除指定用户的缓存。用户信息变更后调用。
func (c *UserCache) Delete(ctx context.Context, userID int64) error {
	key := fmt.Sprintf(userCacheKeyPrefix, userID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache: delete user: %w", err)
	}
	return nil
}
