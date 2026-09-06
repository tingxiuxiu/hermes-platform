// Package http 是 HTTP 适配层入口。
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerHealthRoutes 注册健康检查路由（**不加 API 前缀**，供探针直接访问）。
//
// - /healthz：存活探针，无任何依赖，进程活着即返回 200。
// - /readyz：就绪探针，调用 Readiness 检查 PG + Redis，未就绪返回 503。
func registerHealthRoutes(engine *gin.Engine, deps RouterDeps) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/readyz", func(c *gin.Context) {
		if deps.Readiness != nil {
			if err := deps.Readiness(); err != nil {
				deps.Log.Error("readiness check failed", "error", err)
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
