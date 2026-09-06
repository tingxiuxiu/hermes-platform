// Command api 是 HTTP API 主进程。
//
// 职责：
//   - 加载配置、装配应用（bootstrap.New）
//   - 启动前应用 golang-migrate schema（空库建表；已是最新则跳过）
//   - 首启时幂等创建超级管理员（FIRST_SUPERUSER / FIRST_SUPERUSER_PASSWORD）
//   - 监听 SERVICE_PORT（默认 80），优雅停机
//
// 优雅停机流程：收到 SIGINT/SIGTERM → 停止接收新请求（http.Server.Shutdown）→
// 等待在途请求完成（超时上限）→ 关闭数据库/Redis 连接池。
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hermes-platform/go-service/internal/bootstrap"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/migrations"
	migsql "github.com/hermes-platform/go-service/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 装配应用（含 DB / Redis 连接、全部用例与路由）
	app, err := bootstrap.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("bootstrap app: %w", err)
	}
	defer app.Close()

	if err := applySchema(cfg.Postgres.DSN()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// 首启种子：角色/权限/超级管理员（幂等，可安全重复执行）
	if err := bootstrap.SeedAll(ctx, app, cfg); err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           app.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		app.Log.Info("api: server listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		// 优雅停机：停止接收新请求，等待在途完成
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		app.Log.Info("api: shutdown signal received, draining in-flight requests")
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		app.Log.Info("api: shutdown complete")
		return nil

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}

func applySchema(dsn string) error {
	if err := migrations.ResetBookkeepingIfEmpty(dsn); err != nil {
		return err
	}
	m, err := migrations.NewFS(migsql.FS, dsn)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	return migrations.Up(m)
}
