package services

import (
	"com.hermes.platform/internal/repository"
)

// PermissionService 权限服务接口
type PermissionService interface {
	CheckPermission(userID uint, permissionCode string) (bool, error)
}

// permissionService 权限服务实现
type permissionService struct {
	userRepo repository.UserRepository
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(userRepo repository.UserRepository) PermissionService {
	return &permissionService{userRepo: userRepo}
}

// CheckPermission 检查用户是否具有特定的权限
func (s *permissionService) CheckPermission(userID uint, permissionCode string) (bool, error) {
	// 查找用户及其角色和权限
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}

	// 检查用户的角色是否具有所需的权限
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Code == permissionCode {
				return true, nil
			}
		}
	}

	return false, nil
}
