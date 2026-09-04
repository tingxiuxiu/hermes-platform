package identity_test

import (
	"testing"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

func newUser(t *testing.T, username, email string) *identity.User {
	t.Helper()
	u, err := identity.NewUser(username, email, "$argon2id$fake-hash", identity.UserMetadata{})
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	return u
}

func TestNewUserDefaultsToActive(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	if u.Status() != identity.UserStatusActive {
		t.Errorf("status = %d, want %d", u.Status(), identity.UserStatusActive)
	}
	if !u.IsActive() {
		t.Error("new user must be active")
	}
	if u.IsDeleted() {
		t.Error("new user must not be deleted")
	}
}

func TestUserStatusValidation(t *testing.T) {
	if !(identity.UserStatusActive).Valid() {
		t.Error("0 must be valid")
	}
	if !(identity.UserStatusDisabled).Valid() {
		t.Error("1 must be valid")
	}
	if !(identity.UserStatusDeleted).Valid() {
		t.Error("2 must be valid")
	}
	if (identity.UserStatus(3)).Valid() {
		t.Error("3 must be invalid")
	}

	// 缺陷 D-01 回归：注释里的 -1 / -2 不是合法取值
	for _, raw := range []int16{-1, -2, 3, 99} {
		if _, err := identity.ParseUserStatus(raw); err == nil {
			t.Errorf("status %d must be rejected", raw)
		}
	}
}

func TestSoftDeleteIsIrreversible(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	if err := u.SoftDelete(); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if !u.IsDeleted() {
		t.Fatal("user must be marked deleted")
	}

	// 已删除的账号不允许再改回正常状态
	if err := u.SetStatus(identity.UserStatusActive); err == nil {
		t.Fatal("deleted user must not be restorable")
	}
}

func TestUpdateMetadataMergesOnlyProvidedFields(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	nickname := "爱丽丝"
	phone := "13800000000"
	u.UpdateMetadata(identity.UserMetadata{Nickname: &nickname})
	u.UpdateMetadata(identity.UserMetadata{Phone: &phone})

	if u.Metadata().Nickname == nil || *u.Metadata().Nickname != nickname {
		t.Error("nickname from the first patch must survive")
	}
	if u.Metadata().Phone == nil || *u.Metadata().Phone != phone {
		t.Error("phone from the second patch must be applied")
	}
	if u.Metadata().Avatar != nil {
		t.Error("unset fields must stay nil")
	}
}

func TestUsernameAndEmailValidation(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	if err := u.Rename("ab"); err == nil {
		t.Error("username shorter than 3 chars must be rejected")
	}
	if err := u.ChangeEmail("not-an-email"); err == nil {
		t.Error("email without @ must be rejected")
	}
	if err := u.ChangeEmail(""); err == nil {
		t.Error("empty email must be rejected")
	}
	if err := u.ChangeEmail("alice@localhost"); err == nil {
		t.Error("email without a dot in the domain must be rejected")
	}
	if err := u.ChangeEmail("alice@example.com"); err != nil {
		t.Errorf("valid email must be accepted: %v", err)
	}

	if err := u.Rename("alice-updated"); err != nil {
		t.Errorf("valid rename must succeed: %v", err)
	}
	if u.Username() != "alice-updated" {
		t.Errorf("username = %q", u.Username())
	}
}

// TestLengthLimitsCountCharactersNotBytes 是契约一致性回归：
// Python 侧 Pydantic 的 min_length / max_length 按**字符**计数。
// Go 侧若用 len()（字节数），单个汉字（UTF-8 占 3 字节）会绕过最小长度校验，
// 造成 Go 放行而 Python 拒绝的契约漂移。
func TestLengthLimitsCountCharactersNotBytes(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	// 「编」是 1 个字符、3 个字节
	if err := u.Rename("编"); err == nil {
		t.Error("single CJK character must be rejected for a 3-character minimum")
	}
}

func TestRoleNameLengthCountsCharacters(t *testing.T) {
	// 「编辑」是 2 个字符、6 个字节，应当通过 min_length=2
	if _, err := identity.NewRole("editor", "编辑", ""); err != nil {
		t.Errorf("2 CJK characters must satisfy min_length=2: %v", err)
	}
	// 单个字符必须被拒绝
	if _, err := identity.NewRole("editor", "编", ""); err == nil {
		t.Error("single CJK character must be rejected for min_length=2")
	}
}

// TestPasswordHashIsNotLeaked 保证密码哈希不会出现在任何序列化输出里。
// 这是 T-1.1 的安全验收项：字段未导出，且没有 json tag 可泄漏。
func TestPasswordHashIsNotLeaked(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	// 只能通过显式 getter 读取，不存在 PasswordHash 字段可供 json/日志直接抓取
	if u.PasswordHash() != "$argon2id$fake-hash" {
		t.Error("getter must return the stored hash")
	}

	if err := u.ChangePassword(""); err == nil {
		t.Error("setting an empty hash must be rejected")
	}
	if err := u.ChangePassword("$argon2id$new-hash"); err != nil {
		t.Errorf("change password: %v", err)
	}
	if u.PasswordHash() != "$argon2id$new-hash" {
		t.Error("hash must be updated")
	}
}

func TestIsAdminRules(t *testing.T) {
	adminRole := identity.RestoreRole(identity.RoleSnapshot{ID: 1, Code: "admin", Name: "管理员"})
	superadminRole := identity.RestoreRole(identity.RoleSnapshot{ID: 2, Code: "SuperAdmin", Name: "超级管理员"})
	editorRole := identity.RestoreRole(identity.RoleSnapshot{ID: 3, Code: "editor", Name: "编辑"})
	systemRole := identity.RestoreRole(identity.RoleSnapshot{ID: 4, Code: "administrator", Name: "系统管理员"})

	t.Run("username admin is admin", func(t *testing.T) {
		u := newUser(t, "admin", "admin@example.com")
		if !u.IsAdmin() {
			t.Error(`user named "admin" must be admin`)
		}
	})

	t.Run("admin role grants admin", func(t *testing.T) {
		u := newUser(t, "bob", "bob@example.com")
		u.SetRoles([]*identity.Role{adminRole})
		if !u.IsAdmin() {
			t.Error("admin role must grant admin")
		}
	})

	t.Run("superadmin role is case insensitive", func(t *testing.T) {
		u := newUser(t, "carol", "carol@example.com")
		u.SetRoles([]*identity.Role{superadminRole})
		if !u.IsAdmin() {
			t.Error("SuperAdmin must be recognised regardless of case")
		}
	})

	t.Run("administrator role grants admin", func(t *testing.T) {
		u := newUser(t, "dave", "dave@example.com")
		u.SetRoles([]*identity.Role{systemRole})
		if !u.IsAdmin() {
			t.Error("administrator role must grant admin")
		}
	})

	t.Run("plain user is not admin", func(t *testing.T) {
		u := newUser(t, "eve", "eve@example.com")
		u.SetRoles([]*identity.Role{editorRole})
		if u.IsAdmin() {
			t.Error("editor role must not grant admin")
		}
	})

	t.Run("no roles is not admin", func(t *testing.T) {
		u := newUser(t, "frank", "frank@example.com")
		if u.IsAdmin() {
			t.Error("user without roles must not be admin")
		}
	})
}

func TestRecordLogin(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")
	if u.LastLoginAt() != nil {
		t.Fatal("last login must be nil initially")
	}

	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	u.RecordLogin(now, "203.0.113.7")

	if u.LastLoginAt() == nil || !u.LastLoginAt().Equal(now) {
		t.Error("last login time must be recorded")
	}
	if u.LastLoginIP() != "203.0.113.7" {
		t.Errorf("last login ip = %q", u.LastLoginIP())
	}
}

func TestRestoreUserFallsBackOnInvalidStatus(t *testing.T) {
	// 数据库有 CHECK 约束兜底，但应用层也要能容忍脏数据，
	// 避免一个非法状态值让整个请求 500。
	u := identity.RestoreUser(identity.UserSnapshot{
		ID:     7,
		Status: 99,
	})
	if u.Status() != identity.UserStatusDisabled {
		t.Errorf("invalid status must fall back to disabled, got %d", u.Status())
	}
	if u.ID() != 7 {
		t.Errorf("id = %d, want 7", u.ID())
	}
}

func TestDomainErrorsCarryExpectedKinds(t *testing.T) {
	cases := []struct {
		err        error
		wantKind   errors.Kind
		wantStatus int
	}{
		{identity.ErrInvalidCredentials, errors.KindUnauthenticated, 401},
		{identity.ErrUserInactive, errors.KindBadRequest, 400},
		{identity.ErrUserNotFound, errors.KindNotFound, 404},
		{identity.ErrUsernameTaken, errors.KindConflict, 409},
		{identity.ErrSystemRoleImmutable, errors.KindPermissionDenied, 403},
		{identity.ErrWeakPassword, errors.KindValidation, 422},
		{identity.ErrAdminRequired, errors.KindPermissionDenied, 403},
	}

	for _, c := range cases {
		if got := errors.KindOf(c.err); got != c.wantKind {
			t.Errorf("%v: kind = %v, want %v", c.err, got, c.wantKind)
		}
		if got := errors.KindOf(c.err).HTTPStatus(); got != c.wantStatus {
			t.Errorf("%v: status = %d, want %d", c.err, got, c.wantStatus)
		}
	}
}
