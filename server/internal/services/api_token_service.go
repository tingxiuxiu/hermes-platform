package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"com.hermes.platform/internal/models"
	"gorm.io/gorm"
)

type APITokenService interface {
	CreateToken(userID uint, name string) (*models.APIToken, error)
	GetTokensByUser(userID uint) ([]models.APIToken, error)
	RevokeToken(tokenID uint, userID uint) error
	GetTokenByID(tokenID uint) (*models.APIToken, error)
}

type TokenValidator interface {
	ValidateToken(token string) (uint, error)
}

type apiTokenService struct {
	db *gorm.DB
}

func NewAPITokenService(db *gorm.DB) APITokenService {
	return &apiTokenService{
		db: db,
	}
}

func (s *apiTokenService) CreateToken(userID uint, name string) (*models.APIToken, error) {
	tokenStr := generateSecureToken(32)

	expiresAt := time.Now().Add(365 * 24 * time.Hour)

	token := &models.APIToken{
		UserID:    userID,
		Token:     tokenStr,
		Name:      name,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}

	err := s.db.Create(token).Error
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *apiTokenService) GetTokensByUser(userID uint) ([]models.APIToken, error) {
	var tokens []models.APIToken
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *apiTokenService) RevokeToken(tokenID uint, userID uint) error {
	result := s.db.Model(&models.APIToken{}).
		Where("id = ? AND user_id = ?", tokenID, userID).
		Update("is_revoked", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("token not found or unauthorized")
	}

	return nil
}

func (s *apiTokenService) GetTokenByID(tokenID uint) (*models.APIToken, error) {
	var token models.APIToken
	err := s.db.First(&token, tokenID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &token, nil
}

func (s *apiTokenService) ValidateToken(token string) (uint, error) {
	var apiToken models.APIToken

	err := s.db.Where("token = ?", token).First(&apiToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("invalid token")
		}
		return 0, err
	}

	if apiToken.IsRevoked {
		return 0, errors.New("token has been revoked")
	}

	if time.Now().After(apiToken.ExpiresAt) {
		return 0, errors.New("token has expired")
	}

	return apiToken.UserID, nil
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
