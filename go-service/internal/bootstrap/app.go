// Package bootstrap 是组合根：唯一允许 import 所有层的地方（ADR-0004）。
//
// 这里只做三件事：构造基础设施 → 装配用例与适配器 → 返回可运行的 HTTP 引擎。
// 任何业务规则都不得出现在本包。
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hermes-platform/go-service/internal/adapter/cache/identity"
	"github.com/hermes-platform/go-service/internal/adapter/http"
	automationhttp "github.com/hermes-platform/go-service/internal/adapter/http/automation"
	dashboardhttp "github.com/hermes-platform/go-service/internal/adapter/http/dashboard"
	identityhttp "github.com/hermes-platform/go-service/internal/adapter/http/identity"
	"github.com/hermes-platform/go-service/internal/adapter/http/middleware"
	systemhttp "github.com/hermes-platform/go-service/internal/adapter/http/system"
	liveadapter "github.com/hermes-platform/go-service/internal/adapter/live"
	automationpersist "github.com/hermes-platform/go-service/internal/adapter/persistence/automation"
	dashboardpersist "github.com/hermes-platform/go-service/internal/adapter/persistence/dashboard"
	persistence "github.com/hermes-platform/go-service/internal/adapter/persistence/identity"
	systempersist "github.com/hermes-platform/go-service/internal/adapter/persistence/system"
	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	appdashboard "github.com/hermes-platform/go-service/internal/application/dashboard"
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	appsystem "github.com/hermes-platform/go-service/internal/application/system"
	"github.com/hermes-platform/go-service/internal/platform/authn"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/database"
	"github.com/hermes-platform/go-service/internal/platform/logging"
	"github.com/hermes-platform/go-service/internal/platform/password"
	"github.com/hermes-platform/go-service/internal/platform/queue"
	appredis "github.com/hermes-platform/go-service/internal/platform/redis"
	"github.com/hermes-platform/go-service/internal/platform/telemetry"
)

// App 是一个已装配完毕的应用实例。
type App struct {
	Cfg    config.Config
	Log    *slog.Logger
	DB     *database.Pool
	Redis  *appredis.Client
	Router *gin.Engine

	// Services 暴露已装配的用例，便于其他进程（worker）复用与测试断言。
	Services Services
}

// Services 汇总各上下文的用例。
type Services struct {
	Auth        *appidentity.AuthUseCase
	Users       *appidentity.UserUseCase
	Roles       *appidentity.RoleUseCase
	Permissions *appidentity.PermissionUseCase

	// Automation 上报与查询用例（Phase 6 前事件走 noop 发布器）。
	AutomationIngest *appautomation.IngestUseCase
	AutomationQuery  *appautomation.QueryUseCase
}

// New 装配整个应用并自行建立数据库与 Redis 连接。
// ctx 只用于启动期的连通性校验，不用于控制生命周期。
func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	rdb, err := appredis.New(ctx, cfg.Redis)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return NewWith(ctx, cfg, db, rdb, newProductionPublisher(cfg, rdb))
}

// newProductionPublisher 构造生产环境的事件发布器（asynq 桥接）。
// 使用与业务 Redis 相同的配置。
func newProductionPublisher(cfg config.Config, rdb *appredis.Client) appautomation.EventPublisher {
	client := queue.NewClient(net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port)), cfg.Redis.DB)
	return NewEventPublisher(client)
}

// NewWith 使用**已存在的**连接装配应用，不负责关闭它们。
//
// 供集成测试使用：测试需要把应用指向 hermes_test 库与独立的 Redis DB，
// 而不是配置里的业务库。这样测试与生产走完全相同的装配路径，
// 不会出现"测试测的是另一套 wiring"的情况。
//
// publisher 是可选的事件发布器：测试不传时用 noop（事件不投递 asynq），
// 生产由 New 传入 asynq 桥接器。
func NewWith(ctx context.Context, cfg config.Config, db *database.Pool, rdb *appredis.Client, publishers ...appautomation.EventPublisher) (*App, error) {
	logger := logging.New(cfg.Log)

	setGinMode(cfg)

	services, err := newIdentityServices(cfg, db, rdb)
	if err != nil {
		db.Close()
		_ = rdb.Close()
		return nil, err
	}

	// automation 的仓储与用例。事件发布器：默认 noop（测试），生产传 asynq 桥接。
	var publisher appautomation.EventPublisher = noopEventPublisher{}
	if len(publishers) > 0 && publishers[0] != nil {
		publisher = publishers[0]
	}
	liveHub := liveadapter.NewHub(rdb)
	automationServices := newAutomationServices(db, publisher, liveHub)
	// dashboard 的读侧：源数据读取 + 快照投影。
	dashboardServices := newDashboardServices(db)
	// system 字典
	systemServices := newSystemServices(db)

	// OpenTelemetry：配置了 OTLP 端点才真正导出，否则 no-op（不阻塞启动）
	const serviceName = "hermes-api"
	telemetry.Init(cfg.OTel.Endpoint, serviceName, logger)

	// 把 automation 用例暴露到 App.Services，供 worker/其他进程复用。
	services.AutomationIngest = automationServices.Ingest
	services.AutomationQuery = automationServices.Query

	router := http.NewRouter(http.RouterDeps{
		Cfg:          cfg,
		Log:          logger,
		Authenticate: buildAuthN(cfg, db, rdb),
		ServiceToken: middleware.ServiceToken(cfg.Automation.ServiceToken),
		Telemetry:    telemetry.InstrumentHTTP(cfg.OTel.Endpoint, serviceName, logger),
		Auth:         identityhttp.NewAuthHandlers(services.Auth),
		Identity:     identityhttp.NewHandlers(services.Users, services.Roles, services.Permissions),
		Automation: automationhttp.NewHandlers(
			automationServices.Ingest,
			automationServices.Query,
			automationServices.Publisher,
			automationServices.Items,
			liveHub,
		),
		Dashboard: dashboardServices.HTTPHandlers,
		System:    systemServices.HTTPHandlers,
		Readiness: func() error {
			if err := db.Ping(ctx); err != nil {
				return err
			}
			return rdb.Ping(ctx).Err()
		},
	})

	return &App{
		Cfg:      cfg,
		Log:      logger,
		DB:       db,
		Redis:    rdb,
		Router:   router,
		Services: services,
	}, nil
}

// newIdentityServices 装配 identity 上下文的全部用例。
func newIdentityServices(cfg config.Config, db *database.Pool, rdb *appredis.Client) (Services, error) {
	users := persistence.NewUserRepo(db)
	roles := persistence.NewRoleRepo(db)
	permissions := persistence.NewPermissionRepo(db)
	cache := identity.NewUserCache(rdb, cfg.Auth.UserCacheTTL)

	hasher := password.New()
	issuer := authn.NewIssuer(cfg.Auth)
	refreshStore := authn.NewRefreshStore(rdb, cfg.Auth.RefreshTokenTTL)
	revoker := authn.NewRevoker(rdb)
	guard := authn.NewLoginGuard(rdb, cfg.Auth.MaxFailedLogins, cfg.Auth.LoginLockDuration)

	services := Services{
		Auth: appidentity.NewAuthUseCase(
			users, roles, hasher, issuer, refreshStore, revoker, guard, clockAdapter{}),
		Users:       appidentity.NewUserUseCase(users, roles, hasher, revoker, cache),
		Roles:       appidentity.NewRoleUseCase(roles, permissions),
		Permissions: appidentity.NewPermissionUseCase(permissions),
	}

	return services, nil
}

// automationServices 汇总 automation 上下文的用例与基础设施引用。
type automationServices struct {
	Ingest    *appautomation.IngestUseCase
	Query     *appautomation.QueryUseCase
	Publisher appautomation.EventPublisher
	Items     appautomation.ItemRepository
}

// newAutomationServices 装配 automation 上下文。
// 事件发布由调用方注入（生产 asynq 桥接 / 测试 noop）。
func newAutomationServices(db *database.Pool, publisher appautomation.EventPublisher, live appautomation.LivePublisher) automationServices {
	executions := automationpersist.NewExecutionRepo(db)
	items := automationpersist.NewItemRepo(db)
	steps := automationpersist.NewStepRepo(db)
	attachments := automationpersist.NewAttachmentRepo(db)
	pipelines := automationpersist.NewPipelineRepo(db)

	ingest := appautomation.NewIngestUseCase(
		executions, items, steps, attachments, pipelines, clockAdapter{}, live)
	query := appautomation.NewQueryUseCase(
		executions, items, steps, attachments, pipelines)

	return automationServices{
		Ingest:    ingest,
		Query:     query,
		Publisher: publisher,
		Items:     items,
	}
}

// dashboardServices 汇总 dashboard 上下文的用例与 HTTP 处理器。
type dashboardServices struct {
	HTTPHandlers *dashboardhttp.Handlers
}

// newDashboardServices 装配 dashboard 上下文。
func newDashboardServices(db *database.Pool) dashboardServices {
	snapshots := dashboardpersist.NewSnapshotRepo(db)
	source := dashboardpersist.NewExecutionSourceReader(db)

	projector := appdashboard.NewProjectorUseCase(snapshots, source, clockAdapter{})
	overview := appdashboard.NewOverviewUseCase(
		snapshots, source, projector, clockAdapter{}, 30*time.Second)
	runningCases := appdashboard.NewRunningCasesUseCase(
		snapshots, source, projector, clockAdapter{}, 30*time.Second)

	return dashboardServices{
		HTTPHandlers: dashboardhttp.NewHandlers(overview, runningCases),
	}
}

// systemServices 汇总 system 上下文的用例与 HTTP 处理器。
type systemServices struct {
	HTTPHandlers *systemhttp.Handlers
}

// newSystemServices 装配 system 上下文（系统字典）。
func newSystemServices(db *database.Pool) systemServices {
	dicts := appsystem.NewDictUseCase(systempersist.NewDictRepo(db))
	return systemServices{
		HTTPHandlers: systemhttp.NewHandlers(dicts),
	}
}

// Close 释放全部资源。
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}

func buildAuthN(cfg config.Config, db *database.Pool, rdb *appredis.Client) gin.HandlerFunc {
	return middleware.AuthN(middleware.AuthNDeps{
		Issuer:  authn.NewIssuer(cfg.Auth),
		Revoker: authn.NewRevoker(rdb),
		Users:   persistence.NewUserRepo(db),
		Cache:   identity.NewUserCache(rdb, cfg.Auth.UserCacheTTL),
	})
}

func setGinMode(cfg config.Config) {
	if cfg.IsLocal() {
		gin.SetMode(gin.DebugMode)
		return
	}
	gin.SetMode(gin.ReleaseMode)
}

// clockAdapter 把 platform/clock 的 System 适配成 application 层的 Clock 端口。
// 单独包一层而不是直接传 clock.System{}，是为了让测试可以替换实现。
type clockAdapter struct{}

func (clockAdapter) Now() time.Time { return time.Now().UTC() }
