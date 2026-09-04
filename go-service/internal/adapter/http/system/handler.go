// Package system 承载 system 上下文的 HTTP 处理器（系统字典）。
package system

import (
	"strconv"

	"github.com/gin-gonic/gin"

	dtosystem "github.com/hermes-platform/go-service/internal/adapter/http/dto/system"
	appsystem "github.com/hermes-platform/go-service/internal/application/system"
	"github.com/hermes-platform/go-service/internal/domain/system"
	"github.com/hermes-platform/go-service/internal/platform/errors"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// Handlers 聚合 system 上下文的处理器（4 个字典接口）。
type Handlers struct {
	dicts *appsystem.DictUseCase
}

// NewHandlers 构造 system 处理器。
func NewHandlers(dicts *appsystem.DictUseCase) *Handlers {
	return &Handlers{dicts: dicts}
}

// ListDicts 处理 GET /system/dicts（JWT 可访问）。
func (h *Handlers) ListDicts(c *gin.Context) {
	var q dtosystem.DictListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, 422, "查询参数不合法")
		return
	}

	items, err := h.dicts.List(c.Request.Context(), appsystem.ListQuery{
		DictType:   q.DictType,
		Category:   q.Category,
		OnlyActive: q.OnlyActive,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "获取字典列表成功", dtosystem.ToDictItemList(items))
}

// CreateDict 处理 POST /system/dicts（管理员）。
func (h *Handlers) CreateDict(c *gin.Context) {
	var req dtosystem.CreateDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	// code 范围校验：DB 列是 SMALLINT（req.Code 已经过 required 校验，非 nil）
	if *req.Code < -32768 || *req.Code > 32767 {
		response.Error(c, errors.Errorf(errors.KindValidation,
			"code 必须在 -32768 ~ 32767 范围内（数据库列为 SMALLINT）"))
		return
	}

	entry, err := h.dicts.Create(c.Request.Context(), req.ToCreateCommand())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "创建字典项成功", dtosystem.ToDictItem(entry))
}

// UpdateDict 处理 PUT /system/dicts/{id}（管理员）。
func (h *Handlers) UpdateDict(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, 422, "路径参数不合法")
		return
	}

	var req dtosystem.UpdateDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 422, "请求参数不合法")
		return
	}

	entry, err := h.dicts.Update(c.Request.Context(), id, req.ToUpdateCommand())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "更新字典项成功", dtosystem.ToDictItem(entry))
}

// DeleteDict 处理 DELETE /system/dicts/{id}（管理员）。
func (h *Handlers) DeleteDict(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, 422, "路径参数不合法")
		return
	}

	if err := h.dicts.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "删除字典项成功", nil)
}

// ensure 编译期引用领域包（保持依赖方向可追踪）。
var _ = system.DictStatusActive
