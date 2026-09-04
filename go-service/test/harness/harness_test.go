package harness

import (
	"context"
	"os"
	"testing"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

// TestResetTestDatabase 强制重建测试库。
//
// 必须**手动触发**：默认 go test 会跳过，只有设置 HERMES_RESET_TEST_DB=1 才执行。
//
// 原因：这个测试会 DROP DATABASE，而 `go test ./...` 默认并行运行多个包，
// 本包的 DROP 会把同样使用 hermes_test 的 test/integration 全部打挂
// （表现为 "database hermes_test does not exist" 或 "terminating connection due to
// administrator command"）。重建库是改了 migrations/ 之后的运维操作，
// 不该出现在常规测试运行里。
//
// 触发方式：./scripts/tasks.sh testdb-reset
func TestResetTestDatabase(t *testing.T) {
	if os.Getenv("HERMES_RESET_TEST_DB") != "1" {
		t.Skip("set HERMES_RESET_TEST_DB=1 to force-recreate the test database")
	}
	if err := ResetTestDatabase(); err != nil {
		t.Fatalf("reset test database: %v", err)
	}
	t.Log("test database recreated and migrated")
}

// TestSetupTruncatesBetweenTests 验证 harness 的隔离能力：
// 两个测试共用同一个进程，但彼此看不到对方写入的数据。
func TestSetupTruncatesBetweenTests(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	if _, err := env.DB.Exec(ctx,
		`INSERT INTO roles (code, name, is_system) VALUES ('probe', 'probe role', false)`,
	); err != nil {
		t.Fatalf("insert probe row: %v", err)
	}

	var count int
	if err := env.DB.QueryRow(ctx, `SELECT count(*) FROM roles`).Scan(&count); err != nil {
		t.Fatalf("count roles: %v", err)
	}
	// 2 个基线角色（admin/viewer）+ 本次插入的 probe
	if count != 3 {
		t.Fatalf("expected 3 roles after insert, got %d", count)
	}
}

func TestSetupTruncatesBetweenTests_Second(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	var count int
	if err := env.DB.QueryRow(ctx, `SELECT count(*) FROM roles`).Scan(&count); err != nil {
		t.Fatalf("count roles: %v", err)
	}
	// 上一个测试插入的 probe 必须已被清掉，且基线种子被重建
	if count != 2 {
		t.Fatalf("isolation broken: expected 2 baseline roles, got %d", count)
	}

	var probe int
	if err := env.DB.QueryRow(ctx,
		`SELECT count(*) FROM roles WHERE code = 'probe'`).Scan(&probe); err != nil {
		t.Fatalf("count probe roles: %v", err)
	}
	if probe != 0 {
		t.Fatalf("isolation broken: probe role leaked across tests")
	}
}

// TestBaselineSeedIsRestored 验证 TRUNCATE 之后基线种子被重建：
// 后续用例依赖 admin 角色与权限码存在。
func TestBaselineSeedIsRestored(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	for _, code := range []string{"admin", "viewer"} {
		var exists bool
		if err := env.DB.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM roles WHERE code = $1)`, code).Scan(&exists); err != nil {
			t.Fatalf("check role %s: %v", code, err)
		}
		if !exists {
			t.Errorf("baseline role %q missing after truncate", code)
		}
	}

	var bound int
	if err := env.DB.QueryRow(ctx, `
		SELECT count(*) FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		WHERE r.code = 'admin'`).Scan(&bound); err != nil {
		t.Fatalf("count admin permissions: %v", err)
	}
	if bound != len(identity.SystemPermissions) {
		t.Errorf("admin should hold all %d permissions, got %d",
			len(identity.SystemPermissions), bound)
	}
}

// TestSchemaIsComplete 断言 15 张表全部就绪，
// 防止迁移漏建表时后续测试以难懂的方式失败。
func TestSchemaIsComplete(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	for _, table := range allTables {
		var exists bool
		err := env.DB.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
			table).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s is missing from the test schema", table)
		}
	}
}

// TestSysDictAllowsMultipleCodes 是缺陷 S-1 的回归测试：
// alembic 曾给 sys_dict 加了单列 UNIQUE(dict_type)，导致同一字典类型只能存一行。
func TestSysDictAllowsMultipleCodes(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	_, err := env.DB.Exec(ctx, `
		INSERT INTO sys_dict (dict_type, code, label, dict_name) VALUES
			('case_status', 0, 'running', '用例状态'),
			('case_status', 1, 'passed',  '用例状态'),
			('case_status', 2, 'failed',  '用例状态')
	`)
	if err != nil {
		t.Fatalf("inserting multiple codes for the same dict_type must succeed: %v", err)
	}

	var count int
	if err := env.DB.QueryRow(ctx,
		`SELECT count(*) FROM sys_dict WHERE dict_type = 'case_status'`).Scan(&count); err != nil {
		t.Fatalf("count dict entries: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 dict entries, got %d", count)
	}
}

// TestAttachmentSupportsItemAndStep 是评审确认项 Q1 的回归测试：
// execution_case_step_attachments 必须同时支持 case 级（item_id）与 step 级（step_id）附件。
func TestAttachmentSupportsItemAndStep(t *testing.T) {
	env := Setup(t)
	ctx := context.Background()

	_, err := env.DB.Exec(ctx, `
		INSERT INTO execution_case_step_attachments
			(item_id, step_id, attachment_type, file_name, url) VALUES
			(100, NULL, 'screenshot', 'case-level.png', 'https://example.com/case.png'),
			(NULL, 200, 'log',        'step-level.log',  'https://example.com/step.log')
	`)
	if err != nil {
		t.Fatalf("attachment table must accept both item-scoped and step-scoped rows: %v", err)
	}
}
