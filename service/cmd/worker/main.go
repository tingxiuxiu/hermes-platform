// Command worker 是 asynq 消费端（替代 celery-worker）。
//
// 职责：启动 asynq Server，消费 dashboard 快照投影任务。
// 任务来源：
//   - 上报链路的事件（经 bootstrap.EventPublisher 投递）；
//   - scheduler 定时投递的 refresh:all（兜底）。
//
// 优雅停机：asynq Config.ShutdownTimeout=30s，等待在途任务完成。
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
	"time"

	"github.com/hermes-platform/go-service/internal/adapter/persistence/automation"
	"github.com/hermes-platform/go-service/internal/adapter/persistence/dashboard"
	"github.com/hermes-platform/go-service/internal/adapter/worker"
	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/database"
	"github.com/hermes-platform/go-service/internal/platform/logging"
	"github.com/hermes-platform/go-service/internal/platform/queue"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("worker: %v", err)
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

	// worker 只读 DB（投影源表）与 Redis（asynq broker）。
	db, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer db.Close()

	// 装配 dashboard 投影器
	snapshots := dashboard.NewSnapshotRepo(db)
	source := dashboard.NewExecutionSourceReader(db)
	projector := appdashboard.NewProjectorUseCase(snapshots, source, systemClock{})

	ingest := appautomation.NewIngestUseCase(
		automation.NewExecutionRepo(db),
		automation.NewItemRepo(db),
		automation.NewStepRepo(db),
		automation.NewAttachmentRepo(db),
		automation.NewPipelineRepo(db),
		systemClock{},
		nil,
	)

	// 装配 asynq Server
	redisAddr := net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port))
	server := queue.NewServer(redisAddr, cfg.Redis.DB,
		map[string]int{queue.QueueDashboard: 5, queue.QueueDefault: 5})

	// 注册任务处理器
	handler := worker.NewDashboardHandler(projector, logger)
	abortTimeout := time.Duration(cfg.Automation.HeartbeatTimeoutSeconds) * time.Second
	abortHandler := worker.NewAbortHandler(ingest, abortTimeout, logger)
	mux := queue.ServeMux()
	handler.Register(mux)
	abortHandler.Register(mux)

	// 启动消费（在 goroutine 中，等待退出信号）
	errCh := make(chan error, 1)
	go func() {
		logger.Info("worker: dashboard consumer started")
		errCh <- server.Start(mux)
	}()

	select {
	case <-ctx.Done():
		logger.Info("worker: shutdown signal received, waiting for in-flight tasks")
		return nil
	case err := <-errCh:
		return fmt.Errorf("asynq server: %w", err)
	}
}

// systemClock 适配 dashboard 应用层的 Clock 端口。
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
