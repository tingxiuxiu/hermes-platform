package identity_test

import (
	"testing"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

func TestNewRoleIsNeverSystem(t *testing.T) {
	r, err := identity.NewRole("editor", "编辑人员", "可以编辑内容")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}
	if r.IsSystem() {
		t.Error("roles created at runtime must never be system roles")
	}
	if !r.CanDelete() {
		t.Error("non-system roles must be deletable")
	}
}

func TestSystemRoleCannotBeDeleted(t *testing.T) {
	r := identity.RestoreRole(identity.RoleSnapshot{
		ID:       1,
		Code:     "admin",
		Name:     "管理员",
		IsSystem: true,
	})
	if r.CanDelete() {
		t.Error("system roles must not be deletable")
	}
}

func TestRoleFieldValidation(t *testing.T) {
	if _, err := identity.NewRole("e", "编辑人员", ""); err == nil {
		t.Error("code shorter than 2 chars must be rejected")
	}
	if _, err := identity.NewRole("editor", "编", ""); err == nil {
		t.Error("name shorter than 2 chars must be rejected")
	}

	r, err := identity.NewRole("editor", "编辑人员", "")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}
	if err := r.Rename("短"); err == nil {
		t.Error("rename to a too-short name must be rejected")
	}
	r.ChangeDescription("清空为空白")
	if r.Description() != "清空为空白" {
		t.Errorf("description = %q", r.Description())
	}
	r.ChangeDescription("")
	if r.Description() != "" {
		t.Errorf("empty description must clear the field, got %q", r.Description())
	}
}

func TestRolePermissions(t *testing.T) {
	r := identity.RestoreRole(identity.RoleSnapshot{ID: 1, Code: "editor", Name: "编辑"})
	if len(r.Permissions()) != 0 {
		t.Fatal("freshly restored role must have no permissions")
	}

	perm := mustPermission(t, 1, "user:list", nil)
	r.SetPermissions([]*identity.Permission{perm})

	if len(r.Permissions()) != 1 {
		t.Fatalf("permissions = %d, want 1", len(r.Permissions()))
	}
	if r.Permissions()[0].Code() != "user:list" {
		t.Errorf("permission code = %q", r.Permissions()[0].Code())
	}
}

func TestUserCanResolvePermissionThroughRoles(t *testing.T) {
	u := newUser(t, "alice", "alice@example.com")

	perm := mustPermission(t, 1, "user:list", nil)
	role := identity.RestoreRole(identity.RoleSnapshot{ID: 1, Code: "editor", Name: "编辑"})
	role.SetPermissions([]*identity.Permission{perm})
	u.SetRoles([]*identity.Role{role})

	if !u.HasPermissionCode("user:list") {
		t.Error("user must inherit the permission from its role")
	}
	if u.HasPermissionCode("user:delete") {
		t.Error("user must not have permissions it was not granted")
	}
}
