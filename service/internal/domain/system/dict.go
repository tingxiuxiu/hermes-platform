// Package system 承载 system 上下文的领域模型（系统字典）。
package system

import (
	"strings"
	"time"

	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// DictStatus 是字典项的状态。
type DictStatus string

const (
	DictStatusActive   DictStatus = "active"
	DictStatusInactive DictStatus = "inactive"
)

// Valid 判断状态值是否合法。
func (s DictStatus) Valid() bool {
	switch s {
	case DictStatusActive, DictStatusInactive:
		return true
	default:
		return false
	}
}

// ParseDictStatus 解析字典状态值。
func ParseDictStatus(raw string) (DictStatus, error) {
	s := DictStatus(strings.TrimSpace(raw))
	if !s.Valid() {
		return "", errors.Errorf(errors.KindValidation,
			"非法的字典状态: %q（允许 active/inactive）", raw)
	}
	return s, nil
}

// DictEntry 是一条系统字典项。
//
// 复合唯一（dict_type, code）对应 DB 约束 `uq_sys_dict_type_code`。
//
// ⚠️ 注意：Python 的 alembic 迁移里多了一个**单列** UNIQUE(dict_type)，
// 导致同一个字典类型只能存一行，与复合唯一的语义矛盾（缺陷，见设计文档 §2.4）。
// Go 的初始迁移**不生成**该约束，并在修复迁移中 DROP —— 领域层按正确的
// 复合唯一语义建模，不受错误约束影响。
type DictEntry struct {
	id          int64
	dictType    string
	code        int
	label       string
	dictName    string
	description string
	category    string
	sortOrder   int
	color       string
	status      DictStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// DictSnapshot 用于从数据库重建字典项。
type DictSnapshot struct {
	ID          int64
	DictType    string
	Code        int
	Label       string
	DictName    string
	Description string
	Category    string
	SortOrder   int
	Color       string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDictEntryCommand 是创建字典项的输入。
type NewDictEntryCommand struct {
	DictType    string
	Code        int
	Label       string
	DictName    string
	Description string
	Category    string
	SortOrder   int
	Color       string
	Status      string
}

// NewDictEntry 创建一条字典项。
func NewDictEntry(cmd NewDictEntryCommand) (*DictEntry, error) {
	d := &DictEntry{}
	if err := d.setDictType(cmd.DictType); err != nil {
		return nil, err
	}
	if err := d.setLabel(cmd.Label); err != nil {
		return nil, err
	}
	if err := d.setDictName(cmd.DictName); err != nil {
		return nil, err
	}
	d.code = cmd.Code
	d.description = strings.TrimSpace(cmd.Description)
	d.category = strings.TrimSpace(cmd.Category)
	d.sortOrder = cmd.SortOrder
	d.color = strings.TrimSpace(cmd.Color)

	status := DictStatusActive
	if cmd.Status != "" {
		parsed, err := ParseDictStatus(cmd.Status)
		if err != nil {
			return nil, err
		}
		status = parsed
	}
	d.status = status
	return d, nil
}

// RestoreDictEntry 从数据库重建字典项。
func RestoreDictEntry(s DictSnapshot) *DictEntry {
	status := DictStatus(s.Status)
	if !status.Valid() {
		status = DictStatusActive
	}
	return &DictEntry{
		id:          s.ID,
		dictType:    s.DictType,
		code:        s.Code,
		label:       s.Label,
		dictName:    s.DictName,
		description: s.Description,
		category:    s.Category,
		sortOrder:   s.SortOrder,
		color:       s.Color,
		status:      status,
		createdAt:   s.CreatedAt,
		updatedAt:   s.UpdatedAt,
	}
}

// ---- 读取器 ----

func (d *DictEntry) ID() int64            { return d.id }
func (d *DictEntry) DictType() string     { return d.dictType }
func (d *DictEntry) Code() int            { return d.code }
func (d *DictEntry) Label() string        { return d.label }
func (d *DictEntry) DictName() string     { return d.dictName }
func (d *DictEntry) Description() string  { return d.description }
func (d *DictEntry) Category() string     { return d.category }
func (d *DictEntry) SortOrder() int       { return d.sortOrder }
func (d *DictEntry) Color() string        { return d.color }
func (d *DictEntry) Status() DictStatus   { return d.status }
func (d *DictEntry) CreatedAt() time.Time { return d.createdAt }
func (d *DictEntry) UpdatedAt() time.Time { return d.updatedAt }

// ---- 行为 ----

// UpdateField 应用一次局部更新。nil 表示「不修改该字段」。
//
// 用独立的可空更新结构而不是直接改字段，保证校验在实体内完成，
// 调用方无法绕过规则写入非法状态。
func (d *DictEntry) Update(cmd UpdateDictEntryCommand) error {
	if cmd.Label != nil {
		if err := d.setLabel(*cmd.Label); err != nil {
			return err
		}
	}
	if cmd.DictName != nil {
		if err := d.setDictName(*cmd.DictName); err != nil {
			return err
		}
	}
	if cmd.Description != nil {
		d.description = strings.TrimSpace(*cmd.Description)
	}
	if cmd.Category != nil {
		d.category = strings.TrimSpace(*cmd.Category)
	}
	if cmd.SortOrder != nil {
		d.sortOrder = *cmd.SortOrder
	}
	if cmd.Color != nil {
		d.color = strings.TrimSpace(*cmd.Color)
	}
	if cmd.Status != nil {
		parsed, err := ParseDictStatus(*cmd.Status)
		if err != nil {
			return err
		}
		d.status = parsed
	}
	return nil
}

// UpdateDictEntryCommand 是更新字典项的输入。
// 指针字段为 nil 表示不修改该字段（区别于零值）。
type UpdateDictEntryCommand struct {
	Label       *string
	DictName    *string
	Description *string
	Category    *string
	SortOrder   *int
	Color       *string
	Status      *string
}

func (d *DictEntry) setDictType(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrDictTypeRequired
	}
	if len(v) > 64 {
		return ErrDictTypeTooLong
	}
	d.dictType = v
	return nil
}

func (d *DictEntry) setLabel(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrDictLabelRequired
	}
	if len(v) > 128 {
		return ErrDictLabelTooLong
	}
	d.label = v
	return nil
}

func (d *DictEntry) setDictName(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrDictNameRequired
	}
	if len(v) > 128 {
		return ErrDictNameTooLong
	}
	d.dictName = v
	return nil
}
