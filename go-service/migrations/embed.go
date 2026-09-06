package migrations

import "embed"

// FS 内嵌本目录全部 golang-migrate SQL，供 API 启动时应用 schema，
// 不依赖进程工作目录是否在 go-service/。
//
//go:embed *.sql
var FS embed.FS
