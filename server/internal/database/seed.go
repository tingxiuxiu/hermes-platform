package database

import (
	"log"

	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/models"
	"gorm.io/gorm"
)

// SeedData 初始化默认数据
func SeedData(db *gorm.DB) {
	// 初始化权限数据
	initPermissions(db)

	// 初始化角色数据
	initRoles(db)

	// 初始化默认用户
	initUsers(db)
}

// 初始化权限数据
func initPermissions(db *gorm.DB) {
	permissions := []models.Permission{
		{Name: "查看测试任务", Description: "查看测试任务", Code: "test_task:view"},
		{Name: "创建测试任务", Description: "创建测试任务", Code: "test_task:create"},
		{Name: "编辑测试任务", Description: "编辑测试任务", Code: "test_task:edit"},
		{Name: "删除测试任务", Description: "删除测试任务", Code: "test_task:delete"},
		{Name: "查看测试详情", Description: "查看测试详情", Code: "test_detail:view"},
		{Name: "创建测试详情", Description: "创建测试详情", Code: "test_detail:create"},
		{Name: "编辑测试详情", Description: "编辑测试详情", Code: "test_detail:edit"},
		{Name: "删除测试详情", Description: "删除测试详情", Code: "test_detail:delete"},
		{Name: "查看测试记录", Description: "查看测试记录", Code: "test_record:view"},
		{Name: "创建测试记录", Description: "创建测试记录", Code: "test_record:create"},
		{Name: "编辑测试记录", Description: "编辑测试记录", Code: "test_record:edit"},
		{Name: "删除测试记录", Description: "删除测试记录", Code: "test_record:delete"},
		{Name: "查看测试步骤详情", Description: "查看测试步骤详情", Code: "test_step_detail:view"},
		{Name: "创建测试步骤详情", Description: "创建测试步骤详情", Code: "test_step_detail:create"},
		{Name: "编辑测试步骤详情", Description: "编辑测试步骤详情", Code: "test_step_detail:edit"},
		{Name: "删除测试步骤详情", Description: "删除测试步骤详情", Code: "test_step_detail:delete"},
		{Name: "管理用户", Description: "管理系统用户", Code: "user:manage"},
		{Name: "管理角色", Description: "管理系统角色", Code: "role:manage"},
		{Name: "管理权限", Description: "管理系统权限", Code: "permission:manage"},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := db.Where("code = ?", permission.Code).FirstOrCreate(&existingPermission, permission).Error; err != nil {
			log.Printf("Failed to create permission %s: %v", permission.Name, err)
		}
	}
}

// 初始化角色数据
func initRoles(db *gorm.DB) {
	roles := []struct {
		Role        models.Role
		Permissions []string
	}{
		{
			Role: models.Role{
				Name:        "普通用户",
				Description: "只能查看测试数据",
			},
			Permissions: []string{
				"test_task:view",
				"test_detail:view",
				"test_record:view",
				"test_step_detail:view",
			},
		},
		{
			Role: models.Role{
				Name:        "测试用户",
				Description: "可以创建、查看、编辑测试数据",
			},
			Permissions: []string{
				"test_task:view",
				"test_task:create",
				"test_task:edit",
				"test_detail:view",
				"test_detail:create",
				"test_detail:edit",
				"test_record:view",
				"test_record:create",
				"test_record:edit",
				"test_step_detail:view",
				"test_step_detail:create",
				"test_step_detail:edit",
			},
		},
		{
			Role: models.Role{
				Name:        "管理员",
				Description: "拥有所有权限",
			},
			Permissions: []string{
				"test_task:view",
				"test_task:create",
				"test_task:edit",
				"test_task:delete",
				"test_detail:view",
				"test_detail:create",
				"test_detail:edit",
				"test_detail:delete",
				"test_record:view",
				"test_record:create",
				"test_record:edit",
				"test_record:delete",
				"test_step_detail:view",
				"test_step_detail:create",
				"test_step_detail:edit",
				"test_step_detail:delete",
				"user:manage",
				"role:manage",
				"permission:manage",
			},
		},
	}

	for _, roleData := range roles {
		var existingRole models.Role
		if err := db.Where("name = ?", roleData.Role.Name).FirstOrCreate(&existingRole, roleData.Role).Error; err != nil {
			log.Printf("Failed to create role %s: %v", roleData.Role.Name, err)
			continue
		}

		var permissions []models.Permission
		if err := db.Where("code IN ?", roleData.Permissions).Find(&permissions).Error; err != nil {
			log.Printf("Failed to find permissions for role %s: %v", roleData.Role.Name, err)
			continue
		}

		if err := db.Model(&existingRole).Association("Permissions").Replace(permissions); err != nil {
			log.Printf("Failed to assign permissions to role %s: %v", roleData.Role.Name, err)
		}
	}
}

// 初始化默认管理员用户
func initUsers(db *gorm.DB) {
	users := []struct {
		Name      string
		Email     string
		Password  string
		RoleNames []string
	}{
		{
			Name:      "普通用户",
			Email:     "user@example.com",
			Password:  "user123",
			RoleNames: []string{"普通用户"},
		},
		{
			Name:      "测试用户",
			Email:     "tester@example.com",
			Password:  "tester123",
			RoleNames: []string{"测试用户"},
		},
		{
			Name:      "管理员",
			Email:     "admin@example.com",
			Password:  "admin123",
			RoleNames: []string{"管理员"},
		},
	}

	for _, userData := range users {
		var existingUser models.User
		if err := db.Where("email = ?", userData.Email).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				hashedPassword, err := auth.HashPassword(userData.Password)
				if err != nil {
					log.Printf("Failed to hash password for user %s: %v", userData.Email, err)
					continue
				}

				user := models.User{
					Name:     userData.Name,
					Email:    userData.Email,
					Password: hashedPassword,
				}

				if err := db.Create(&user).Error; err != nil {
					log.Printf("Failed to create user %s: %v", userData.Email, err)
					continue
				}

				var roles []models.Role
				if err := db.Where("name IN ?", userData.RoleNames).Find(&roles).Error; err != nil {
					log.Printf("Failed to find roles for user %s: %v", userData.Email, err)
					continue
				}

				if len(roles) == 0 {
					log.Printf("No roles found for user %s with role names %v", userData.Email, userData.RoleNames)
					continue
				}

				if err := db.Model(&user).Association("Roles").Replace(roles); err != nil {
					log.Printf("Failed to assign roles to user %s: %v", userData.Email, err)
				}
			} else {
				log.Printf("Failed to query user %s: %v", userData.Email, err)
			}
		}
	}
}
