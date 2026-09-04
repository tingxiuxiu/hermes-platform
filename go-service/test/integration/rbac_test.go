package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hermes-platform/go-service/test/harness"
)

// adminEnv 建立一个带管理员的测试环境，返回其令牌。
// 管理员账号用固定用户名 "admin"——这正是 Python 侧的判定规则之一。
func adminEnv(t *testing.T) (*harness.Env, string) {
	t.Helper()

	env := harness.Setup(t)
	result := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")
	return env, result.AccessToken
}

func TestNonAdminCannotListUsers(t *testing.T) {
	env := harness.Setup(t)
	username, email := uniqueUser(t)
	result := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodGet, api(env)+"/users", "", bearer(result.AccessToken))
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin listing users: status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestAdminCanListUsers(t *testing.T) {
	env, token := adminEnv(t)

	w := do(t, env, http.MethodGet, api(env)+"/users", "", bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("admin listing users: status = %d, body=%s", w.Code, w.Body.String())
	}

	env2 := decode(t, w)
	var data struct {
		Total    int               `json:"total"`
		Page     int               `json:"page"`
		PageSize int               `json:"page_size"`
		Items    []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(env2.Data, &data); err != nil {
		t.Fatalf("decode user list: %v", err)
	}
	if data.Total != 1 {
		t.Errorf("total = %d, want 1", data.Total)
	}
	if len(data.Items) != 1 {
		t.Errorf("items = %d, want 1", len(data.Items))
	}
}

// TestUserListExcludesSoftDeleted 是核心业务规则回归：
// 软删除（status=2）的用户不得出现在任何列表中。
func TestUserListExcludesSoftDeleted(t *testing.T) {
	env, token := adminEnv(t)
	username, email := uniqueUser(t)
	victim := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/users/batch-delete",
		`{"user_ids":[`+itoa(victim.User.ID)+`]}`, bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("batch delete: status = %d, body=%s", w.Code, w.Body.String())
	}

	list := do(t, env, http.MethodGet, api(env)+"/users", "", bearer(token))
	body := list.Body.String()
	if containsUserID(body, victim.User.ID) {
		t.Errorf("soft-deleted user %d must not appear in the list: %s", victim.User.ID, body)
	}

	// 详情接口同样必须视为不存在
	detail := do(t, env, http.MethodGet, api(env)+"/users/"+itoa(victim.User.ID), "", bearer(token))
	if detail.Code != http.StatusNotFound {
		t.Fatalf("soft-deleted user detail: status = %d, want 404", detail.Code)
	}
}

func TestAdminCanDisableAndEnableUser(t *testing.T) {
	env, token := adminEnv(t)
	username, email := uniqueUser(t)
	target := registerUser(t, env, username, "StrongPass123!", email)

	disable := do(t, env, http.MethodPut, api(env)+"/users/"+itoa(target.User.ID)+"/status",
		`{"user_id":`+itoa(target.User.ID)+`,"status":1}`, bearer(token))
	if disable.Code != http.StatusOK {
		t.Fatalf("disable: status = %d, body=%s", disable.Code, disable.Body.String())
	}

	// 被禁用后该用户的令牌立即失效
	after := do(t, env, http.MethodGet, api(env)+"/users/1", "", bearer(target.AccessToken))
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user token must be revoked, got %d", after.Code)
	}

	// 重新登录应被拒（400 Inactive user，对齐 Python）
	login := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"`+username+`","password":"StrongPass123!"}`, nil)
	if login.Code != http.StatusBadRequest {
		t.Fatalf("disabled user login: status = %d, want 400; body=%s", login.Code, login.Body.String())
	}

	enable := do(t, env, http.MethodPut, api(env)+"/users/"+itoa(target.User.ID)+"/status",
		`{"user_id":`+itoa(target.User.ID)+`,"status":0}`, bearer(token))
	if enable.Code != http.StatusOK {
		t.Fatalf("enable: status = %d, body=%s", enable.Code, enable.Body.String())
	}
}

func TestAdminCanResetPassword(t *testing.T) {
	env, token := adminEnv(t)
	username, email := uniqueUser(t)
	target := registerUser(t, env, username, "StrongPass123!", email)

	w := do(t, env, http.MethodPost, api(env)+"/users/"+itoa(target.User.ID)+"/reset-password",
		`{"user_id":`+itoa(target.User.ID)+`,"new_password":"ResetPass456!"}`, bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("reset password: status = %d, body=%s", w.Code, w.Body.String())
	}

	// 旧令牌必须失效
	after := do(t, env, http.MethodGet, api(env)+"/users/1", "", bearer(target.AccessToken))
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("token must be revoked after reset, got %d", after.Code)
	}

	login := do(t, env, http.MethodPost, api(env)+"/login",
		`{"username":"`+username+`","password":"ResetPass456!"}`, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login with reset password must succeed, got %d; body=%s", login.Code, login.Body.String())
	}
}

func TestRoleOptionsReturnsRolesNotUsers(t *testing.T) {
	env, token := adminEnv(t)

	w := do(t, env, http.MethodGet, api(env)+"/users/role-options", "", bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("role options: status = %d, body=%s", w.Code, w.Body.String())
	}

	env2 := decode(t, w)
	var roles []struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(env2.Data, &roles); err != nil {
		t.Fatalf("decode role options: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 baseline roles (admin/viewer), got %d", len(roles))
	}

	codes := map[string]bool{}
	for _, r := range roles {
		codes[r.Code] = true
	}
	for _, want := range []string{"admin", "viewer"} {
		if !codes[want] {
			t.Errorf("baseline role %q missing from role-options", want)
		}
	}
}

func TestAdminCanAssignRoles(t *testing.T) {
	env, token := adminEnv(t)
	username, email := uniqueUser(t)
	target := registerUser(t, env, username, "StrongPass123!", email)

	// 先拿到角色列表
	rolesW := do(t, env, http.MethodGet, api(env)+"/roles", "", bearer(token))
	var roles []struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	rawRoles := decode(t, rolesW)
	if err := json.Unmarshal(rawRoles.Data, &roles); err != nil {
		t.Fatalf("decode roles: %v", err)
	}
	if len(roles) == 0 {
		t.Fatal("baseline roles must exist")
	}

	assign := do(t, env, http.MethodPut, api(env)+"/users/"+itoa(target.User.ID)+"/roles",
		`{"user_id":`+itoa(target.User.ID)+`,"role_ids":[`+itoa(roles[0].ID)+`]}`, bearer(token))
	if assign.Code != http.StatusOK {
		t.Fatalf("assign roles: status = %d, body=%s", assign.Code, assign.Body.String())
	}

	detail := do(t, env, http.MethodGet, api(env)+"/users/"+itoa(target.User.ID), "", bearer(token))
	var detailData struct {
		Roles []struct {
			ID   int64  `json:"id"`
			Code string `json:"code"`
		} `json:"roles"`
	}
	if err := json.Unmarshal(decode(t, detail).Data, &detailData); err != nil {
		t.Fatalf("decode user detail: %v", err)
	}
	if len(detailData.Roles) != 1 || detailData.Roles[0].ID != roles[0].ID {
		t.Errorf("assigned roles = %+v, want [%d]", detailData.Roles, roles[0].ID)
	}
}

func TestSystemRoleCannotBeDeleted(t *testing.T) {
	env, token := adminEnv(t)

	rolesW := do(t, env, http.MethodGet, api(env)+"/roles", "", bearer(token))
	var roles []struct {
		ID       int64  `json:"id"`
		Code     string `json:"code"`
		IsSystem bool   `json:"is_system"`
	}
	if err := json.Unmarshal(decode(t, rolesW).Data, &roles); err != nil {
		t.Fatalf("decode roles: %v", err)
	}

	for _, r := range roles {
		if !r.IsSystem {
			continue
		}
		w := do(t, env, http.MethodDelete, api(env)+"/roles/"+itoa(r.ID), "", bearer(token))
		if w.Code != http.StatusForbidden {
			t.Errorf("deleting system role %q: status = %d, want 403; body=%s",
				r.Code, w.Code, w.Body.String())
		}
	}
}

func TestAdminCanCreateUpdateAndDeleteCustomRole(t *testing.T) {
	env, token := adminEnv(t)

	create := do(t, env, http.MethodPost, api(env)+"/roles",
		`{"code":"editor","name":"编辑人员","description":"可以编辑内容"}`, bearer(token))
	if create.Code != http.StatusOK {
		t.Fatalf("create role: status = %d, body=%s", create.Code, create.Body.String())
	}

	var created struct {
		ID       int64  `json:"id"`
		Code     string `json:"code"`
		IsSystem bool   `json:"is_system"`
	}
	if err := json.Unmarshal(decode(t, create).Data, &created); err != nil {
		t.Fatalf("decode created role: %v", err)
	}
	if created.IsSystem {
		t.Error("runtime-created roles must never be system roles")
	}

	// 重复 code 必须冲突
	dup := do(t, env, http.MethodPost, api(env)+"/roles",
		`{"code":"editor","name":"编辑人员2"}`, bearer(token))
	if dup.Code != http.StatusConflict {
		t.Fatalf("duplicate role code: status = %d, want 409", dup.Code)
	}

	update := do(t, env, http.MethodPut, api(env)+"/roles/"+itoa(created.ID),
		`{"name":"内容编辑"}`, bearer(token))
	if update.Code != http.StatusOK {
		t.Fatalf("update role: status = %d, body=%s", update.Code, update.Body.String())
	}

	del := do(t, env, http.MethodDelete, api(env)+"/roles/"+itoa(created.ID), "", bearer(token))
	if del.Code != http.StatusOK {
		t.Fatalf("delete custom role: status = %d, body=%s", del.Code, del.Body.String())
	}
}

func TestPermissionTreeAndList(t *testing.T) {
	env, token := adminEnv(t)

	list := do(t, env, http.MethodGet, api(env)+"/permissions", "", bearer(token))
	if list.Code != http.StatusOK {
		t.Fatalf("list permissions: status = %d, body=%s", list.Code, list.Body.String())
	}
	var perms []struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	if err := json.Unmarshal(decode(t, list).Data, &perms); err != nil {
		t.Fatalf("decode permissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("baseline permissions must exist")
	}

	tree := do(t, env, http.MethodGet, api(env)+"/permissions/tree", "", bearer(token))
	if tree.Code != http.StatusOK {
		t.Fatalf("permission tree: status = %d, body=%s", tree.Code, tree.Body.String())
	}
	var nodes []struct {
		Code     string `json:"code"`
		Children []struct {
			Code string `json:"code"`
		} `json:"children"`
	}
	if err := json.Unmarshal(decode(t, tree).Data, &nodes); err != nil {
		t.Fatalf("decode permission tree: %v", err)
	}
	if len(nodes) != len(perms) {
		t.Errorf("tree roots = %d, flat list = %d: flat permissions must all appear as roots",
			len(nodes), len(perms))
	}
}

func TestAdminCanCreateAndDeletePermission(t *testing.T) {
	env, token := adminEnv(t)

	create := do(t, env, http.MethodPost, api(env)+"/permissions",
		`{"code":"report:export","name":"导出报表","resource_type":3,"path":"/reports/export","method":"POST"}`,
		bearer(token))
	if create.Code != http.StatusOK {
		t.Fatalf("create permission: status = %d, body=%s", create.Code, create.Body.String())
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(decode(t, create).Data, &created); err != nil {
		t.Fatalf("decode created permission: %v", err)
	}

	dup := do(t, env, http.MethodPost, api(env)+"/permissions",
		`{"code":"report:export","name":"导出报表2"}`, bearer(token))
	if dup.Code != http.StatusConflict {
		t.Fatalf("duplicate permission code: status = %d, want 409", dup.Code)
	}

	del := do(t, env, http.MethodDelete, api(env)+"/permissions/"+itoa(created.ID), "", bearer(token))
	if del.Code != http.StatusOK {
		t.Fatalf("delete permission: status = %d, body=%s", del.Code, del.Body.String())
	}
}

// TestBatchStatusReturnsActualCount 对齐 Python 行为：
// 批量操作的返回文案必须反映**实际命中**的行数，而不是请求传入的长度。
func TestBatchStatusReturnsActualCount(t *testing.T) {
	env, token := adminEnv(t)

	// 传入一个不存在的 ID，实际只应命中 1 个（admin 自己）
	w := do(t, env, http.MethodPost, api(env)+"/users/batch-status",
		`{"user_ids":[1,999999],"status":1}`, bearer(token))
	if w.Code != http.StatusOK {
		t.Fatalf("batch status: status = %d, body=%s", w.Code, w.Body.String())
	}
	if got := decode(t, w).Message; got == "" {
		t.Fatal("batch status must return a message")
	}
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func containsUserID(body string, id int64) bool {
	return len(body) > 0 && (indexOf(body, `"id":`+itoa(id)) >= 0 ||
		indexOf(body, `"id": `+itoa(id)) >= 0)
}

func indexOf(s, sub string) int {
	if sub == "" {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
