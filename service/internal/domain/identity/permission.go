package identity

import (
	"strings"
	"time"
)

// ResourceType 权限资源类型，对应 permissions.resource_type。
type ResourceType int16

const (
	ResourceTypeMenu   ResourceType = 1 // 菜单
	ResourceTypeButton ResourceType = 2 // 按钮 / 功能
	ResourceTypeAPI    ResourceType = 3 // API 接口
)

// Valid 判断资源类型是否合法（与 CHECK 约束 chk_permissions_type 一致）。
func (t ResourceType) Valid() bool {
	switch t {
	case ResourceTypeMenu, ResourceTypeButton, ResourceTypeAPI:
		return true
	default:
		return false
	}
}

// Permission 是权限实体，支持通过 parent_id 构成树形结构。
type Permission struct {
	id           int64
	parentID     *int64
	code         string
	name         string
	resourceType ResourceType
	path         string
	method       string
	createdAt    time.Time
}

// PermissionSnapshot 用于从数据库重建权限。
type PermissionSnapshot struct {
	ID           int64
	ParentID     *int64
	Code         string
	Name         string
	ResourceType int16
	Path         string
	Method       string
	CreatedAt    time.Time
}

// NewPermission 构造权限节点。
func NewPermission(code, name string, resourceType ResourceType, path, method string) (*Permission, error) {
	p := &Permission{}
	if err := p.setCode(code); err != nil {
		return nil, err
	}
	if err := p.setName(name); err != nil {
		return nil, err
	}
	if !resourceType.Valid() {
		return nil, ErrInvalidResourceType
	}
	p.resourceType = resourceType
	p.path = path
	p.method = method
	return p, nil
}

func RestorePermission(s PermissionSnapshot) *Permission {
	rt := ResourceType(s.ResourceType)
	if !rt.Valid() {
		rt = ResourceTypeMenu
	}
	return &Permission{
		id:           s.ID,
		parentID:     s.ParentID,
		code:         s.Code,
		name:         s.Name,
		resourceType: rt,
		path:         s.Path,
		method:       s.Method,
		createdAt:    s.CreatedAt,
	}
}

func (p *Permission) ID() int64                  { return p.id }
func (p *Permission) ParentID() *int64           { return p.parentID }
func (p *Permission) Code() string               { return p.code }
func (p *Permission) Name() string               { return p.name }
func (p *Permission) ResourceType() ResourceType { return p.resourceType }
func (p *Permission) Path() string               { return p.path }
func (p *Permission) Method() string             { return p.method }
func (p *Permission) CreatedAt() time.Time       { return p.createdAt }

// SetParent 挂到父节点下。传 nil 表示成为根节点。
func (p *Permission) SetParent(parentID *int64) { p.parentID = parentID }

func (p *Permission) setCode(code string) error {
	code = strings.TrimSpace(code)
	if textLen(code) < 2 || textLen(code) > 128 {
		return ErrInvalidPermissionCode
	}
	p.code = code
	return nil
}

func (p *Permission) setName(name string) error {
	name = strings.TrimSpace(name)
	if textLen(name) < 2 || textLen(name) > 64 {
		return ErrInvalidPermissionName
	}
	p.name = name
	return nil
}

// PermissionNode 是权限树的节点，供 GET /permissions/tree 返回。
// 树结构属于读模型，不污染 Permission 实体本身。
type PermissionNode struct {
	Permission *Permission
	Children   []PermissionNode
}

// BuildPermissionTree 把平铺的权限列表组装成树。
//
// 接收指针切片以匹配仓储的返回形态，避免在用例层做一次无意义的拷贝。
//
// 关键行为（对齐 Python get_permission_tree）：
//   - parent_id 指向不存在的节点时，该节点**作为根节点**处理（孤儿节点不上浮丢失）；
//   - 自引用的节点同样作为根节点（否则会形成环）；
//   - 同层节点保持输入顺序（Python 侧按 id ASC 查询，调用方保证）。
//
// 实现要点：先用指针节点建树，最后再一次性转成值切片。
// 直接对值切片做「先 append 到 roots、后给子节点追加 Children」会在
// 三层以上嵌套时丢子树——因为 append 进 roots 的是那一刻的副本。
func BuildPermissionTree(perms []*Permission) []PermissionNode {
	nodes := make([]*buildNode, len(perms))
	byID := make(map[int64]*buildNode, len(perms))
	for i, p := range perms {
		n := &buildNode{perm: p}
		nodes[i] = n
		byID[p.ID()] = n
	}

	isChild := make([]bool, len(perms))
	for i, p := range perms {
		parentID := p.ParentID()
		if parentID == nil {
			continue
		}
		parent, ok := byID[*parentID]
		if !ok || parent == nodes[i] {
			// 父节点缺失或自引用：作为根节点，避免节点丢失或形成环
			continue
		}
		parent.children = append(parent.children, nodes[i])
		isChild[i] = true
	}

	roots := make([]PermissionNode, 0, len(perms))
	for i, n := range nodes {
		if !isChild[i] {
			roots = append(roots, n.toNode())
		}
	}
	return roots
}

// buildNode 是建树过程中的中间结构，用指针保证任意深度的子树都能正确挂载。
type buildNode struct {
	perm     *Permission
	children []*buildNode
}

func (n *buildNode) toNode() PermissionNode {
	out := PermissionNode{
		Permission: n.perm,
		Children:   make([]PermissionNode, 0, len(n.children)),
	}
	for _, c := range n.children {
		out.Children = append(out.Children, c.toNode())
	}
	return out
}
