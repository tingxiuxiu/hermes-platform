package services

import (
	"context"
	"encoding/json"
	"time"

	"com.hermes.platform/internal/utils"
)

// CacheService 缓存服务接口
type CacheService interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Delete(key string) error
	Exists(key string) (bool, error)
}

// cacheService 缓存服务实现
type cacheService struct {}

// NewCacheService 创建缓存服务实例
func NewCacheService() CacheService {
	return &cacheService{}
}

// Set 设置缓存
func (s *cacheService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	client := utils.GetRedisClient()
	if client == nil {
		return nil // Redis 未初始化，跳过缓存
	}

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (s *cacheService) Get(key string, dest interface{}) error {
	ctx := context.Background()
	client := utils.GetRedisClient()
	if client == nil {
		return nil // Redis 未初始化，跳过缓存
	}

	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func (s *cacheService) Delete(key string) error {
	ctx := context.Background()
	client := utils.GetRedisClient()
	if client == nil {
		return nil // Redis 未初始化，跳过缓存
	}

	return client.Del(ctx, key).Err()
}

// Exists 检查缓存是否存在
func (s *cacheService) Exists(key string) (bool, error) {
	ctx := context.Background()
	client := utils.GetRedisClient()
	if client == nil {
		return false, nil // Redis 未初始化，返回不存在
	}

	result, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
