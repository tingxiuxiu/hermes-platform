// Package harness 提供集成测试的基础设施。
//
// 设计原则（ADR-0007）：
//   - 所有集成测试连接**真实的** PostgreSQL（hermes_test 库）与真实的 Redis（db15），
//     不使用任何 mock 替身；
//   - 每个测试函数结束时 TRUNCATE 全部表，保证测试之间互不污染；
//   - Setup 强制注册 cleanup，不允许绕过 harness 直接拿连接。
package harness

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	identityrepo "github.com/hermes-platform/go-service/internal/adapter/persistence/identity"
	"github.com/hermes-platform/go-service/internal/bootstrap"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/database"
	"github.com/hermes-platform/go-service/internal/platform/logging"
	"github.com/hermes-platform/go-service/internal/platform/migrations"
	appredis "github.com/hermes-platform/go-service/internal/platform/redis"
)

// DefaultTestDatabaseName 是默认的测试库名。
//
// 每个测试包应通过 TEST_DATABASE_NAME 环境变量使用**自己的**库：
// `go test ./...` 会并行运行多个包，共用同一个库会导致
// 一个包 TRUNCATE 掉另一个包正在使用的数据。
const DefaultTestDatabaseName = "hermes_test"

// 全部业务表，TRUNCATE 时按此列表清空。
// RESTART IDENTITY 让自增主键归零，测试断言可以依赖确定的 id 值。
var allTables = []string{
	"automation_dashboard_execution_item_snapshots",
	"automation_dashboard_execution_snapshots",
	"automation_dashboard_daily_trends",
	"automation_dashboard_summary_snapshots",
	"execution_case_step_attachments",
	"execution_item_steps",
	"execution_items",
	"test_executions",
	"jenkins_pipelines",
	"sys_dict",
	"role_permissions",
	"user_roles",
	"permissions",
	"roles",
	"users",
}

var (
	prepareOnce sync.Once
	prepareErr  error
	sharedCfg   config.Config
)

// Env 是一个测试的运行环境。
type Env struct {
	Cfg   config.Config
	Log   *slog.Logger
	DB    *database.Pool
	Redis *appredis.Client
}

// Setup 准备测试环境并保证清理。
//
// 首次调用会在 hermes_test 库上执行迁移；后续调用复用。
// 每个测试函数都会注册 TRUNCATE cleanup，因此测试之间数据互不可见。
func Setup(t *testing.T) *Env {
	t.Helper()

	prepareOnce.Do(func() {
		sharedCfg, prepareErr = prepareTestDatabase()
	})
	if prepareErr != nil {
		t.Fatalf("harness: prepare test database: %v", prepareErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.New(ctx, testPostgresConfig(sharedCfg))
	if err != nil {
		t.Fatalf("harness: connect test database: %v", err)
	}
	rdb, err := appredis.NewForDB(ctx, sharedCfg.Redis, sharedCfg.Test.RedisDB)
	if err != nil {
		pool.Close()
		t.Fatalf("harness: connect test redis: %v", err)
	}

	env := &Env{
		Cfg:   sharedCfg,
		Log:   logging.NewNop(),
		DB:    pool,
		Redis: rdb,
	}

	t.Cleanup(func() {
		env.Truncate(t)
		_ = rdb.FlushDB(context.Background()).Err()
		pool.Close()
		_ = rdb.Close()
	})

	// 建立基线。seed 幂等，因此在 Setup 与 Truncate 里都要调用：
	// 只放在 Truncate 里的话，「建库后的第一个用例」会拿不到 admin 角色。
	env.Seed(t)

	return env
}

// Truncate 清空全部业务表、重置自增序列，并重建基线种子数据。
//
// App 用测试库与测试 Redis 装配出一个完整应用（含路由表与全部中间件）。
//
// 走的是与生产完全相同的 bootstrap.NewWith 路径，
// 避免「测试用的是另一套 wiring」导致问题只在生产环境暴露。
func (e *Env) App(t *testing.T) *bootstrap.App {
	t.Helper()

	app, err := bootstrap.NewWith(context.Background(), e.Cfg, e.DB, e.Redis)
	if err != nil {
		t.Fatalf("harness: bootstrap app: %v", err)
	}
	return app
}

// Router 返回装配好的 gin 引擎，供 httptest 直接驱动。
func (e *Env) Router(t *testing.T) *gin.Engine {
	t.Helper()
	return e.App(t).Router
}

// 种子必须重建：TRUNCATE 会把 roles / permissions 一并清空，
// 若不补种，后续用例就拿不到 admin 角色。种子逻辑与生产启动共用
// bootstrap.SeedRolesAndPermissions，不会漂移。
func (e *Env) Truncate(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmt := "TRUNCATE TABLE " + joinTables(allTables) + " RESTART IDENTITY CASCADE"
	if _, err := e.DB.Exec(ctx, stmt); err != nil {
		t.Fatalf("harness: truncate: %v", err)
	}

	seeder := identityrepo.NewSeedRepo(e.DB)
	if err := bootstrap.SeedRolesAndPermissions(ctx, seeder); err != nil {
		t.Fatalf("harness: reseed baseline: %v", err)
	}
}

// Seed 对外暴露手动补种入口，供需要自定义基线数据的用例使用。
func (e *Env) Seed(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if err := bootstrap.SeedRolesAndPermissions(ctx, identityrepo.NewSeedRepo(e.DB)); err != nil {
		t.Fatalf("harness: seed: %v", err)
	}
}

// ServiceHeaders 返回自动化上报接口使用的服务令牌请求头。
func (e *Env) ServiceHeaders() map[string]string {
	return map[string]string{"X-Service-Token": e.Cfg.Automation.ServiceToken}
}

// WrongServiceHeaders 返回一个必然鉴权失败的服务令牌请求头。
func (e *Env) WrongServiceHeaders() map[string]string {
	return map[string]string{"X-Service-Token": "definitely-not-the-token"}
}

// ---------------------------------------------------------------------------
// 内部实现
// ---------------------------------------------------------------------------

// prepareTestDatabase 确保 hermes_test 库存在且迁移到最新版本。
func prepareTestDatabase() (config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	adminDSN, err := withDatabaseName(cfg.Test.DatabaseURL, "postgres")
	if err != nil {
		return config.Config{}, err
	}

	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return config.Config{}, fmt.Errorf("connect to maintenance db: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	dbName := cfg.Test.DatabaseName

	var exists bool
	err = conn.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName).Scan(&exists)
	if err != nil {
		return config.Config{}, fmt.Errorf("check test database: %w", err)
	}
	if !exists {
		if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
			return config.Config{}, fmt.Errorf("create test database: %w", err)
		}
	}

	if err := runMigrations(cfg.Test.DatabaseURL); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

// ResetTestDatabase 强制重建测试库。开发期 schema 变更后用它刷新。
func ResetTestDatabase() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	adminDSN, err := withDatabaseName(cfg.Test.DatabaseURL, "postgres")
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(ctx) }()

	dbName := cfg.Test.DatabaseName

	if err := terminateBackends(ctx, conn, dbName); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName); err != nil {
		return fmt.Errorf("drop test database: %w", err)
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		return fmt.Errorf("create test database: %w", err)
	}

	prepareOnce = sync.Once{}
	return runMigrations(cfg.Test.DatabaseURL)
}

func terminateBackends(ctx context.Context, conn *pgx.Conn, dbName string) error {
	_, err := conn.Exec(ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity "+
			"WHERE datname = $1 AND pid <> pg_backend_pid()",
		dbName)
	if err != nil {
		return fmt.Errorf("terminate backends: %w", err)
	}
	return nil
}

func runMigrations(dsn string) error {
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	m, err := migrations.New(migrationsDir, dsn)
	if err != nil {
		return err
	}
	if err := migrations.Up(m); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// findMigrationsDir 从当前目录向上查找 migrations 目录，
// 使测试在任意子包目录下运行都能定位到迁移文件。
func findMigrationsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("migrations directory not found (searched upward from %s)", cwd)
}

func testPostgresConfig(cfg config.Config) config.PostgresConfig {
	// 测试库连接参数沿用主库，仅替换连接串来源：
	// 解析 TEST_DATABASE_URL 得到 host/port/user/db，避免两处配置漂移。
	u, err := parseURL(cfg.Test.DatabaseURL)
	if err != nil {
		return cfg.Postgres
	}
	pc := cfg.Postgres
	pc.Host = u.Hostname()
	if u.Port() != "" {
		fmt.Sscanf(u.Port(), "%d", &pc.Port)
	}
	pc.User = u.User.Username()
	if pw, ok := u.User.Password(); ok {
		pc.Password = pw
	}
	pc.Database = cfg.Test.DatabaseName
	if v := u.Query().Get("sslmode"); v != "" {
		pc.SSLMode = v
	}
	// 测试串行执行，连接池不必太大
	pc.MinConns = 1
	pc.MaxConns = 5
	return pc
}

func joinTables(tables []string) string {
	out := ""
	for i, t := range tables {
		if i > 0 {
			out += ", "
		}
		out += t
	}
	return out
}
