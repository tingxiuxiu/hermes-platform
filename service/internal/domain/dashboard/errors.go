package dashboard

import "github.com/hermes-platform/go-service/internal/platform/errors"

// ErrExecutionNotFound 源表中不存在该 execution。
var ErrExecutionNotFound = errors.New(errors.KindNotFound, "执行记录不存在")
