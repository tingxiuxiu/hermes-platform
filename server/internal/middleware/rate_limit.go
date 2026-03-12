package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"com.hermes.platform/internal/utils"
)

// RateLimit 限流中间件
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP 作为限流键
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// 获取 Redis 客户端
		client := utils.GetRedisClient()
		if client == nil {
			// Redis 未初始化，跳过限流
			c.Next()
			return
		}

		ctx := context.Background()
		now := time.Now().Unix()
		windowStart := now - int64(window.Seconds())

		// 移除过期的请求记录
		client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

		// 获取当前请求数
		count, err := client.ZCard(ctx, key).Result()
		if err != nil {
			// Redis 错误，跳过限流
			c.Next()
			return
		}

		// 检查是否超过限制
		if count >= int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// 添加当前请求记录
		client.ZAdd(ctx, key, 
			redis.Z{
				Score:  float64(now),
				Member: fmt.Sprintf("%d", now),
			},
		)

		// 设置键过期时间
		client.Expire(ctx, key, window)

		c.Next()
	}
}
