// Package system 承载 system 上下文的应用层：系统字典的查询与维护。
package system

import (
	"context"

	"github.com/hermes-platform/go-service/internal/domain/system"
)

// DictRepository 是字典项的持久化端口。
type DictRepository interface {
	// Create 写入一条字典项。违反复合唯一约束时返回 ErrDictDuplicate。
	Create(ctx context.Context, d *system.DictEntry) (int64, error)

	// GetByID 按主键读取。
	GetByID(ctx context.Context, id int64) (*system.DictEntry, error)

	// List 按条件查询，按 sort_order 升序、id 升序。
	// dictType / category 为空串表示不筛选。
	List(ctx context.Context, dictType, category string, onlyActive bool) ([]*system.DictEntry, error)

	// Update 更新一条字典项。
	Update(ctx context.Context, d *system.DictEntry) error

	// Delete 删除一条字典项。
	Delete(ctx context.Context, id int64) error
}
