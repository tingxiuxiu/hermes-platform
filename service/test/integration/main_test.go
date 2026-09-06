package integration

import (
	"os"
	"testing"
)

// TestMain 为本测试包分配独立的数据库与 Redis DB。
//
// 为什么必须隔离：`go test ./...` 会**并行**运行多个包。
// test/harness 与 test/integration 若共用同一个库，
// 一个包在 cleanup 里的 TRUNCATE 会清掉另一个包正在断言的数据，
// 表现为随机、难以复现的失败。
//
// 同理，Redis 也必须分开：cleanup 里的 FLUSHDB 是整库清空，
// 会把另一个包刚写入的刷新令牌与登录计数一起抹掉。
func TestMain(m *testing.M) {
	os.Setenv("TEST_DATABASE_NAME", "hermes_integration")
	os.Setenv("TEST_REDIS_DB", "14")
	os.Exit(m.Run())
}
