package utils

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"com.hermes.platform/internal/config"
)

// RedisClient Redis 客户端实例
var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	return nil
}

// GetRedisClient 获取 Redis 客户端实例
func GetRedisClient() *redis.Client {
	return RedisClient
}
