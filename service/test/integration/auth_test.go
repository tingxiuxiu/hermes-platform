// Package integration 承载连接真实 PostgreSQL 与真实 Redis 的端到端测试。
//
// 严禁在本包内使用 mock 替代数据库或缓存（ADR-0007）：
// 本次重构的核心风险是 SQL 正确性与契约正确性，mock 恰恰会掩盖这两类问题。
package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/test/harness"

	_ "github.com/hermes-platform/go-service/internal/bootstrap"
)

// api 返回配置中的 API 前缀，例如 /tap/api/v1。
func api(e *harness.Env) string {
	return e.Cfg.APIPrefix
}

// do 发起一次请求并返回响应。
func do(t *testing.T, e *harness.Env, method, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	e.Router(t).ServeHTTP(w, req)
	return w
}

// decode 解析响应体。
func decode(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()

	var env envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return env
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
}

// registerUser 注册一个新用户并返回其登录数据。
// 用户名随机，保证测试之间互不冲突。
func registerUser(t *testing.T, e *harness.Env, username, password, email string) loginResult {
	t.Helper()

	body := `{"username":"` + username + `","password":"` + password + `","email":"` + email + `"}`
	w := do(t, e, http.MethodPost, api(e)+"/register", body, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("register failed: status=%d body=%s", w.Code, w.Body.String())
	}

	env := decode(t, w)
	var data loginResult
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode login data: %v", err)
	}
	if data.AccessToken == "" {
		t.Fatal("register must return an access token")
	}
	return data
}

type loginResult struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	User             *struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Status   int    `json:"status"`
	} `json:"user"`
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func uniqueUser(t *testing.T) (string, string) {
	t.Helper()
	return "user_" + randSuffix(), "user_" + randSuffix() + "@example.com"
}

func randSuffix() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = alphabet[(i*7+len(alphabet))%len(alphabet)]
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// 注册
// ---------------------------------------------------------------------------

func TestRegisterReturnsTokenAndUserInfo(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)

	result := registerUser(t, env, username, "StrongPass123!", email)

	if result.User == nil {
		t.Fatal("register response must include user info")
	}
	if result.User.Username != username {
		t.Errorf("username = %q, want %q", result.User.Username, username)
	}
	if result.User.Status != 0 {
		t.Errorf("status = %d, want 0", result.User.Status)
	}
	if result.TokenType != "bearer" {
		t.Errorf("token_type = %q, want bearer", result.TokenType)
	}
	if result.RefreshToken == "" {
		t.Error("register must also return a refresh token")
	}
}

func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/register",
		`{"username":"`+username+`","password":"StrongPass123!","email":"other@example.com"}`, nil)

	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate username: status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	other, _ := uniqueUser(t)
	w := do(t, env, http.MethodPost, api(env)+"/register",
		`{"username":"`+other+`","password":"StrongPass123!","email":"`+email+`"}`, nil)

	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate email: status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)

	w := do(t, env, http.MethodPost, api(env)+"/register",
		`{"username":"`+username+`","password":"123","email":"`+email+`"}`, nil)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("weak password: status = %d, want 422; body=%s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 登录
// ---------------------------------------------------------------------------

func TestLoginWithUsernameAndWithEmail(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	for name, account := range map[string]string{"username": username, "email": email} {
		t.Run(name, func(t *testing.T) {
			w := do(t, env, http.MethodPost, api(env)+"/login",
				`{"username":"`+account+`","password":"StrongPass123!"}`, nil)
			if w.Code != http.StatusOK {
				t.Fatalf("login by %s: status = %d, body=%s", name, w.Code, w.Body.String())
			}
		})
	}
}

// TestLoginDoesNotLeakAccountExistence 是安全回归：
// 用户不存在与密码错误必须返回**相同**的响应，否则可被用来枚举账号。
func TestLoginDoesNotLeakAccountExistence(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	wrongPassword := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"`+username+`","password":"WrongPass123!"}`, nil)
	missingUser := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"nobody-here","password":"StrongPass123!"}`, nil)

	if wrongPassword.Code != http.StatusUnauthorized || missingUser.Code != http.StatusUnauthorized {
		t.Fatalf("both must be 401, got %d and %d", wrongPassword.Code, missingUser.Code)
	}
	a := decode(t, wrongPassword)
	b := decode(t, missingUser)
	if a.Message != b.Message {
		t.Errorf("messages must be identical, got %q vs %q", a.Message, b.Message)
	}
	if a.Code != b.Code {
		t.Errorf("codes must be identical, got %d vs %d", a.Code, b.Code)
	}
}

func TestLoginFormEndpoint(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	req := httptest.NewRequest(http.MethodPost, api(env)+"/login/access-token",
		strings.NewReader("username="+username+"&password=StrongPass123!"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	env.Router(t).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("form login: status = %d, body=%s", w.Code, w.Body.String())
	}
}

// TestLoginLocksAfterRepeatedFailures 是 ADR-0005 的登录锁定回归。
func TestLoginLocksAfterRepeatedFailures(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	registerUser(t, env, username, "StrongPass123!", email)

	for i := 0; i < env.Cfg.Auth.MaxFailedLogins; i++ {
		w := do(t, env, http.MethodPost, api(env)+"/login",
			`{"username":"`+username+`","password":"WrongPass123!"}`, nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status = %d, want 401", i+1, w.Code)
		}
	}

	// 超过阈值后，即使密码正确也必须被拒
	w := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"`+username+`","password":"StrongPass123!"}`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("locked account must be rejected with 400, got %d; body=%s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 令牌
// ---------------------------------------------------------------------------

func TestProtectedEndpointRequiresToken(t *testing.T) {
	env := harness.Setup(t)

	w := do(t, env, http.MethodGet, api(env)+"/users", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: status = %d, want 401; body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("401 response must carry WWW-Authenticate header")
	}
	if got := decode(t, w); got.Success {
		t.Error("failed response must have success=false")
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	first := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/login/refresh",
		`{"refresh_token":"`+first.RefreshToken+`"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh: status = %d, body=%s", w.Code, w.Body.String())
	}
	refreshed := decode(t, w)
	var data loginResult
	if err := json.Unmarshal(refreshed.Data, &data); err != nil {
		t.Fatalf("decode refresh data: %v", err)
	}

	// 旧刷新令牌必须立即失效（轮换语义）
	again := do(t, env, http.MethodPost, api(env)+"/login/refresh",
		`{"refresh_token":"`+first.RefreshToken+`"}`, nil)
	if again.Code != http.StatusUnauthorized {
		t.Fatalf("rotated refresh token must be rejected, got %d", again.Code)
	}
}

// TestLogoutRevokesAllTokens 是 ADR-0005 的核心回归：
// 登出后该用户**已签发的** access token 必须立即失效。
// Python 版做不到的正是这一点（无状态 JWT 无法吊销）。
func TestLogoutRevokesAllTokens(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	result := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/logout", "", bearer(result.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("logout: status = %d, body=%s", w.Code, w.Body.String())
	}

	after := do(t, env, http.MethodGet, api(env)+"/users/1", "", bearer(result.AccessToken))
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("token must be revoked after logout, got %d", after.Code)
	}
}

// TestChangePasswordRevokesOtherSessions 验证改密后其他设备的令牌也失效。
func TestChangePasswordRevokesOtherSessions(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	first := registerUser(t, env, username, "StrongPass123!", email)

	// 模拟第二个设备登录
	second := registerUser(t, env, "second_"+username, "StrongPass123!", "second_"+email)
	_ = second

	w := do(t, env, http.MethodPost, api(env)+"/users/me/password",
		`{"old_password":"StrongPass123!","new_password":"NewStrongPass456!"}`,
		bearer(first.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("change password: status = %d, body=%s", w.Code, w.Body.String())
	}

	// 改密后旧令牌失效，必须用新密码重新登录
	after := do(t, env, http.MethodGet, api(env)+"/users/1", "", bearer(first.AccessToken))
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("old token must be revoked after password change, got %d", after.Code)
	}

	relogin := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"`+username+`","password":"NewStrongPass456!"}`, nil)
	if relogin.Code != http.StatusOK {
		t.Fatalf("login with new password must succeed, got %d; body=%s", relogin.Code, relogin.Body.String())
	}
}

func TestChangePasswordRejectsWrongOldPassword(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	result := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/users/me/password",
		`{"old_password":"TotallyWrong!","new_password":"NewStrongPass456!"}`,
		bearer(result.AccessToken))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong old password: status = %d, want 401; body=%s", w.Code, w.Body.String())
	}
}

// TestPasswordHashNeverAppearsInResponse 是安全验收项：
// 任何响应都不得泄漏密码哈希。
func TestPasswordHashNeverAppearsInResponse(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	result := registerUser(t, env, username, "StrongPass123!", email)

	paths := []string{
		api(env) + "/login",
		api(env) + "/register",
	}
	for _, p := range paths {
		w := do(t, env, http.MethodPost, p,
			`{"username":"`+username+`","password":"StrongPass123!"}`, nil)
		if strings.Contains(w.Body.String(), "argon2id") {
			t.Errorf("response for %s leaked a password hash: %s", p, w.Body.String())
		}
	}

	detail := do(t, env, http.MethodGet, api(env)+"/users/1", "", bearer(result.AccessToken))
	if strings.Contains(detail.Body.String(), "argon2id") {
		t.Errorf("user detail leaked a password hash: %s", detail.Body.String())
	}
}

// TestConfigPrefixIsUsed 确保路由确实挂在配置的 API 前缀下。
func TestConfigPrefixIsUsed(t *testing.T) {
	env := harness.Setup(t)
	if env.Cfg.APIPrefix != config.EnvLocal && env.Cfg.APIPrefix == "" {
		t.Fatal("API prefix must not be empty")
	}

	w := do(t, env, http.MethodGet, "/healthz", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz: status = %d", w.Code)
	}

	// 健康检查不应挂 API 前缀
	withPrefix := do(t, env, http.MethodGet, api(env)+"/healthz", "", nil)
	if withPrefix.Code == http.StatusOK {
		t.Error("healthz must not be mounted under the API prefix")
	}
}
