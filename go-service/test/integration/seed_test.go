package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hermes-platform/go-service/internal/bootstrap"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/test/harness"
)

// TestSeedFirstSuperuser 验证首启超级管理员幂等创建。
func TestSeedFirstSuperuser(t *testing.T) {
	env := harness.Setup(t)

	// 配置 FIRST_SUPERUSER
	env.Cfg.Bootstrap.FirstSuperuser = "superadmin"
	env.Cfg.Bootstrap.FirstSuperuserPassword = "SuperSecret123!"

	// 执行 seed
	if err := bootstrap.SeedAll(context.Background(), env.App(t), env.Cfg); err != nil {
		t.Fatalf("seed all: %v", err)
	}

	// 用创建的超级管理员登录，验证可用
	body := `{"username":"superadmin","password":"SuperSecret123!"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/login", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login as seeded superadmin: status=%d body=%s", w.Code, w.Body.String())
	}

	// 验证是管理员（能访问需要 ADMIN 的接口）
	var login struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &login); err != nil {
		t.Fatalf("decode login: %v", err)
	}

	// 用 admin 权限访问 /roles（ADMIN）
	w2 := doAutomation(t, env, http.MethodGet, api(env)+"/roles", "", bearer(login.AccessToken))
	if w2.Code != http.StatusOK {
		t.Fatalf("access /roles as seeded superadmin: status=%d, want 200 (admin role)", w2.Code)
	}
}

// TestSeedIsIdempotent 验证 seed 可重复执行不报错。
func TestSeedIsIdempotent(t *testing.T) {
	env := harness.Setup(t)

	env.Cfg.Bootstrap.FirstSuperuser = "superadmin2"
	env.Cfg.Bootstrap.FirstSuperuserPassword = "SuperSecret123!"

	if err := bootstrap.SeedAll(context.Background(), env.App(t), env.Cfg); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	// 第二次执行（幂等）
	if err := bootstrap.SeedAll(context.Background(), env.App(t), env.Cfg); err != nil {
		t.Fatalf("second seed (idempotent): %v", err)
	}
}

// TestSeedSkipsWhenNotConfigured 验证未配置 FIRST_SUPERUSER 时跳过（不报错）。
func TestSeedSkipsWhenNotConfigured(t *testing.T) {
	env := harness.Setup(t)

	env.Cfg.Bootstrap.FirstSuperuser = ""
	env.Cfg.Bootstrap.FirstSuperuserPassword = ""
	if err := bootstrap.SeedAll(context.Background(), env.App(t), env.Cfg); err != nil {
		t.Fatalf("seed without FIRST_SUPERUSER must not error: %v", err)
	}
}

// envApp 暴露 env 的应用引用（harness 提供）。
var _ = config.EnvLocal
