// Package errors 提供跨层的错误分类机制。
//
// 设计要点：
//   - domain 层只声明错误，不感知 HTTP；
//   - adapter/http 层通过 KindOf 把错误翻译成 HTTP 状态码与响应 code；
//   - 所有领域错误都必须通过 New/Wrap 构造，保证 Kind 不丢失。
package errors

import (
	"errors"
	"fmt"
)

// Kind 是错误的高层分类。每个 Kind 唯一决定 HTTP 状态码与响应体中的 code。
type Kind int

const (
	KindUnknown Kind = iota
	KindInternal
	// KindBadRequest 对应 400。Python 侧对「账号被禁用」返回 400 Inactive user，
	// 为保持契约一致需要这一档（KindValidation 是 422，语义不同）。
	KindBadRequest
	KindValidation
	KindUnauthenticated
	KindPermissionDenied
	KindNotFound
	KindConflict
	KindUnavailable
)

var kindNames = map[Kind]string{
	KindUnknown:          "unknown",
	KindInternal:         "internal",
	KindBadRequest:       "bad_request",
	KindValidation:       "validation",
	KindUnauthenticated:  "unauthenticated",
	KindPermissionDenied: "permission_denied",
	KindNotFound:         "not_found",
	KindConflict:         "conflict",
	KindUnavailable:      "unavailable",
}

func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return kindNames[KindUnknown]
}

// HTTPStatus 返回该 Kind 对应的 HTTP 状态码。
// 映射关系见 docs/go-service/01-architecture.md §8.3。
func (k Kind) HTTPStatus() int {
	switch k {
	case KindBadRequest:
		return 400
	case KindValidation:
		return 422
	case KindUnauthenticated:
		return 401
	case KindPermissionDenied:
		return 403
	case KindNotFound:
		return 404
	case KindConflict:
		return 409
	case KindUnavailable:
		return 503
	default:
		return 500
	}
}

// Code 返回响应信封中的 code 字段，默认与 HTTP 状态码一致。
func (k Kind) Code() int {
	return k.HTTPStatus()
}

// Error 是贯穿所有层的错误类型。
type Error struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// New 构造一个带 Kind 的错误。
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

// Errorf 用格式化消息构造错误。
func Errorf(kind Kind, format string, args ...any) *Error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, args...)}
}

// Wrap 包装底层错误并保留其 Cause。Kind 由调用方指定（通常来自底层错误）。
func Wrap(kind Kind, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}

// Is 判断 err 是否属于指定 Kind。
func Is(err error, kind Kind) bool {
	if err == nil {
		return false
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == kind
	}
	return false
}

// KindOf 提取错误的 Kind，无法识别时返回 KindInternal。
// 未知错误落到 500 而不是 200，避免把内部故障包装成业务成功。
func KindOf(err error) Kind {
	if err == nil {
		return KindUnknown
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}

// Message 提取对外可展示的错误消息。
// 内部错误不应把底层细节暴露给调用方，统一返回兜底文案。
func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return "internal server error"
}
