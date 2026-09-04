package identity

import (
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// 本文件集中声明 identity 上下文的领域错误（任务 T-1.3）。
//
// 约定：
//   - 错误只描述「发生了什么」，不描述 HTTP 状态码；
//   - 状态码由 platform/errors.Kind 决定，adapter/http 层负责翻译；
//   - 消息文案与 Python 版保持一致，避免前端需要同步改造。

// ---- 认证 ----

// ErrInvalidCredentials 用于登录失败。
// 用户不存在与密码错误**共用**这一个错误，避免通过响应差异枚举账号。
var ErrInvalidCredentials = errors.New(errors.KindUnauthenticated, "用户名或密码错误")

// ErrUserInactive 账号被禁用。对齐 Python 的 400 Inactive user。
var ErrUserInactive = errors.New(errors.KindBadRequest, "账号已被禁用")

// 令牌相关错误（ErrTokenInvalid / ErrTokenExpired / ErrRefreshTokenInvalid /
// ErrTokenRevoked / ErrLoginLocked）刻意**不定义在这里**，而放在 platform/authn：
// 它们是 JWT 与 Redis 这类基础设施的失败模式，不属于领域概念。
// 让 platform 反向依赖 domain 会破坏分层依赖方向（ADR-0004）。

// ---- 用户 ----

// ErrUserNotFound 用户不存在或已被软删除。
var ErrUserNotFound = errors.New(errors.KindNotFound, "用户不存在")

// ErrUsernameTaken 用户名冲突。
var ErrUsernameTaken = errors.New(errors.KindConflict, "用户名已被占用")

// ErrEmailTaken 邮箱冲突。
var ErrEmailTaken = errors.New(errors.KindConflict, "邮箱已被占用")

// ErrUserConflict 用户名或邮箱冲突（无法区分时使用，对齐 Python 注册接口的文案）。
var ErrUserConflict = errors.New(errors.KindConflict, "用户名或邮箱已被注册")

// ErrWeakPassword 密码不符合强度要求。
var ErrWeakPassword = errors.New(errors.KindValidation, "密码长度至少 6 位")

// ErrInvalidUsername 用户名不合法。
var ErrInvalidUsername = errors.New(errors.KindValidation, "用户名长度需在 3 到 64 个字符之间")

// ErrInvalidEmail 邮箱格式不合法。
var ErrInvalidEmail = errors.New(errors.KindValidation, "邮箱格式不正确")

// ErrInvalidStatusForEnableDisable 启用/禁用接口不接受「已删除」状态。
// 软删除只能通过批量删除接口进行，避免两个接口语义重叠。
var ErrInvalidStatusForEnableDisable = errors.New(errors.KindValidation, "状态只能是 0(启用) 或 1(禁用)")

// ErrOldPasswordMismatch 修改密码时原密码错误。
var ErrOldPasswordMismatch = errors.New(errors.KindUnauthenticated, "原密码输入错误")

// ---- 角色 ----

// ErrRoleNotFound 角色不存在。
var ErrRoleNotFound = errors.New(errors.KindNotFound, "角色不存在")

// ErrRoleCodeTaken 角色标识冲突。
var ErrRoleCodeTaken = errors.New(errors.KindConflict, "角色标识已存在")

// ErrSystemRoleImmutable 内置系统角色不可删除。
var ErrSystemRoleImmutable = errors.New(errors.KindPermissionDenied, "系统内置角色不可删除")

// ErrInvalidRoleCode 角色标识格式不合法。
var ErrInvalidRoleCode = errors.New(errors.KindValidation, "角色标识长度需在 2 到 64 个字符之间")

// ErrInvalidRoleName 角色名称格式不合法。
var ErrInvalidRoleName = errors.New(errors.KindValidation, "角色名称长度需在 2 到 64 个字符之间")

// ErrInvalidPermissionCode 权限标识格式不合法。
var ErrInvalidPermissionCode = errors.New(errors.KindValidation, "权限标识长度需在 2 到 128 个字符之间")

// ErrInvalidPermissionName 权限名称格式不合法。
var ErrInvalidPermissionName = errors.New(errors.KindValidation, "权限名称长度需在 2 到 64 个字符之间")

// ErrInvalidResourceType 权限资源类型不合法。
var ErrInvalidResourceType = errors.New(errors.KindValidation, "资源类型必须是 1(菜单)、2(按钮) 或 3(API)")

// ---- 权限 ----

// ErrPermissionNotFound 权限不存在。
var ErrPermissionNotFound = errors.New(errors.KindNotFound, "权限不存在")

// ErrPermissionCodeTaken 权限标识冲突。
var ErrPermissionCodeTaken = errors.New(errors.KindConflict, "权限标识已存在")

// ErrForbidden 通用越权错误。
var ErrForbidden = errors.New(errors.KindPermissionDenied, "权限不足")

// ErrAdminRequired 需要管理员权限。
var ErrAdminRequired = errors.New(errors.KindPermissionDenied, "权限不足：只有管理员才允许执行此操作")

// ErrMissingPermission 缺少指定权限码。message 需要带上具体权限码，用 Errorf 构造。
func ErrMissingPermission(code string) *errors.Error {
	return errors.Errorf(errors.KindPermissionDenied, "权限不足，缺失权限: [%s]", code)
}

// ErrRoleNotFoundByID 带 ID 的角色不存在，message 与 Python 的 f-string 文案对齐。
func ErrRoleNotFoundByID(id int64) *errors.Error {
	return errors.Errorf(errors.KindNotFound, "指定角色 ID %d 不存在", id)
}

// ErrPermissionNotFoundByID 带 ID 的权限不存在。
func ErrPermissionNotFoundByID(id int64) *errors.Error {
	return errors.Errorf(errors.KindNotFound, "指定权限 ID %d 不存在", id)
}

// ErrRoleCodeTakenWithCode 带 code 的角色冲突。
func ErrRoleCodeTakenWithCode(code string) *errors.Error {
	return errors.Errorf(errors.KindConflict, "角色标识 code '%s' 已存在", code)
}

// ErrPermissionCodeTakenWithCode 带 code 的权限冲突。
func ErrPermissionCodeTakenWithCode(code string) *errors.Error {
	return errors.Errorf(errors.KindConflict, "权限标识 code '%s' 已存在", code)
}

// ErrSystemRoleImmutableWithCode 带 code 的系统角色不可删除。
func ErrSystemRoleImmutableWithCode(code string) *errors.Error {
	return errors.Errorf(errors.KindPermissionDenied, "系统内置角色 '%s' 不可删除", code)
}
