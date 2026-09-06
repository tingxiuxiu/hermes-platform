// Package config 集中加载应用配置。
//
// 优先级：环境变量 > go-service/.env > 仓库根 .env > 代码默认值
// 沿用了 Python 版 (.env / docker-compose) 中的变量名，新增项全部带默认值，
// 保证现有 .env 不改也能启动（ADR-0003 / 01-architecture.md §7）。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是完整的应用配置。
type Config struct {
	ProjectName string
	Environment string
	APIPrefix   string

	HTTP       HTTPConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	Auth       AuthConfig
	Asynq      AsynqConfig
	Automation AutomationConfig
	Dashboard  DashboardConfig
	Log        LogConfig
	OTel       OTelConfig
	Test       TestConfig
	Bootstrap  BootstrapConfig
}

type HTTPConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ShutdownWait time.Duration
	CORSOrigins  []string
	FrontendHost string
}

type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MinConns        int32
	MaxConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DSN 返回 pgx 连接串。
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type RedisConfig struct {
	Host         string
	Port         int
	DB           int
	Password     string
	PoolSize     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Addr 返回 host:port。
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type AuthConfig struct {
	SecretKey         string
	Issuer            string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	UserCacheTTL      time.Duration
	MaxFailedLogins   int
	LoginLockDuration time.Duration
	ServiceToken      string
}

type AsynqConfig struct {
	Concurrency    int
	QueueDefault   string
	QueueDashboard string
	Retention      time.Duration
	ShutdownWait   time.Duration
}

type AutomationConfig struct {
	ServiceToken            string
	HeartbeatTimeoutSeconds int
}

type DashboardConfig struct {
	RefreshSeconds int
	TrendDays      int
}

type LogConfig struct {
	Level  string
	Format string
}

type OTelConfig struct {
	Endpoint   string
	Enabled    bool
	SampleRate float64
}

type TestConfig struct {
	DatabaseURL string
	// DatabaseName 允许每个测试包使用独立的数据库。
	// 多个测试包共用同一个库会在并行执行时互相 TRUNCATE 掉对方的数据。
	DatabaseName string
	RedisDB      int
}

type BootstrapConfig struct {
	FirstSuperuser         string
	FirstSuperuserPassword string
}

const (
	EnvLocal      = "local"
	EnvStaging    = "staging"
	EnvProduction = "production"
)

// Load 从 .env 与环境变量加载配置。
// 非 local 环境下敏感项仍为默认值时直接报错，对齐 Python 的
// _enforce_non_default_secrets 行为。
func Load() (Config, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	// 从当前目录逐级向上查找 .env。这样无论是 go run ./cmd/api（cwd=go-service）
	// 还是 go test ./internal/...（cwd=具体子包）都能读到仓库根的 .env。
	for _, dir := range envSearchPaths() {
		v.AddConfigPath(dir)
	}
	if err := v.ReadInConfig(); err != nil {
		// .env 缺失不致命，环境变量可能已全部提供
		_ = err
	}
	v.AutomaticEnv()

	cfg := Config{
		ProjectName: v.GetString("PROJECT_NAME"),
		Environment: v.GetString("ENVIRONMENT"),
		APIPrefix:   v.GetString("API_V1_STR"),
		HTTP: HTTPConfig{
			Port:         0,
			ReadTimeout:  0,
			WriteTimeout: 0,
			IdleTimeout:  0,
			ShutdownWait: 0,
			CORSOrigins:  nil,
			FrontendHost: "",
		},
		Postgres:   PostgresConfig{},
		Redis:      RedisConfig{},
		Auth:       AuthConfig{},
		Asynq:      AsynqConfig{},
		Automation: AutomationConfig{},
		Dashboard:  DashboardConfig{},
		Log:        LogConfig{},
		OTel:       OTelConfig{},
		Test:       TestConfig{},
		Bootstrap:  BootstrapConfig{},
	}

	// ---- 基础 ----
	v.SetDefault("PROJECT_NAME", "Hermes Platform")
	v.SetDefault("ENVIRONMENT", EnvLocal)
	v.SetDefault("API_V1_STR", "/tap/api/v1")
	cfg.ProjectName = v.GetString("PROJECT_NAME")
	cfg.Environment = normaliseEnv(v.GetString("ENVIRONMENT"))
	cfg.APIPrefix = v.GetString("API_V1_STR")

	// ---- HTTP ----
	v.SetDefault("SERVICE_PORT", 80)
	v.SetDefault("HTTP_READ_TIMEOUT", "15s")
	v.SetDefault("HTTP_WRITE_TIMEOUT", "15s")
	v.SetDefault("HTTP_IDLE_TIMEOUT", "60s")
	v.SetDefault("HTTP_SHUTDOWN_WAIT", "20s")
	cfg.HTTP.Port = v.GetInt("SERVICE_PORT")
	cfg.HTTP.ReadTimeout = v.GetDuration("HTTP_READ_TIMEOUT")
	cfg.HTTP.WriteTimeout = v.GetDuration("HTTP_WRITE_TIMEOUT")
	cfg.HTTP.IdleTimeout = v.GetDuration("HTTP_IDLE_TIMEOUT")
	cfg.HTTP.ShutdownWait = v.GetDuration("HTTP_SHUTDOWN_WAIT")
	cfg.HTTP.CORSOrigins = splitAndTrim(v.GetString("BACKEND_CORS_ORIGINS"))
	v.SetDefault("FRONTEND_HOST", "http://localhost:5173")
	cfg.HTTP.FrontendHost = v.GetString("FRONTEND_HOST")

	// ---- PostgreSQL ----
	// SQLALCHEMY_POOL_SIZE / SQLALCHEMY_MAX_OVERFLOW 沿用 Python 变量名：
	// pool_size 作为最小连接数，pool_size + max_overflow 作为最大连接数。
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_SSLMODE", "disable")
	v.SetDefault("SQLALCHEMY_POOL_SIZE", 10)
	v.SetDefault("SQLALCHEMY_MAX_OVERFLOW", 20)
	v.SetDefault("PG_MAX_CONN_LIFETIME", "30m")
	v.SetDefault("PG_MAX_CONN_IDLE_TIME", "5m")

	cfg.Postgres.Host = v.GetString("POSTGRES_SERVER")
	cfg.Postgres.Port = v.GetInt("POSTGRES_PORT")
	cfg.Postgres.User = v.GetString("POSTGRES_USER")
	cfg.Postgres.Password = v.GetString("POSTGRES_PASSWORD")
	cfg.Postgres.Database = v.GetString("POSTGRES_DB")
	cfg.Postgres.SSLMode = v.GetString("POSTGRES_SSLMODE")
	cfg.Postgres.MinConns = int32(v.GetInt("SQLALCHEMY_POOL_SIZE"))
	cfg.Postgres.MaxConns = int32(v.GetInt("PG_POOL_MAX_CONNS"))
	if cfg.Postgres.MaxConns == 0 {
		cfg.Postgres.MaxConns = int32(v.GetInt("SQLALCHEMY_POOL_SIZE")) +
			int32(v.GetInt("SQLALCHEMY_MAX_OVERFLOW"))
	}
	cfg.Postgres.MaxConnLifetime = v.GetDuration("PG_MAX_CONN_LIFETIME")
	cfg.Postgres.MaxConnIdleTime = v.GetDuration("PG_MAX_CONN_IDLE_TIME")

	// ---- Redis ----
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_POOL_SIZE", 10)
	v.SetDefault("REDIS_DIAL_TIMEOUT", "5s")
	v.SetDefault("REDIS_SOCKET_TIMEOUT", "5s")
	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.DB = v.GetInt("REDIS_DB")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.PoolSize = v.GetInt("REDIS_POOL_SIZE")
	cfg.Redis.DialTimeout = v.GetDuration("REDIS_DIAL_TIMEOUT")
	cfg.Redis.ReadTimeout = v.GetDuration("REDIS_SOCKET_TIMEOUT")
	cfg.Redis.WriteTimeout = v.GetDuration("REDIS_SOCKET_TIMEOUT")

	// ---- 鉴权（ADR-0005）----
	v.SetDefault("AUTH_ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("AUTH_REFRESH_TOKEN_TTL", "168h")
	v.SetDefault("AUTH_TOKEN_ISSUER", "hermes-platform")
	v.SetDefault("USER_CACHE_TTL", "300s")
	v.SetDefault("AUTH_MAX_FAILED_LOGINS", 5)
	v.SetDefault("AUTH_LOGIN_LOCK_DURATION", "15m")
	v.SetDefault("SECRET_KEY", "yqVeBt_e1CJeSWYel2DGOVJRxNZiTadofeIjmHE4jjw")
	cfg.Auth.SecretKey = v.GetString("SECRET_KEY")
	cfg.Auth.Issuer = v.GetString("AUTH_TOKEN_ISSUER")
	cfg.Auth.AccessTokenTTL = v.GetDuration("AUTH_ACCESS_TOKEN_TTL")
	cfg.Auth.RefreshTokenTTL = v.GetDuration("AUTH_REFRESH_TOKEN_TTL")
	cfg.Auth.UserCacheTTL = v.GetDuration("USER_CACHE_TTL")
	cfg.Auth.MaxFailedLogins = v.GetInt("AUTH_MAX_FAILED_LOGINS")
	cfg.Auth.LoginLockDuration = v.GetDuration("AUTH_LOGIN_LOCK_DURATION")
	cfg.Auth.ServiceToken = v.GetString("AUTOMATION_SERVICE_TOKEN")

	// ---- asynq ----
	v.SetDefault("ASYNQ_CONCURRENCY", 10)
	v.SetDefault("ASYNQ_QUEUE_DEFAULT", "default")
	v.SetDefault("ASYNQ_QUEUE_DASHBOARD", "dashboard")
	v.SetDefault("ASYNQ_RETENTION", "72h")
	v.SetDefault("ASYNQ_SHUTDOWN_WAIT", "20s")
	cfg.Asynq.Concurrency = v.GetInt("ASYNQ_CONCURRENCY")
	cfg.Asynq.QueueDefault = v.GetString("ASYNQ_QUEUE_DEFAULT")
	cfg.Asynq.QueueDashboard = v.GetString("ASYNQ_QUEUE_DASHBOARD")
	cfg.Asynq.Retention = v.GetDuration("ASYNQ_RETENTION")
	cfg.Asynq.ShutdownWait = v.GetDuration("ASYNQ_SHUTDOWN_WAIT")

	// ---- automation / dashboard ----
	v.SetDefault("AUTOMATION_SERVICE_TOKEN", "changethis")
	cfg.Automation.ServiceToken = v.GetString("AUTOMATION_SERVICE_TOKEN")
	v.SetDefault("AUTOMATION_HEARTBEAT_TIMEOUT_SECONDS", 60)
	cfg.Automation.HeartbeatTimeoutSeconds = v.GetInt("AUTOMATION_HEARTBEAT_TIMEOUT_SECONDS")
	v.SetDefault("AUTOMATION_DASHBOARD_REFRESH_SECONDS", 60)
	v.SetDefault("AUTOMATION_DASHBOARD_TREND_DAYS", 7)
	cfg.Dashboard.RefreshSeconds = v.GetInt("AUTOMATION_DASHBOARD_REFRESH_SECONDS")
	cfg.Dashboard.TrendDays = v.GetInt("AUTOMATION_DASHBOARD_TREND_DAYS")

	// ---- 日志 ----
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	cfg.Log.Level = v.GetString("LOG_LEVEL")
	cfg.Log.Format = v.GetString("LOG_FORMAT")

	// ---- 可观测性 ----
	v.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	v.SetDefault("OTEL_SAMPLE_RATE", 1.0)
	cfg.OTel.Endpoint = v.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
	cfg.OTel.Enabled = cfg.OTel.Endpoint != ""
	cfg.OTel.SampleRate = v.GetFloat64("OTEL_SAMPLE_RATE")

	// ---- 测试 ----
	// DatabaseName 可被每个测试包覆盖（TEST_DATABASE_NAME），
	// 让各包拥有独立数据库，避免并行执行时互相清数据。
	v.SetDefault("TEST_DATABASE_NAME", "hermes_test")
	v.SetDefault("TEST_DATABASE_URL", "")
	v.SetDefault("TEST_REDIS_DB", 15)
	cfg.Test.DatabaseName = v.GetString("TEST_DATABASE_NAME")
	cfg.Test.DatabaseURL = v.GetString("TEST_DATABASE_URL")
	if cfg.Test.DatabaseURL == "" {
		// 与 Postgres 配置同源，仅替换库名，保证本地零配置可用
		cfg.Test.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Postgres.User, cfg.Postgres.Password,
			cfg.Postgres.Host, cfg.Postgres.Port,
			cfg.Test.DatabaseName, cfg.Postgres.SSLMode,
		)
	}
	cfg.Test.RedisDB = v.GetInt("TEST_REDIS_DB")

	// ---- 首启管理员 ----
	cfg.Bootstrap.FirstSuperuser = v.GetString("FIRST_SUPERUSER")
	cfg.Bootstrap.FirstSuperuserPassword = v.GetString("FIRST_SUPERUSER_PASSWORD")

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// IsLocal 判断是否为本地环境。
func (c Config) IsLocal() bool { return c.Environment == EnvLocal }

// validate 校验必需项。非 local 环境下敏感项为默认值时报错。
func (c Config) validate() error {
	if c.Postgres.Host == "" {
		return fmt.Errorf("config: POSTGRES_SERVER is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("config: REDIS_HOST is required")
	}

	if c.Environment == EnvLocal {
		return nil
	}

	sensitive := map[string]string{
		"SECRET_KEY":               c.Auth.SecretKey,
		"POSTGRES_PASSWORD":        c.Postgres.Password,
		"AUTOMATION_SERVICE_TOKEN": c.Automation.ServiceToken,
		"FIRST_SUPERUSER_PASSWORD": c.Bootstrap.FirstSuperuserPassword,
	}
	for name, value := range sensitive {
		if value == "" || value == "changethis" {
			return fmt.Errorf(
				"config: %s keeps its default value, refusing to start in %q environment",
				name, c.Environment,
			)
		}
	}
	return nil
}

// envSearchPaths 返回从当前工作目录向上逐级的目录列表（最多 6 层）。
// viper 按顺序尝试，第一个命中的 .env 生效，因此更靠近 CWD 的配置优先级更高。
func envSearchPaths() []string {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{".", ".."}
	}

	paths := make([]string, 0, 7)
	dir := cwd
	for i := 0; i < 6; i++ {
		paths = append(paths, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return paths
}

func normaliseEnv(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "prod", "production":
		return EnvProduction
	case "stag", "staging":
		return EnvStaging
	default:
		return EnvLocal
	}
}

func splitAndTrim(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
