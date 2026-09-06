package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// initSchemaTables 是 000001 建出的探针表。三者都在才认为建表迁移真正执行过。
var initSchemaTables = []string{"roles", "sys_dict", "test_executions"}

// HasInitSchema 报告业务库是否已有 000001 的核心表。
func HasInitSchema(ctx context.Context, dsn string) (bool, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return false, fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	var n int
	err = conn.QueryRow(ctx, `
		SELECT count(*)
		  FROM information_schema.tables
		 WHERE table_schema = 'public'
		   AND table_name = ANY($1)`,
		initSchemaTables,
	).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("probe init schema: %w", err)
	}
	return n == len(initSchemaTables), nil
}

// ResetBookkeepingIfEmpty 在空库上丢掉误标的 golang-migrate 版本表。
//
// 误跑 baseline 会把 000001 标成已执行但并不建表，随后 000002 因
// sys_dict 不存在而 dirty。golang-migrate 的 Force(0) 需要 version 0
// 的 down 文件，不可用。空库直接 DROP schema_migrations 后再 Up。
// 已有 000001 表的库不会动版本表。
func ResetBookkeepingIfEmpty(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	present, err := HasInitSchema(ctx, dsn)
	if err != nil {
		return err
	}
	if present {
		return nil
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	if _, err := conn.Exec(ctx, `DROP TABLE IF EXISTS schema_migrations`); err != nil {
		return fmt.Errorf("drop schema_migrations: %w", err)
	}
	return nil
}
