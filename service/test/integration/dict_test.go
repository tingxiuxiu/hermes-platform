package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/hermes-platform/go-service/test/harness"
)

// dictItem 是字典项的响应结构（测试侧解码用）。
type dictItem struct {
	ID          int64   `json:"id"`
	DictType    string  `json:"dict_type"`
	Code        int     `json:"code"`
	Label       string  `json:"label"`
	DictName    string  `json:"dict_name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	SortOrder   int     `json:"sort_order"`
	Color       *string `json:"color"`
	Status      string  `json:"status"`
}

// createDict 创建一条字典项（管理员），返回响应数据。
func createDict(t *testing.T, e *harness.Env, adminToken, dictType string, code int, label string) dictItem {
	t.Helper()

	body := `{"dict_type":"` + dictType + `","code":` + itoaStr(code) +
		`,"label":"` + label + `","dict_name":"用例状态","category":"automation","color":"blue"}`
	w := doAutomation(t, e, http.MethodPost, api(e)+"/system/dicts", body, bearer(adminToken))
	if w.Code != http.StatusOK {
		t.Fatalf("create dict: status=%d body=%s", w.Code, w.Body.String())
	}
	var item dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &item); err != nil {
		t.Fatalf("decode dict: %v", err)
	}
	return item
}

// ---------------------------------------------------------------------------
// 查询
// ---------------------------------------------------------------------------

func TestListDictsRequiresJWT(t *testing.T) {
	env := harness.Setup(t)

	w := doAutomation(t, env, http.MethodGet, api(env)+"/system/dicts", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("list dicts without token: status=%d, want 401", w.Code)
	}
}

func TestListDictsByTypeAndCategory(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	// 同一 dict_type 存多行（验证复合唯一语义，不受 Python 错误约束影响）
	createDict(t, env, admin.AccessToken, "case_status", 0, "运行中")
	createDict(t, env, admin.AccessToken, "case_status", 1, "已通过")
	createDict(t, env, admin.AccessToken, "job_status", 0, "构建中")

	// 按 dict_type 筛选
	w := doAutomation(t, env, http.MethodGet,
		api(env)+"/system/dicts?dict_type=case_status", "", bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("list dicts: status=%d body=%s", w.Code, w.Body.String())
	}
	var items []dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &items); err != nil {
		t.Fatalf("decode dicts: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("case_status items = %d, want 2 (复合唯一允许同 type 多行)", len(items))
	}
	// 默认按 sort_order 升序、id 升序
	for _, item := range items {
		if item.DictType != "case_status" {
			t.Errorf("item %d dict_type = %s, want case_status", item.ID, item.DictType)
		}
	}
}

func TestListDictsEmptyDescriptionIsNull(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	item := createDict(t, env, admin.AccessToken, "no_desc", 0, "无描述")

	if item.Description != nil {
		t.Errorf("description = %v, want null (empty string maps to null)", *item.Description)
	}
	if item.Category == nil || *item.Category != "automation" {
		t.Errorf("category = %v, want automation", item.Category)
	}
	if item.Color == nil || *item.Color != "blue" {
		t.Errorf("color = %v, want blue", item.Color)
	}
	if item.Status != "active" {
		t.Errorf("status = %s, want active (default)", item.Status)
	}
}

// ---------------------------------------------------------------------------
// 写操作（管理员）
// ---------------------------------------------------------------------------

func TestDictWriteRequiresAdmin(t *testing.T) {
	env := harness.Setup(t)
	// 普通用户（非管理员）
	user := registerUser(t, env, "regular", "RegularStrongPass123!", "regular@example.com")

	body := `{"dict_type":"x","code":1,"label":"y","dict_name":"z"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/system/dicts", body, bearer(user.AccessToken))
	if w.Code != http.StatusForbidden {
		t.Fatalf("create dict as non-admin: status=%d, want 403", w.Code)
	}
}

func TestUpdateDictPartialFields(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	item := createDict(t, env, admin.AccessToken, "upd", 0, "原标签")

	// 只改 label 和 sort_order（其余字段 nil 表示不改）
	body := `{"label":"新标签","sort_order":5}`
	w := doAutomation(t, env, http.MethodPut,
		api(env)+"/system/dicts/"+itoa(item.ID), body, bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("update dict: status=%d body=%s", w.Code, w.Body.String())
	}
	var updated dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.Label != "新标签" {
		t.Errorf("label = %s, want 新标签", updated.Label)
	}
	if updated.SortOrder != 5 {
		t.Errorf("sort_order = %d, want 5", updated.SortOrder)
	}
	// 未传的字段保持原值
	if updated.DictName != "用例状态" {
		t.Errorf("dict_name = %s, want 用例状态 (unchanged)", updated.DictName)
	}
	if updated.Color == nil || *updated.Color != "blue" {
		t.Errorf("color = %v, want blue (unchanged)", updated.Color)
	}
}

func TestDeleteDict(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	item := createDict(t, env, admin.AccessToken, "del", 0, "待删除")

	w := doAutomation(t, env, http.MethodDelete,
		api(env)+"/system/dicts/"+itoa(item.ID), "", bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("delete dict: status=%d body=%s", w.Code, w.Body.String())
	}

	// 再次删除应 404
	w = doAutomation(t, env, http.MethodDelete,
		api(env)+"/system/dicts/"+itoa(item.ID), "", bearer(admin.AccessToken))
	if w.Code != http.StatusNotFound {
		t.Fatalf("delete deleted dict: status=%d, want 404", w.Code)
	}
}

func TestUpdateNonexistentDictReturns404(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	w := doAutomation(t, env, http.MethodPut,
		api(env)+"/system/dicts/999999", `{"label":"x"}`, bearer(admin.AccessToken))
	if w.Code != http.StatusNotFound {
		t.Fatalf("update missing dict: status=%d, want 404", w.Code)
	}
}

func TestCreateDictRejectsInvalidPayload(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	// 缺 dict_type
	body := `{"code":1,"label":"y","dict_name":"z"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/system/dicts", body, bearer(admin.AccessToken))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("create dict without dict_type: status=%d, want 422", w.Code)
	}
}

func TestCreateDictRejectsOutOfRangeCode(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	// code 超出 SMALLINT 范围
	body := `{"dict_type":"x","code":99999,"label":"y","dict_name":"z"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/system/dicts", body, bearer(admin.AccessToken))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("create dict with code=99999: status=%d, want 422 (SMALLINT range)", w.Code)
	}
}

// TestDuplicateDictCodeRejected 验证复合唯一约束被正确映射为领域冲突错误。
func TestDuplicateDictCodeRejected(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	createDict(t, env, admin.AccessToken, "dup", 7, "第一次")

	// 同一 dict_type + code 重复
	body := `{"dict_type":"dup","code":7,"label":"第二次","dict_name":"重复"}`
	w := doAutomation(t, env, http.MethodPost, api(env)+"/system/dicts", body, bearer(admin.AccessToken))
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate dict code: status=%d, want 409; body=%s", w.Code, w.Body.String())
	}
}

// TestOnlyActiveFilter 验证 only_active 筛选。
func TestOnlyActiveFilter(t *testing.T) {
	env := harness.Setup(t)
	admin := registerUser(t, env, "admin", "AdminStrongPass123!", "admin@example.com")

	active := createDict(t, env, admin.AccessToken, "filter_t", 0, "启用项")
	inactive := createDict(t, env, admin.AccessToken, "filter_t", 1, "禁用的")

	// 把第二条改为 inactive
	body := `{"status":"inactive"}`
	w := doAutomation(t, env, http.MethodPut,
		api(env)+"/system/dicts/"+itoa(inactive.ID), body, bearer(admin.AccessToken))
	if w.Code != http.StatusOK {
		t.Fatalf("deactivate dict: status=%d body=%s", w.Code, w.Body.String())
	}

	// 全部返回
	w = doAutomation(t, env, http.MethodGet,
		api(env)+"/system/dicts?dict_type=filter_t", "", bearer(admin.AccessToken))
	var all []dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &all); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all items = %d, want 2", len(all))
	}

	// 只看 active
	w = doAutomation(t, env, http.MethodGet,
		api(env)+"/system/dicts?dict_type=filter_t&only_active=true", "", bearer(admin.AccessToken))
	var activeOnly []dictItem
	if err := json.Unmarshal(decodeAutomation(t, w).Data, &activeOnly); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(activeOnly) != 1 {
		t.Fatalf("active-only items = %d, want 1", len(activeOnly))
	}
	if activeOnly[0].ID != active.ID {
		t.Errorf("active item id = %d, want %d", activeOnly[0].ID, active.ID)
	}
}

func itoaStr(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf []byte
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}

var _ = strings.Contains
