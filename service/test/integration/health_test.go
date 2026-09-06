package integration

import (
	"net/http"
	"testing"

	"github.com/hermes-platform/go-service/test/harness"
)

// TestHealthz 验证存活探针：无依赖，直接返回 200。
func TestHealthz(t *testing.T) {
	env := harness.Setup(t)
	w := doAutomation(t, env, http.MethodGet, "/healthz", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz: status=%d, want 200", w.Code)
	}
}

// TestReadyz 验证就绪探针：ping PG + Redis，正常时返回 200。
func TestReadyz(t *testing.T) {
	env := harness.Setup(t)
	w := doAutomation(t, env, http.MethodGet, "/readyz", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("readyz: status=%d, want 200; body=%s", w.Code, w.Body.String())
	}
}

// TestLoginRefreshEndpoint 验证 Go 侧新增的 /login/refresh 已注册（契约 #5）。
func TestLoginRefreshEndpointRegistered(t *testing.T) {
	env := harness.Setup(t)

	// 无有效 token 访问 refresh 应被鉴权拒绝（非 404），证明路由已注册。
	// 具体状态码可能是 401（无凭证）或 422（凭证格式问题），都说明路由存在。
	w := doAutomation(t, env, http.MethodPost, api(env)+"/login/refresh", `{}`, nil)
	if w.Code == http.StatusNotFound {
		t.Fatal("login/refresh returned 404, route is NOT registered")
	}
}
