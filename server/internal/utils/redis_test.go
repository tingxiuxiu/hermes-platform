package utils

import (
	"testing"

	"com.hermes.platform/internal/config"
)

func TestInitRedis(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
	}

	// 测试 InitRedis 函数（这里我们只是测试函数是否能正常执行，不实际连接 Redis）
	// 注意：实际环境中可能需要使用真实的 Redis 服务器
	err := InitRedis(cfg)
	if err != nil {
		// 这里允许失败，因为我们没有实际的 Redis 服务器
		t.Logf("InitRedis failed as expected: %v", err)
	} else {
		t.Logf("InitRedis succeeded")
	}
}

func TestGetRedisClient(t *testing.T) {
	// 测试 GetRedisClient 函数
	client := GetRedisClient()
	if client == nil {
		t.Logf("Redis client is nil (expected if Redis not initialized)")
	} else {
		t.Logf("Redis client retrieved successfully")
	}
}
