// Package database 封装 pgx 连接池与事务管理。
//
// 事务边界由 application 层通过 TxManager 声明（ADR-0004 关键纪律第 3 条），
// Repository 方法统一接收 DBTX 接口，从而同时兼容连接池与事务。
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hermes-platform/go-service/internal/platform/config"
)

// DBTX 是 pgx 的查询执行抽象。
// *pgxpool.Pool 与 pgx.Tx 都实现了它，sqlc 生成的代码也接受同样的接口，
// 因此同一个查询函数可以在「有事务」与「无事务」两种场景下复用。
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// TxManager 声明事务边界。
type TxManager interface {
	// WithTx 在单个事务内执行 fn。
	// fn 返回 error 时回滚，返回 nil 时提交；panic 会被恢复并转为 error 回滚。
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Pool 是带健康检查与生命周期管理的连接池。
type Pool struct {
	*pgxpool.Pool
}

// New 建立连接池并验证连通性。
func New(ctx context.Context, cfg config.PostgresConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: parse dsn: %w", err)
	}

	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return &Pool{Pool: pool}, nil
}

type txKey struct{}

// WithTx 在事务中执行 fn，并把事务注入 ctx。
// Repository 通过 Executor(ctx) 取到该事务，无需显式传参。
func (p *Pool) WithTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("database: begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			err = fmt.Errorf("database: panic in tx: %v", p)
			return
		}
		if err != nil {
			// 上下文取消时 Rollback 必然失败，不应覆盖原始错误
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return
		}
		if cerr := tx.Commit(ctx); cerr != nil {
			err = fmt.Errorf("database: commit tx: %w", cerr)
		}
	}()

	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return nil
}

// Executor 返回当前上下文的执行器：
// 事务中返回 pgx.Tx，否则返回连接池。
// 这让 Repository 无需关心自己是否处在事务中。
func Executor(ctx context.Context, pool *Pool) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return pool
}

// TxFrom 显式取出上下文中的事务，不存在时返回 nil。
func TxFrom(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx)
	return tx
}

// InTx 判断当前上下文是否处于事务中。
func InTx(ctx context.Context) bool {
	return TxFrom(ctx) != nil
}

// IsUniqueViolation 判断是否为唯一约束冲突。
// 用于把数据库约束错误翻译成领域冲突错误（如用户名重复）。
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// IsForeignKeyViolation 判断是否为外键约束冲突。
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

// IsNoRows 判断是否查询无结果。
func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
