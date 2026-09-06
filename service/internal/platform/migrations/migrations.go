// Package migrations 封装 golang-migrate 的构造细节，
// 让 cmd/migrate 与 test/harness 共用同一套行为（ADR-0008）。
package migrations

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // pgx5 驱动
)

// DriverScheme 是 golang-migrate 中 pgx/v5 驱动注册的名字。
// 它与业务 DSN 的 "postgres" scheme 不同，必须显式转换。
const DriverScheme = "pgx5"

// DatabaseURL 把业务 DSN 的 scheme 换成驱动注册名。
func DatabaseURL(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse dsn: %w", err)
	}
	u.Scheme = DriverScheme
	return u.String(), nil
}

// New 构造 migrate 实例。dir 是迁移文件目录，dsn 是业务库连接串。
//
// 数据源用 iofs 而不是 file：
// file 源在 Windows 上无法正确处理带盘符的绝对路径
// （"file://D:\..." 会在 URL 解析阶段因冒号被当作端口而失败，
// "file:///D:/..." 又会被解析成相对路径 "."）。iofs 直接吃 os.DirFS，
// 跨平台行为一致。
func New(dir, dsn string) (*migrate.Migrate, error) {
	dbURL, err := DatabaseURL(dsn)
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(os.DirFS(dir), ".")
	if err != nil {
		return nil, fmt.Errorf("open migrations source %q: %w", dir, err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		return nil, fmt.Errorf("create migrate: %w", err)
	}
	return m, nil
}

// NewFS 从任意 fs.FS 构造 migrate 实例（通常是 embed.FS）。
func NewFS(files fs.FS, dsn string) (*migrate.Migrate, error) {
	dbURL, err := DatabaseURL(dsn)
	if err != nil {
		return nil, err
	}
	source, err := iofs.New(files, ".")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		return nil, fmt.Errorf("create migrate: %w", err)
	}
	return m, nil
}

// Up 应用全部迁移。已处于最新版本时返回 nil（ErrNoChange 不视为错误）。
func Up(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// Down 回滚全部迁移。
func Down(m *migrate.Migrate) error {
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
