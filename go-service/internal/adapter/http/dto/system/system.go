// Package system 定义 system 上下文的 HTTP 请求/响应 DTO。
//
// 字段严格取自 docs/03-api-contract.md §7 的 Go 侧新增最小集
// （Python 只有模型无路由，契约以设计文档为准）。
package system

import (
	"github.com/hermes-platform/go-service/internal/domain/system"
)

// DictItem 是字典项的响应结构。
type DictItem struct {
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

// DictListQuery 是 GET /system/dicts 的查询参数。
type DictListQuery struct {
	DictType string `form:"dict_type"`
	Category string `form:"category"`
	// OnlyActive 为 true 时只返回 active 项（query: only_active=true）
	OnlyActive bool `form:"only_active"`
}

// CreateDictRequest 是 POST /system/dicts 的请求体。
//
// ⚠️ code 用 **指针**：go-playground/validator 的 `required` 把零值当作「未提供」，
// 若用 `int` + `required` 会导致 code=0（合法值，如「运行中」的状态码）被拒 422。
// 这是本项目第二次踩这个坑（首次是 identity 的 status 字段，见变更日志 BUG-4）——
// 凡是「零值是合法业务值」的必填数值字段，一律用指针。
type CreateDictRequest struct {
	DictType    string `json:"dict_type" binding:"required,max=64"`
	Code        *int   `json:"code" binding:"required"`
	Label       string `json:"label" binding:"required,max=128"`
	DictName    string `json:"dict_name" binding:"required,max=128"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"max=64"`
	SortOrder   int    `json:"sort_order"`
	Color       string `json:"color" binding:"max=32"`
	Status      string `json:"status"`
}

// UpdateDictRequest 是 PUT /system/dicts/{id} 的请求体。
// 指针字段为 nil 表示不修改（区别于零值）。
type UpdateDictRequest struct {
	Label       *string `json:"label" binding:"omitempty,max=128"`
	DictName    *string `json:"dict_name" binding:"omitempty,max=128"`
	Description *string `json:"description"`
	Category    *string `json:"category" binding:"omitempty,max=64"`
	SortOrder   *int    `json:"sort_order"`
	Color       *string `json:"color" binding:"omitempty,max=32"`
	Status      *string `json:"status"`
}

// ToUpdateCommand 把请求 DTO 转成领域更新命令。
func (r UpdateDictRequest) ToUpdateCommand() system.UpdateDictEntryCommand {
	return system.UpdateDictEntryCommand{
		Label:       r.Label,
		DictName:    r.DictName,
		Description: r.Description,
		Category:    r.Category,
		SortOrder:   r.SortOrder,
		Color:       r.Color,
		Status:      r.Status,
	}
}

// ToCreateCommand 把请求 DTO 转成领域创建命令。
// Code 已经过 binding:"required" 校验，这里解引用是安全的。
func (r CreateDictRequest) ToCreateCommand() system.NewDictEntryCommand {
	code := 0
	if r.Code != nil {
		code = *r.Code
	}
	return system.NewDictEntryCommand{
		DictType:    r.DictType,
		Code:        code,
		Label:       r.Label,
		DictName:    r.DictName,
		Description: r.Description,
		Category:    r.Category,
		SortOrder:   r.SortOrder,
		Color:       r.Color,
		Status:      r.Status,
	}
}

// ToDictItem 把领域字典项转成响应 DTO。
// 空串的描述/分类/颜色以 null 呈现（与契约示例一致）。
func ToDictItem(d *system.DictEntry) DictItem {
	return DictItem{
		ID:          d.ID(),
		DictType:    d.DictType(),
		Code:        d.Code(),
		Label:       d.Label(),
		DictName:    d.DictName(),
		Description: nullIfEmpty(d.Description()),
		Category:    nullIfEmpty(d.Category()),
		SortOrder:   d.SortOrder(),
		Color:       nullIfEmpty(d.Color()),
		Status:      string(d.Status()),
	}
}

// ToDictItemList 批量转换。
func ToDictItemList(items []*system.DictEntry) []DictItem {
	out := make([]DictItem, 0, len(items))
	for _, d := range items {
		out = append(out, ToDictItem(d))
	}
	return out
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}
