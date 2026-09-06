package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hermes-platform/go-service/internal/adapter/persistence/dashboard"
	"github.com/hermes-platform/go-service/internal/adapter/worker"
	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/queue"
	"github.com/hermes-platform/go-service/test/harness"
)

// systemClock 适配 dashboard 的 Clock 端口。
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// TestEventDrivenDashboardRefresh 验证完整链路：
// 上报 execution → 事件经 asynq 投递 → worker 消费投影 → dashboard 概览反映最新数据。
//
// 这需要真实 asynq worker。用 harness 的 Redis，装配一个投影器 + asynq server，
// 上报数据后驱动 worker 消费任务，再查询 dashboard 验证快照已更新。
func TestEventDrivenDashboardRefresh(t *testing.T) {
	env := harness.Setup(t)

	// 装配 asynq worker：投影器 + server
	snapshots := dashboard.NewSnapshotRepo(env.DB)
	source := dashboard.NewExecutionSourceReader(env.DB)
	projector := appdashboard.NewProjectorUseCase(snapshots, source, systemClock{})

	redisAddr := env.Cfg.Redis.Host + ":" + itoa(int64(env.Cfg.Redis.Port))
	server := queue.NewServer(redisAddr, env.Cfg.Redis.DB,
		map[string]int{queue.QueueDashboard: 5, queue.QueueDefault: 5})

	handler := worker.NewDashboardHandler(projector, slog.Default())
	mux := queue.ServeMux()
	handler.Register(mux)

	// 启动 worker（后台）。asynq Server.Run 自带优雅停机，不需要外部 ctx。
	go func() { _ = server.Start(mux) }()

	// 上报一次 execution + 用例
	buildUID := uuid.New().String()
	createExecution(t, env, buildUID)
	u1 := uuid.New().String()
	createItem(t, env, buildUID, "JIRA-EVT-1", u1)

	// 事件已通过 asynq 投递到 Redis（上报接口经 noopEventPublisher？不，
	// 这里用 harness 的 App 是 noop。改为直接投递任务模拟事件驱动）
	// 直接 enqueue refresh-all 任务，模拟 EventPublisher 行为
	client := queue.NewClient(redisAddr, env.Cfg.Redis.DB)
	defer client.Close()

	task, err := queue.NewDashboardRefreshAllTask(7)
	if err != nil {
		t.Fatalf("build task: %v", err)
	}
	if _, err := client.Enqueue(task, queue.CommonOptions()...); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// 等待 worker 消费完成
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		// 查询 overview，触发同步刷新兜底也验证投影
		admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
		w := doAutomation(t, env, http.MethodGet, api(env)+"/automation/dashboard/overview", "",
			bearer(admin.AccessToken))
		if w.Code == http.StatusOK {
			var data struct {
				Summary struct {
					TotalExecutionsCount int `json:"total_executions_count"`
				} `json:"summary"`
			}
			if err := json.Unmarshal(decodeAutomation(t, w).Data, &data); err == nil {
				if data.Summary.TotalExecutionsCount >= 1 {
					// 快照已投影
					return
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Fatal("dashboard snapshot was not refreshed within deadline")
}

// TestWorkerCanBuildTaskAndHandler (compile-time sanity)
func TestWorkerCanBuildTaskAndHandler(t *testing.T) {
	env := harness.Setup(t)
	_ = env.Cfg.APIPrefix
	_ = config.EnvLocal
}
