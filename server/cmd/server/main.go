package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"com.hermes.platform/internal/api"
	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/database"
	"com.hermes.platform/internal/middleware"
	"com.hermes.platform/internal/utils"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 JWT 配置
	auth.InitJWT(cfg)

	// 初始化数据库
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	database.AutoMigrate(db)

	// 初始化默认数据
	database.SeedData(db)

	// 初始化 Redis
	if err := utils.InitRedis(cfg); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		// Redis 连接失败不阻止服务器启动
	}

	// 初始化 Gin 路由
	r := gin.Default()

	// 注册中间件
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	// 应用限流中间件：每分钟最多 60 个请求
	r.Use(middleware.RateLimit(60, time.Minute))

	// 注册 API 路由
	api.RegisterRoutes(r, db)

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server running on %s in %s mode\n", serverAddr, cfg.Env)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
