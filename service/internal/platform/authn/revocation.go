package authn

import (
	"context"
	"fmt"

	"github.com/hermes-platform/go-service/internal/platform/redis"
)

const tokenVersionKeyPrefix = "user:%d:token_version"

// Revoker 通过 Redis 中的令牌版本号实现「一次操作吊销该用户全部令牌」（ADR-0005）。
//
// 为什么用版本号而不是黑名单：
//   - 黑名单需要为每个已签发令牌存一条记录，内存随签发量线性增长；
//   - 版本号每个用户只有一个 key，登出/改密/禁用只需一次 INCR。
//
// 版本号 key **不设 TTL**：
// 若 TTL 到期后版本号回落到 0，此前被吊销的旧令牌（版本 0）会重新变成有效，
// 这是不可接受的安全回退。代价是每个用户常驻一个极小的 key。
type Revoker struct {
	rdb *redis.Client
}

// NewRevoker 构造吊销器。
func NewRevoker(rdb *redis.Client) *Revoker {
	return &Revoker{rdb: rdb}
}

// Version 返回用户当前的令牌版本号。key 不存在时返回 0（初始版本）。
// Redis 不可用时返回 0 而不报错：令牌校验降级为「只校验签名与过期时间」，
// 与 Python 版「Redis 异常不影响主流程」的行为一致。
func (r *Revoker) Version(ctx context.Context, userID int64) (int64, error) {
	key := fmt.Sprintf(tokenVersionKeyPrefix, userID)

	v, err := r.rdb.Get(ctx, key).Int64()
	if err != nil {
		if redis.IsNil(err) {
			return 0, nil
		}
		return 0, nil
	}
	return v, nil
}

// Revoke 递增用户的令牌版本号，使其全部已签发令牌立即失效。
// 返回递增后的版本号。
func (r *Revoker) Revoke(ctx context.Context, userID int64) (int64, error) {
	key := fmt.Sprintf(tokenVersionKeyPrefix, userID)

	v, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("authn: revoke user %d: %w", userID, err)
	}
	return v, nil
}
