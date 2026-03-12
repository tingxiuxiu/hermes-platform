package services

import (
	"testing"

	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
	"gorm.io/gorm"
)

type mockUserRepository struct {
	users  map[uint]*models.User
	nextID uint
}

func (m *mockUserRepository) Create(user *models.User) error {
	m.nextID++
	user.ID = m.nextID
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Update(user *models.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return gorm.ErrRecordNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Delete(id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) FindByID(id uint) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *mockUserRepository) FindByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) List(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, int64(len(users)), nil
}

func (m *mockUserRepository) AssignRoles(userID uint, roleIDs []uint) error {
	return nil
}

func newMockUserRepository() repository.UserRepository {
	return &mockUserRepository{
		users:  make(map[uint]*models.User),
		nextID: 0,
	}
}

type mockRoleRepository struct {
	roles map[string]*models.Role
}

func (m *mockRoleRepository) FindByName(name string) (*models.Role, error) {
	role, exists := m.roles[name]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return role, nil
}

func (m *mockRoleRepository) FindByID(id uint) (*models.Role, error) {
	for _, role := range m.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRoleRepository) List() ([]models.Role, error) {
	var roles []models.Role
	for _, role := range m.roles {
		roles = append(roles, *role)
	}
	return roles, nil
}

func newMockRoleRepository() repository.RoleRepository {
	return &mockRoleRepository{
		roles: map[string]*models.Role{
			"user": {Name: "user"},
		},
	}
}

func TestAuthService_Register(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	name := "testuser"
	email := "test@example.com"
	password := "test123"

	user, err := service.Register(name, email, password)
	if err != nil {
		t.Errorf("Register returned error: %v", err)
	}
	if user == nil {
		t.Error("Register returned nil user")
	}
	if user.Name != name {
		t.Errorf("Expected name %s, got %s", name, user.Name)
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	name := "testuser"
	email := "test@example.com"
	password := "test123"

	_, err := service.Register(name, email, password)
	if err != nil {
		t.Errorf("Register returned error: %v", err)
	}

	token, err := service.Login(email, password)
	if err != nil {
		t.Errorf("Login returned error: %v", err)
	}
	if token == "" {
		t.Error("Login returned empty token")
	}
}

func TestAuthService_Login_Invalid(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	name := "testuser"
	email := "test@example.com"
	password := "test123"

	_, err := service.Register(name, email, password)
	if err != nil {
		t.Errorf("Register returned error: %v", err)
	}

	_, err = service.Login(email, "wrongpassword")
	if err == nil {
		t.Error("Login should return error for wrong password")
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	name := "testuser"
	email := "test@example.com"
	password := "test123"
	newPassword := "newpassword123"

	user, err := service.Register(name, email, password)
	if err != nil {
		t.Errorf("Register returned error: %v", err)
	}

	err = service.ChangePassword(user.ID, password, newPassword)
	if err != nil {
		t.Errorf("ChangePassword returned error: %v", err)
	}

	token, err := service.Login(email, newPassword)
	if err != nil {
		t.Errorf("Login with new password returned error: %v", err)
	}
	if token == "" {
		t.Error("Login with new password returned empty token")
	}
}

func TestAuthService_ListUsers(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	_, _ = service.Register("user1", "user1@example.com", "password123")
	_, _ = service.Register("user2", "user2@example.com", "password123")

	users, total, err := service.ListUsers(1, 10)
	if err != nil {
		t.Errorf("ListUsers returned error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	user, _ := service.Register("testuser", "test@example.com", "password123")

	err := service.UpdateUser(user.ID, "newname")
	if err != nil {
		t.Errorf("UpdateUser returned error: %v", err)
	}

	updatedUser, _ := service.GetUserByID(user.ID)
	if updatedUser.Name != "newname" {
		t.Errorf("Expected name 'newname', got '%s'", updatedUser.Name)
	}
}

func TestAuthService_DeleteUser(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	user, _ := service.Register("testuser", "test@example.com", "password123")

	err := service.DeleteUser(user.ID)
	if err != nil {
		t.Errorf("DeleteUser returned error: %v", err)
	}

	_, err = service.GetUserByID(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestAuthService_AssignRolesToUser(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
	auth.InitJWT(cfg)

	userRepo := newMockUserRepository()
	roleRepo := newMockRoleRepository()
	service := NewAuthService(userRepo, roleRepo)

	user, _ := service.Register("testuser", "test@example.com", "password123")

	err := service.AssignRolesToUser(user.ID, []uint{1})
	if err != nil {
		t.Errorf("AssignRolesToUser returned error: %v", err)
	}
}
