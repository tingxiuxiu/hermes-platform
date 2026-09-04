// Command scheduler 是 asynq 定时投递端（替代 celery-beat）。
//
// 职责：按 AUTOMATION_DASHBOARD_REFRESH_SECONDS 周期投递 dashboard:refresh:all，
// 作为事件驱动的兜底——即使上报链路的事件投递失败，读模型也能定期自愈。
//
// ⚠️ **单实例约束**：scheduler 必须只部署一个副本。多实例会重复投递
// refresh:all 任务（asynq 的 Unique(5m) 只能部分缓解，无法完全消除重复）。
// 见 docs/01-architecture.md §4。
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/logging"
	"github.com/hermes-platform/go-service/internal/platform/queue"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("scheduler: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger := logging.New(cfg.Log)

	refreshSec := cfg.Dashboard.RefreshSeconds
	if refreshSec <= 0 {
		refreshSec = 60
	}
	trendDays := cfg.Dashboard.TrendDays
	if trendDays <= 0 {
		trendDays = 7
	}

	redisAddr := net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port))
	scheduler := queue.NewScheduler(redisAddr, cfg.Redis.DB)

	// 注册周期任务：每隔 refreshSec 秒投递 refresh:all。
	cronSpec := fmt.Sprintf("@every %ds", refreshSec)
	task, err := queue.NewDashboardRefreshAllTask(trendDays)
	if err != nil {
		return fmt.Errorf("build refresh-all task: %w", err)
	}

	entryID, err := scheduler.Register(cronSpec, task, queue.CommonOptions()...)
	if err != nil {
		return fmt.Errorf("register scheduler: %w", err)
	}
	logger.Info("scheduler: registered refresh-all",
		"entry_id", entryID, "cron", cronSpec, "trend_days", trendDays)

	abortTask, err := queue.NewAbortStaleTask()
	if err != nil {
		return fmt.Errorf("build abort-stale task: %w", err)
	}
	abortID, err := scheduler.Register("@every 15s", abortTask, queue.AbortStaleOptions()...)
	if err != nil {
		return fmt.Errorf("register abort-stale scheduler: %w", err)
	}
	logger.Info("scheduler: registered abort-stale", "entry_id", abortID)

	// scheduler 需要后台运行以便接收信号；用 goroutine 执行并等待。
	errCh := make(chan error, 1)
	go func() {
		if err := scheduler.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("scheduler: shutdown signal received")
		return nil
	case err := <-errCh:
		return fmt.Errorf("scheduler run: %w", err)
	}
}
