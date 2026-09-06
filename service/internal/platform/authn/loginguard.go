package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/hermes-platform/go-service/internal/platform/redis"
)

const loginFailKeyPrefix = "login:fail:"

// LoginGuard 防御暴力破解：按 (用户名, 来源 IP) 统计连续失败次数，
// 超过阈值后在一段时间内直接拒绝，不再校验密码。
//
// 之所以带上 IP：避免攻击者针对同一个已知用户名、从不同 IP 轮询时，
// 把正常用户「挤」进锁定状态（仅按用户名计数会造成拒绝服务）。
type LoginGuard struct {
	rdb         *redis.Client
	maxAttempts int
	lockFor     time.Duration
}

// NewLoginGuard 构造登录守卫。
func NewLoginGuard(rdb *redis.Client, maxAttempts int, lockFor time.Duration) *LoginGuard {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if lockFor <= 0 {
		lockFor = 15 * time.Minute
	}
	return &LoginGuard{rdb: rdb, maxAttempts: maxAttempts, lockFor: lockFor}
}

func (g *LoginGuard) key(username, clientIP string) string {
	return loginFailKeyPrefix + redis.HashKey(username, clientIP)
}

// Check 判断是否处于锁定状态。锁定时返回 ErrLoginLocked。
// Redis 不可用时放行，避免基础设施故障导致全体用户无法登录。
func (g *LoginGuard) Check(ctx context.Context, username, clientIP string) error {
	count, err := g.rdb.Get(ctx, g.key(username, clientIP)).Int()
	if err != nil {
		if redis.IsNil(err) {
			return nil
		}
		return nil // 降级放行
	}
	if count >= g.maxAttempts {
		return ErrLoginLocked
	}
	return nil
}

// RecordFailure 记录一次登录失败并刷新锁定时长。
func (g *LoginGuard) RecordFailure(ctx context.Context, username, clientIP string) error {
	key := g.key(username, clientIP)

	pipe := g.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, g.lockFor)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("authn: record login failure: %w", err)
	}
	_ = incr
	return nil
}

// Clear 清除失败计数，登录成功后调用。
func (g *LoginGuard) Clear(ctx context.Context, username, clientIP string) {
	_ = g.rdb.Del(ctx, g.key(username, clientIP)).Err()
}
