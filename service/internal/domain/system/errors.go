package system

import "github.com/hermes-platform/go-service/internal/platform/errors"

// 本文件集中声明 system 上下文的领域错误。

// ErrDictNotFound 字典项不存在。
var ErrDictNotFound = errors.New(errors.KindNotFound, "字典项不存在")

// ErrDictTypeRequired 字典类型必填。
var ErrDictTypeRequired = errors.New(errors.KindValidation, "dict_type 不能为空")

// ErrDictTypeTooLong 字典类型超长（DB 列 VARCHAR(64)）。
var ErrDictTypeTooLong = errors.New(errors.KindValidation, "dict_type 长度不能超过 64 个字符")

// ErrDictLabelRequired 标签必填。
var ErrDictLabelRequired = errors.New(errors.KindValidation, "label 不能为空")

// ErrDictLabelTooLong 标签超长（DB 列 VARCHAR(128)）。
var ErrDictLabelTooLong = errors.New(errors.KindValidation, "label 长度不能超过 128 个字符")

// ErrDictNameRequired 字典名称必填。
var ErrDictNameRequired = errors.New(errors.KindValidation, "dict_name 不能为空")

// ErrDictNameTooLong 字典名称超长（DB 列 VARCHAR(128)）。
var ErrDictNameTooLong = errors.New(errors.KindValidation, "dict_name 长度不能超过 128 个字符")

// ErrDictDuplicate 同一 dict_type 下 code 已存在（违反复合唯一约束）。
var ErrDictDuplicate = errors.New(errors.KindConflict,
	"同一字典类型下该 code 已存在")
