package authn

import "github.com/hermes-platform/go-service/internal/platform/errors"

// 令牌与服务令牌相关的错误。
//
// 这些错误放在 platform 而不是 domain/identity，理由：
// JWT 解析、Redis 吊销、登录锁定都是**基础设施的失败模式**，不是领域概念。
// 若定义在 domain，platform/authn 就必须反向 import domain，破坏依赖方向（ADR-0004）。

// ErrTokenInvalid JWT 无法解析、签名不通过或类型不匹配。
var ErrTokenInvalid = errors.New(errors.KindUnauthenticated, "无效的访问令牌")

// ErrTokenExpired JWT 已过期。与 ErrTokenInvalid 区分，
// 便于前端决定是刷新（401 expired）还是重新登录（401 invalid）。
var ErrTokenExpired = errors.New(errors.KindUnauthenticated, "访问令牌已过期")

// ErrTokenRevoked 令牌版本落后于服务端，即已被登出/改密/禁用操作吊销。
var ErrTokenRevoked = errors.New(errors.KindUnauthenticated, "令牌已被吊销，请重新登录")

// ErrRefreshTokenInvalid 刷新令牌不存在、已轮换或被吊销。
var ErrRefreshTokenInvalid = errors.New(errors.KindUnauthenticated, "刷新令牌无效或已过期")

// ErrLoginLocked 连续登录失败过多被临时锁定。
var ErrLoginLocked = errors.New(errors.KindBadRequest, "登录失败次数过多，请稍后再试")

// ErrServiceTokenInvalid 自动化上报接口的服务令牌缺失或错误。
var ErrServiceTokenInvalid = errors.New(errors.KindUnauthenticated, "Invalid or missing service token")

// ErrMissingCredentials 请求缺少 Authorization 头。
var ErrMissingCredentials = errors.New(errors.KindUnauthenticated, "缺少访问令牌")
