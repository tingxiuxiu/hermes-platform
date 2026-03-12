package services

import (
	"errors"

	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(name, email, password string) (*models.User, error)
	Login(email, password string) (string, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	GetUserByID(userID uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUsers(page, pageSize int) ([]models.User, int64, error)
	UpdateUser(userID uint, name string) error
	DeleteUser(userID uint) error
	AssignRolesToUser(userID uint, roleIDs []uint) error
}

type authService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *authService) Register(name, email, password string) (*models.User, error) {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	defaultRole, err := s.roleRepo.FindByName("普通用户")
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		user.Roles = []models.Role{*defaultRole}
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", err
	}

	err = auth.VerifyPassword(user.Password, password)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	token, err := auth.GenerateToken(user.ID, user.Email, roles)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	err = auth.VerifyPassword(user.Password, oldPassword)
	if err != nil {
		return errors.New("invalid old password")
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	err = s.userRepo.Update(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *authService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *authService) ListUsers(page, pageSize int) ([]models.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

func (s *authService) UpdateUser(userID uint, name string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Name = name
	return s.userRepo.Update(user)
}

func (s *authService) DeleteUser(userID uint) error {
	return s.userRepo.Delete(userID)
}

func (s *authService) AssignRolesToUser(userID uint, roleIDs []uint) error {
	return s.userRepo.AssignRoles(userID, roleIDs)
}
