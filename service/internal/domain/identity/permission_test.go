package identity_test

import (
	"testing"
	"time"

	"github.com/hermes-platform/go-service/internal/domain/identity"
)

func mustPermission(t *testing.T, id int64, code string, parentID *int64) *identity.Permission {
	t.Helper()
	return identity.RestorePermission(identity.PermissionSnapshot{
		ID:           id,
		ParentID:     parentID,
		Code:         code,
		Name:         "权限 " + code,
		ResourceType: int16(identity.ResourceTypeAPI),
		CreatedAt:    time.Now(),
	})
}

func TestResourceTypeValidation(t *testing.T) {
	for _, rt := range []identity.ResourceType{
		identity.ResourceTypeMenu,
		identity.ResourceTypeButton,
		identity.ResourceTypeAPI,
	} {
		if !rt.Valid() {
			t.Errorf("resource type %d must be valid", rt)
		}
	}
	for _, raw := range []int16{0, 4, -1} {
		if (identity.ResourceType(raw)).Valid() {
			t.Errorf("resource type %d must be invalid", raw)
		}
	}
}

func TestBuildPermissionTreeNesting(t *testing.T) {
	childID := int64(1)
	root := mustPermission(t, 1, "user:list", nil)
	child := mustPermission(t, 2, "user:create", &childID)

	tree := identity.BuildPermissionTree([]*identity.Permission{root, child})

	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if tree[0].Permission.Code() != "user:list" {
		t.Fatalf("root code = %q", tree[0].Permission.Code())
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Permission.Code() != "user:create" {
		t.Fatalf("child code = %q", tree[0].Children[0].Permission.Code())
	}
}

// TestOrphanPermissionBecomesRoot 是 Python get_permission_tree 的行为复刻：
// parent_id 指向不存在的节点时，该节点作为根节点返回，而不是被丢弃。
func TestOrphanPermissionBecomesRoot(t *testing.T) {
	missingParent := int64(999)
	orphan := mustPermission(t, 1, "user:orphan", &missingParent)
	normal := mustPermission(t, 2, "user:list", nil)

	tree := identity.BuildPermissionTree([]*identity.Permission{orphan, normal})

	if len(tree) != 2 {
		t.Fatalf("orphan must be kept as a root, expected 2 roots, got %d", len(tree))
	}
}

func TestSelfReferencedPermissionBecomesRoot(t *testing.T) {
	selfID := int64(1)
	self := mustPermission(t, 1, "user:self", &selfID)

	tree := identity.BuildPermissionTree([]*identity.Permission{self})

	if len(tree) != 1 {
		t.Fatalf("self-referencing node must become a root, got %d roots", len(tree))
	}
	if len(tree[0].Children) != 0 {
		t.Fatal("self-referencing node must have no children")
	}
}

func TestBuildPermissionTreeKeepsSiblingOrder(t *testing.T) {
	parentID := int64(1)
	parent := mustPermission(t, 1, "user", nil)
	a := mustPermission(t, 2, "user:list", &parentID)
	b := mustPermission(t, 3, "user:create", &parentID)
	c := mustPermission(t, 4, "user:delete", &parentID)

	tree := identity.BuildPermissionTree([]*identity.Permission{parent, a, b, c})

	children := tree[0].Children
	want := []string{"user:list", "user:create", "user:delete"}
	if len(children) != len(want) {
		t.Fatalf("children = %d, want %d", len(children), len(want))
	}
	for i, code := range want {
		if children[i].Permission.Code() != code {
			t.Errorf("child[%d] = %q, want %q", i, children[i].Permission.Code(), code)
		}
	}
}

func TestBuildPermissionTreeEmpty(t *testing.T) {
	tree := identity.BuildPermissionTree(nil)
	if len(tree) != 0 {
		t.Fatalf("empty input must produce empty tree, got %d roots", len(tree))
	}
}

// TestBuildPermissionTreeHandlesDeepNesting 锁住一个真实缺陷：
// 早期实现先把根节点 append 进结果切片（那一刻是副本），之后再给它挂子节点，
// 导致三层以上嵌套时孙节点丢失。必须在指针结构上建树、最后再转值。
func TestBuildPermissionTreeHandlesDeepNesting(t *testing.T) {
	rootID := int64(1)
	childID := int64(2)

	root := mustPermission(t, 1, "user", nil)
	child := mustPermission(t, 2, "user:list", &rootID)
	grandchild := mustPermission(t, 3, "user:list:export", &childID)

	tree := identity.BuildPermissionTree([]*identity.Permission{root, child, grandchild})

	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 child on root, got %d", len(tree[0].Children))
	}

	kids := tree[0].Children[0].Children
	if len(kids) != 1 {
		t.Fatalf("grandchild was lost: root.child has %d children, want 1", len(kids))
	}
	if kids[0].Permission.Code() != "user:list:export" {
		t.Fatalf("grandchild code = %q", kids[0].Permission.Code())
	}
}

func TestPermissionNameAndCodeValidation(t *testing.T) {
	if _, err := identity.NewPermission("a", "权限", identity.ResourceTypeAPI, "", ""); err == nil {
		t.Error("code shorter than 2 chars must be rejected")
	}
	if _, err := identity.NewPermission("user:list", "权", identity.ResourceTypeAPI, "", ""); err == nil {
		t.Error("name shorter than 2 chars must be rejected")
	}
	if _, err := identity.NewPermission("user:list", "用户列表", identity.ResourceType(9), "", ""); err == nil {
		t.Error("invalid resource type must be rejected")
	}
}
