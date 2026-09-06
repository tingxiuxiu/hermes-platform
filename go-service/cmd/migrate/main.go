// Command migrate 是 golang-migrate 的 CLI 包装。
//
// 用法：
//
//	migrate -cmd up                 应用全部迁移（新库）
//	migrate -cmd down [-steps N]    回滚（默认回滚 1 步）
//	migrate -cmd version            打印当前版本与 dirty 状态
//	migrate -cmd force -version N   强制把版本号设为 N（用于存量库打基线）
//	migrate -cmd baseline           等价于 force 1：跳过建表，仅应用后续修复迁移
//
// 存量库（已由 alembic 建表）用 -cmd baseline，避免 000001 建表冲突。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"

	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/migrations"
)

func main() {
	cmd := flag.String("cmd", "up", "up | down | version | force | baseline")
	steps := flag.Int("steps", 1, "down 的步数")
	version := flag.Int("version", 1, "force 的目标版本号")
	path := flag.String("path", "migrations", "迁移文件目录")
	dsn := flag.String("dsn", "", "数据库连接串，缺省使用配置中的 Postgres DSN")
	flag.Parse()

	if err := run(*cmd, *steps, *version, *path, *dsn); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run(cmd string, steps, version int, path, dsn string) error {
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		dsn = cfg.Postgres.DSN()
	}

	if cmd == "up" {
		if err := migrations.ResetBookkeepingIfEmpty(dsn); err != nil {
			return err
		}
	}

	m, err := newMigrate(dsn, path)
	if err != nil {
		return err
	}

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && !isNoChange(err) {
			return err
		}
		fmt.Println("migrate: up completed")

	case "down":
		if err := m.Steps(-steps); err != nil && !isNoChange(err) {
			return err
		}
		fmt.Printf("migrate: down %d step(s) completed\n", steps)

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)

	case "force":
		if err := m.Force(version); err != nil {
			return err
		}
		fmt.Printf("migrate: forced version to %d\n", version)

	case "baseline":
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		present, err := migrations.HasInitSchema(ctx, dsn)
		if err != nil {
			return err
		}
		if !present {
			return fmt.Errorf("baseline is only for DBs that already have 000001 tables (roles, sys_dict, test_executions).\nThis database looks empty. From go-service run:\n  go run ./cmd/migrate -cmd up")
		}
		// 存量库打基线：把 000001 标记为已执行，然后继续应用后续修复迁移。
		if err := m.Force(1); err != nil {
			return fmt.Errorf("force baseline: %w", err)
		}
		fmt.Println("migrate: baseline set to 000001, applying remaining migrations")
		if err := m.Up(); err != nil && !isNoChange(err) {
			return err
		}
		fmt.Println("migrate: baseline + up completed")

	case "verify":
		return verify(dsn, path)

	default:
		return fmt.Errorf("unknown command %q", cmd)
	}

	v, dirty, _ := m.Version()
	fmt.Printf("current version=%d dirty=%v\n", v, dirty)
	return nil
}

// verify 在一次性的临时库上验证迁移可逆：建库 → up → down → up → 删库。
// 不依赖 psql，全部通过 pgx 完成，保证 CI 与本地行为一致。
func verify(dsn, path string) error {
	adminDSN, scratchDSN, err := splitScratchDSN(dsn, verifyDatabaseName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return fmt.Errorf("connect to maintenance db: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	dropScratch := func() error {
		// migrate 实例会保留连接，直接 DROP 会因「有会话在访问」失败（SQLSTATE 55006）。
		// 先终止该库上的所有后端连接，兼容不支持 DROP ... WITH (FORCE) 的旧版本。
		_, err := conn.Exec(ctx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity "+
				"WHERE datname = $1 AND pid <> pg_backend_pid()",
			verifyDatabaseName)
		if err != nil {
			return fmt.Errorf("terminate backends: %w", err)
		}
		if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+verifyDatabaseName); err != nil {
			return fmt.Errorf("drop scratch db: %w", err)
		}
		return nil
	}

	if err := dropScratch(); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+verifyDatabaseName); err != nil {
		return fmt.Errorf("create scratch db: %w", err)
	}
	defer func() {
		if err := dropScratch(); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: warning: %v\n", err)
		}
	}()

	fmt.Printf("migrate: verifying on scratch db %q\n", verifyDatabaseName)

	run := func(action string, fn func(*migrate.Migrate) error) error {
		m, err := newMigrate(scratchDSN, path)
		if err != nil {
			return fmt.Errorf("%s: %w", action, err)
		}
		if err := fn(m); err != nil && !isNoChange(err) {
			return fmt.Errorf("%s: %w", action, err)
		}
		v, dirty, _ := m.Version()
		fmt.Printf("migrate: %-6s -> version=%d dirty=%v\n", action, v, dirty)
		if dirty {
			return fmt.Errorf("%s: schema left in dirty state", action)
		}
		return nil
	}

	if err := run("up", func(m *migrate.Migrate) error { return m.Up() }); err != nil {
		return err
	}
	if err := run("down", func(m *migrate.Migrate) error { return m.Down() }); err != nil {
		return err
	}
	if err := run("up", func(m *migrate.Migrate) error { return m.Up() }); err != nil {
		return err
	}

	fmt.Println("migrate: up -> down -> up verified, no dirty state")
	return nil
}

const verifyDatabaseName = "hermes_migrate_verify"

// splitScratchDSN 把业务库 DSN 改写为：连到 postgres 维护库 + 指向临时库的 DSN。
func splitScratchDSN(dsn, scratchDB string) (adminDSN, scratch string, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("parse dsn: %w", err)
	}
	admin := *u
	admin.Path = "/postgres"

	target := *u
	target.Path = "/" + scratchDB

	return admin.String(), target.String(), nil
}

// newMigrate 构造 migrate 实例。
// scheme 转换与 Windows 路径处理统一由 platform/migrations 处理，
// 与 test/harness 保持同一套行为。
func newMigrate(dsn, path string) (*migrate.Migrate, error) {
	return migrations.New(path, dsn)
}

func isNoChange(err error) bool {
	return err == migrate.ErrNoChange
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}
