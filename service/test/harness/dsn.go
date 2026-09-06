package harness

import (
	"fmt"
	"net/url"
)

// parseURL 解析数据库连接串。
func parseURL(dsn string) (*url.URL, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	return u, nil
}

// withDatabaseName 返回把连接串的库名替换为 name 后的新连接串。
// 用于从业务库 DSN 派生出维护库（postgres）DSN，从而执行 CREATE/DROP DATABASE。
func withDatabaseName(dsn, name string) (string, error) {
	u, err := parseURL(dsn)
	if err != nil {
		return "", err
	}
	u.Path = "/" + name
	return u.String(), nil
}
