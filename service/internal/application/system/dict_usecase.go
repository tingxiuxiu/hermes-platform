package system

import (
	"context"

	"github.com/hermes-platform/go-service/internal/domain/system"
)

// DictUseCase 编排字典项的查询与维护（T-7.2）。
type DictUseCase struct {
	dicts DictRepository
}

// NewDictUseCase 构造字典用例。
func NewDictUseCase(dicts DictRepository) *DictUseCase {
	return &DictUseCase{dicts: dicts}
}

// ListQuery 是字典查询条件。空串表示不筛选。
type ListQuery struct {
	DictType string
	Category string
	// OnlyActive 为 true 时只返回 active 状态的字典项。
	OnlyActive bool
}

// List 查询字典项。
func (uc *DictUseCase) List(ctx context.Context, q ListQuery) ([]*system.DictEntry, error) {
	return uc.dicts.List(ctx, q.DictType, q.Category, q.OnlyActive)
}

// Create 新增一条字典项，返回新记录。
func (uc *DictUseCase) Create(ctx context.Context, cmd system.NewDictEntryCommand) (*system.DictEntry, error) {
	entry, err := system.NewDictEntry(cmd)
	if err != nil {
		return nil, err
	}

	id, err := uc.dicts.Create(ctx, entry)
	if err != nil {
		return nil, err
	}
	return uc.dicts.GetByID(ctx, id)
}

// Update 更新一条字典项，返回更新后的记录。
func (uc *DictUseCase) Update(ctx context.Context, id int64, cmd system.UpdateDictEntryCommand) (*system.DictEntry, error) {
	entry, err := uc.dicts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := entry.Update(cmd); err != nil {
		return nil, err
	}
	if err := uc.dicts.Update(ctx, entry); err != nil {
		return nil, err
	}
	return uc.dicts.GetByID(ctx, id)
}

// Delete 删除一条字典项。
func (uc *DictUseCase) Delete(ctx context.Context, id int64) error {
	// 先确认存在（GetByID 会在不存在时返回 NotFound，避免静默删除）
	if _, err := uc.dicts.GetByID(ctx, id); err != nil {
		return err
	}
	return uc.dicts.Delete(ctx, id)
}
