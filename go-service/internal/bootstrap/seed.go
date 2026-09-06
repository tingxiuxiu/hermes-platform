// Package bootstrap 是组合根（Composition Root）：
// 唯一允许 import 所有层的地方，负责装配依赖（ADR-0004）。
//
// 本文件承载幂等种子数据。种子刻意**不放在数据库迁移里**，原因：
//   - 集成测试在用例之间会 TRUNCATE 全部表，迁移里的种子会被一并清掉，
//     导致后续用例拿不到 admin 角色；
//   - 放在 Go 代码里，测试与生产启动共用同一份种子逻辑，不会漂移。
package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	persistence "github.com/hermes-platform/go-service/internal/adapter/persistence/identity"
	appidentity "github.com/hermes-platform/go-service/internal/application/identity"
	"github.com/hermes-platform/go-service/internal/domain/identity"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

// Seeder 是种子数据写入端口，由 adapter/persistence/identity 实现。
// bootstrap 只依赖接口，不依赖具体实现。
type Seeder interface {
	// UpsertRoles 幂等写入角色，返回 code -> id 映射。
	UpsertRoles(ctx context.Context, roles []identity.RoleSeed) (map[string]int64, error)
	// UpsertPermissions 幂等写入权限，返回 code -> id 映射。
	UpsertPermissions(ctx context.Context, perms []identity.PermissionSeed) (map[string]int64, error)
	// SyncRolePermissions 把角色的权限集合对齐到给定权限码列表（覆盖式）。
	SyncRolePermissions(ctx context.Context, roleCode string, permissionCodes []string) error
	// AllPermissionCodes 返回库中全部权限码，用于给 admin 授予全量权限。
	AllPermissionCodes(ctx context.Context) ([]string, error)
}

// SeedRolesAndPermissions 幂等地写入预置角色、权限与绑定关系。
// 可安全重复执行：集成测试在每次 TRUNCATE 之后调用它重建基线。
func SeedRolesAndPermissions(ctx context.Context, s Seeder) error {
	if _, err := s.UpsertRoles(ctx, identity.SystemRoles); err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}
	if _, err := s.UpsertPermissions(ctx, identity.SystemPermissions); err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}

	all, err := s.AllPermissionCodes(ctx)
	if err != nil {
		return fmt.Errorf("seed: list permission codes: %w", err)
	}
	if err := s.SyncRolePermissions(ctx, "admin", all); err != nil {
		return fmt.Errorf("seed: grant all to admin: %w", err)
	}
	if err := s.SyncRolePermissions(ctx, "viewer", identity.ViewerPermissions); err != nil {
		return fmt.Errorf("seed: grant read-only to viewer: %w", err)
	}
	return nil
}

// SeedAll 是启动期的总入口：角色/权限基线 + 首启超级管理员。
//
// 幂等：可安全重复执行。由 cmd/api 在服务启动时调用。
func SeedAll(ctx context.Context, app *App, cfg config.Config) error {
	// 1. 角色/权限基线
	seeder := persistence.NewSeedRepo(app.DB)
	if err := SeedRolesAndPermissions(ctx, seeder); err != nil {
		return err
	}

	// 2. 首启超级管理员（若配置了 FIRST_SUPERUSER）
	if err := seedFirstSuperuser(ctx, app, cfg); err != nil {
		return err
	}
	return nil
}

// seedFirstSuperuser 幂等创建超级管理员并绑定 admin 角色。
//
// FIRST_SUPERUSER 与 Python 一致，可以是邮箱（EmailStr，如 admin@example.com）
// 或纯用户名。若是邮箱：email 用原值，username 用 @ 前的本地部分。
//
// 幂等语义：用户已存在（含软删除）时不做任何操作，直接返回。
func seedFirstSuperuser(ctx context.Context, app *App, cfg config.Config) error {
	raw := strings.TrimSpace(cfg.Bootstrap.FirstSuperuser)
	password := cfg.Bootstrap.FirstSuperuserPassword
	if raw == "" {
		app.Log.Info("seed: FIRST_SUPERUSER not set, skipping superuser creation")
		return nil
	}
	if password == "" {
		app.Log.Warn("seed: FIRST_SUPERUSER set but FIRST_SUPERUSER_PASSWORD empty, skipping")
		return nil
	}

	username, email := superuserIdentity(raw)
	users := app.Services.Users

	// 幂等：按用户名或邮箱命中即跳过（登录同样支持两者）
	if _, err := users.GetByUsername(ctx, raw); err == nil {
		app.Log.Info("seed: superuser already exists, skipping", "username", username, "email", email)
		return nil
	} else if !errors.Is(err, errors.KindNotFound) {
		return fmt.Errorf("seed: check superuser existence: %w", err)
	}
	if raw != username {
		if _, err := users.GetByUsername(ctx, username); err == nil {
			app.Log.Info("seed: superuser already exists, skipping", "username", username)
			return nil
		} else if !errors.Is(err, errors.KindNotFound) {
			return fmt.Errorf("seed: check superuser existence: %w", err)
		}
	}

	// 找 admin 角色 ID
	var adminRoleID int64
	roles, err := app.Services.Roles.List(ctx)
	if err != nil {
		return fmt.Errorf("seed: list roles: %w", err)
	}
	for _, r := range roles {
		if r.Code() == "admin" {
			adminRoleID = r.ID()
			break
		}
	}
	if adminRoleID == 0 {
		return fmt.Errorf("seed: admin role not found (roles may not be seeded)")
	}

	user, err := users.Create(ctx, appidentity.CreateUserCommand{
		Username: username,
		Password: password,
		Email:    email,
		RoleIDs:  []int64{adminRoleID},
	})
	if err != nil {
		return fmt.Errorf("seed: create superuser: %w", err)
	}
	app.Log.Info("seed: superuser created",
		"username", username, "email", email, "user_id", user.ID())
	return nil
}

// superuserIdentity 把 FIRST_SUPERUSER 拆成 username + email。
// Python 配置是 EmailStr，不能再拼接 @hermes.local。
func superuserIdentity(raw string) (username, email string) {
	if identity.ValidEmail(raw) {
		local, _, _ := strings.Cut(raw, "@")
		if n := utf8.RuneCountInString(local); n >= 3 && n <= 64 {
			return local, raw
		}
		return "admin", raw
	}
	return raw, raw + "@hermes.local"
}
