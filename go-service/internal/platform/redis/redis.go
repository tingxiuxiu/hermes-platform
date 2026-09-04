// Package redis 封装 go-redis 客户端。
package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/hermes-platform/go-service/internal/platform/config"
)

// Client 是 Redis 客户端包装。
type Client struct {
	*goredis.Client
}

// New 建立客户端并验证连通性。
func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr(),
		Username:     "",
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	return &Client{Client: client}, nil
}

// NewForDB 使用相同连接参数但不同 DB index 建立客户端。
// 测试通过它把数据隔离到独立 DB（ADR-0007 要求测试用 db15）。
func NewForDB(ctx context.Context, cfg config.RedisConfig, db int) (*Client, error) {
	cfg.DB = db
	return New(ctx, cfg)
}

// HashKey 对可变长输入做定长散列，用于构造稳定的 Redis key。
// 例如把用户名 + IP 作为登录失败计数的 key，避免特殊字符破坏 key 结构。
func HashKey(parts ...string) string {
	h := sha256.New()
	for i, p := range parts {
		if i > 0 {
			h.Write([]byte{0})
		}
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// IsNil 判断是否为 Redis 的 key 不存在错误。
func IsNil(err error) bool {
	return err == goredis.Nil
}
