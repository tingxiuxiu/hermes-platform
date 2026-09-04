// Package http 是 HTTP 适配层的入口，承载**唯一路由表**。
//
// 本文件是契约一致性校验（tools/contract-diff）的基准：
// 路由的路径、方法、鉴权中间件必须与 Python 侧 router.py 一一对应。
// 修改本文件时请同步更新 docs/go-service/03-api-contract.md。
package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	automationhttp "github.com/hermes-platform/go-service/internal/adapter/http/automation"
	dashboardhttp "github.com/hermes-platform/go-service/internal/adapter/http/dashboard"
	identityhttp "github.com/hermes-platform/go-service/internal/adapter/http/identity"
	"github.com/hermes-platform/go-service/internal/adapter/http/middleware"
	systemhttp "github.com/hermes-platform/go-service/internal/adapter/http/system"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// RouterDeps 是构建路由表所需的全部依赖。
type RouterDeps struct {
	Cfg config.Config
	Log *slog.Logger

	// Authenticate 是已装配好的 JWT 鉴权中间件。
	Authenticate gin.HandlerFunc
	// ServiceToken 是上报接口的服务令牌鉴权中间件。
	ServiceToken gin.HandlerFunc
	// Telemetry 是 OpenTelemetry HTTP 中间件。nil 时用透传 no-op（测试/未配置）。
	Telemetry gin.HandlerFunc

	// Auth 处理认证接口，Identity 处理用户/角色/权限接口。
	Auth     *identityhttp.AuthHandlers
	Identity *identityhttp.Handlers
	// Automation 处理自动化上报与查询接口。
	Automation *automationhttp.Handlers
	// Dashboard 处理 dashboard 概览与运行中 case 查询接口。
	Dashboard *dashboardhttp.Handlers
	// System 处理系统字典接口。
	System *systemhttp.Handlers

	// Readiness 用于 /readyz 的依赖检查，返回 error 表示未就绪。
	Readiness func() error
}

// NewRouter 构建完整的路由表。
//
// 中间件顺序：RequestID → Recover → Logger → (Telemetry) → CORS → (AuthN) → handler
// Telemetry 在 Logger 之后，便于把请求 span 与日志关联；no-op 模式下是透传。
func NewRouter(deps RouterDeps) *gin.Engine {
	engine := gin.New()

	telemetry := deps.Telemetry
	if telemetry == nil {
		// 未提供时用透传（集成测试/本地降级）
		telemetry = func(c *gin.Context) { c.Next() }
	}

	engine.Use(
		middleware.RequestID(),
		middleware.Recover(deps.Log),
		middleware.Logger(deps.Log),
		telemetry,
		middleware.CORS(deps.Cfg.HTTP),
	)

	// 健康检查不加 API 前缀，供 K8s / Compose 探针直接访问
	registerHealthRoutes(engine, deps)

	api := engine.Group(deps.Cfg.APIPrefix)
	registerIdentityRoutes(api, deps)
	registerAutomationRoutes(api, deps)
	registerSystemRoutes(api, deps)

	return engine
}

// registerSystemRoutes 注册 system 上下文的字典接口（契约 §7，4 个接口）。
//
// 分两组：
//   - 查询：用户 JWT 鉴权（前端下拉框需要读字典）
//   - 写操作（增/改/删）：管理员权限
func registerSystemRoutes(api *gin.RouterGroup, deps RouterDeps) {
	system := api.Group("/system", deps.Authenticate)
	{
		// 查询（JWT）
		system.GET("/dicts", deps.System.ListDicts)

		// 写操作（管理员）
		system.POST("/dicts", middleware.RequireAdmin(), deps.System.CreateDict)
		system.PUT("/dicts/:id", middleware.RequireAdmin(), deps.System.UpdateDict)
		system.DELETE("/dicts/:id", middleware.RequireAdmin(), deps.System.DeleteDict)
	}
}

// registerAutomationRoutes 注册 automation 上下文的全部接口。
//
// 分两组（与 Python router.py 一致）：
//   - 上报类：ServiceToken 鉴权（pytest/CI Runner 调用，无用户态）
//   - 查询类：用户 JWT 鉴权（Dashboard 前端调用）
func registerAutomationRoutes(api *gin.RouterGroup, deps RouterDeps) {
	automation := api.Group("/automation")

	// 上报类（服务令牌）
	report := automation.Group("", deps.ServiceToken)
	{
		report.POST("/executions", deps.Automation.CreateExecution)
		report.PATCH("/executions/:build_uid", deps.Automation.UpdateExecution)
		report.POST("/executions/:build_uid/heartbeat", deps.Automation.Heartbeat)
		report.POST("/executions/:build_uid/items", deps.Automation.CreateItem)
		report.PATCH("/executions/:build_uid/items/:case_uid", deps.Automation.UpdateItem)
		report.POST("/executions/:build_uid/items/:item_id/steps", deps.Automation.AddSteps)
		report.POST("/items/:case_uid/steps", deps.Automation.UpsertStep)
	}

	// 查询类（用户 JWT）
	query := automation.Group("", deps.Authenticate)
	{
		query.GET("/pipelines", deps.Automation.ListPipelines)
		query.GET("/executions", deps.Automation.ListExecutions)
		query.GET("/executions/:build_uid/events", deps.Automation.StreamEvents)
		query.GET("/executions/:build_uid/live", deps.Automation.GetLive)
		query.GET("/executions/:build_uid/items/attempts", deps.Automation.GetCaseAttempts)
		query.GET("/executions/:build_uid/items", deps.Automation.ListExecutionItems)
		query.GET("/executions/:build_uid", deps.Automation.GetExecution)
		query.GET("/items/:case_uid", deps.Automation.GetItem)
		query.GET("/cases/history", deps.Automation.GetCaseHistory)

		// Dashboard 概览（也挂在 /automation 下）
		query.GET("/dashboard/overview", deps.Dashboard.GetOverview)
		query.GET("/dashboard/executions/:execution_id/cases", deps.Dashboard.GetRunningCases)
	}
}

// registerIdentityRoutes 注册 identity 上下文的全部接口。
//
// 分两组：
//   - public：无需认证（登录、注册、刷新）
//   - authed：需要有效 JWT；其中 requireAdmin 的额外要求管理员
func registerIdentityRoutes(api *gin.RouterGroup, deps RouterDeps) {
	public := api.Group("")
	{
		public.POST("/login/access-token", deps.Auth.LoginForm)
		public.POST("/login", deps.Auth.Login)
		public.POST("/register", deps.Auth.Register)
		public.POST("/login/refresh", deps.Auth.Refresh)
	}

	authed := api.Group("", deps.Authenticate)
	requireAdmin := middleware.RequireAdmin()

	// ---- 认证 ----
	authed.POST("/logout", deps.Auth.Logout)

	// ---- 用户管理（11 个接口）----
	users := authed.Group("/users")
	{
		users.GET("", requireAdmin, deps.Identity.Users.List)
		users.POST("", requireAdmin, deps.Identity.Users.Create)
		// 静态路径必须在 :user_id 之前注册才能被优先匹配
		users.GET("/role-options", deps.Identity.Users.RoleOptions)
		users.POST("/me/password", deps.Identity.Users.ChangeOwnPassword)
		users.POST("/batch-status", requireAdmin, deps.Identity.Users.BatchUpdateStatus)
		users.POST("/batch-delete", requireAdmin, deps.Identity.Users.BatchDelete)

		users.GET("/:user_id", deps.Identity.Users.Detail)
		users.PUT("/:user_id", requireAdmin, deps.Identity.Users.Update)
		users.PUT("/:user_id/roles", requireAdmin, deps.Identity.Users.AssignRoles)
		users.PUT("/:user_id/status", requireAdmin, deps.Identity.Users.UpdateStatus)
		users.POST("/:user_id/reset-password", requireAdmin, deps.Identity.Users.ResetPassword)
	}

	// ---- 角色管理（6 个接口，全部要求管理员）----
	roles := authed.Group("/roles", requireAdmin)
	{
		roles.GET("", deps.Identity.Roles.List)
		roles.POST("", deps.Identity.Roles.Create)
		roles.GET("/:role_id", deps.Identity.Roles.Detail)
		roles.PUT("/:role_id", deps.Identity.Roles.Update)
		roles.DELETE("/:role_id", deps.Identity.Roles.Delete)
		roles.PUT("/:role_id/permissions", deps.Identity.Roles.AssignPermissions)
	}

	// ---- 权限管理（4 个接口，全部要求管理员）----
	permissions := authed.Group("/permissions", requireAdmin)
	{
		// /tree 必须注册在 /:permission_id 之前
		permissions.GET("/tree", deps.Identity.Permissions.Tree)
		permissions.GET("", deps.Identity.Permissions.List)
		permissions.POST("", deps.Identity.Permissions.Create)
		permissions.DELETE("/:permission_id", deps.Identity.Permissions.Delete)
	}
}

// ensure response 包被引用：错误响应统一走 response.Error，
// 这里的空引用只是为了让 import 路径保持稳定。
var _ = response.OK
