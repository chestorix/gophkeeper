package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/chestorix/gophkeeper/internal/errors"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	repo      interfaces.Repository
	logger    *logrus.Logger
	jwtSecret string
}

type Claims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func NewService(repo interfaces.Repository, logger *logrus.Logger, jwtSecret string) Service {
	return Service{repo: repo, logger: logger, jwtSecret: jwtSecret}
}

func (s *Service) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (s *Service) generateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Login:  user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) Register(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	existingUser, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}
	user := &models.User{
		Login:        login,
		PasswordHash: s.hashPassword(password),
	}
	if err := s.repo.CreateUsers(ctx, user); err != nil {
		return nil, errors.ErrCreateUserFailed
	}
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.ErrGenerateTokenFailed
	}
	return &models.AuthResponse{Token: token}, nil
}
func (s *Service) Login(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}
	if user.PasswordHash != s.hashPassword(password) {
		return nil, errors.ErrInvalidCredentials
	}
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.ErrGenerateTokenFailed
	}
	return &models.AuthResponse{Token: token}, nil
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.ErrInvalidToken
	}
	return claims.UserID, nil
}
func (s *Service) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

func (s *Service) SaveData(ctx context.Context, userID string, data *models.SecretItemData) error {
	data.UserID = userID
	return s.repo.SaveSecretData(ctx, data)
}

func (s *Service) GetData(ctx context.Context, userID string, dataID string) (*models.SecretItemData, error) {
	return s.repo.GetSecretDataByID(ctx, dataID, userID)
}

func (s *Service) GetUserData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error) {
	return s.repo.GetUserSecretData(ctx, userID, lastSync)
}

func (s *Service) UpdateData(ctx context.Context, userID string, data *models.SecretItemData) error {
	data.UserID = userID
	return s.repo.UpdateSecretData(ctx, data)
}

func (s *Service) DeleteData(ctx context.Context, userID string, dataID string) error {
	return s.repo.DeleteSecretData(ctx, dataID, userID)
}

func (s *Service) SyncData(ctx context.Context, userID string, req *models.SyncRequest) (*models.SyncResponse, error) {
	serverData, err := s.GetUserData(ctx, userID, req.LastSync)
	if err != nil {
		return nil, errors.ErrGetSecretDataFailed
	}
	for i := range req.Data {
		if err := s.SaveData(ctx, userID, &req.Data[i]); err != nil {
			s.logger.Warnf("Failed to save client secret data: %v", errors.ErrSaveSecretDataFailed)
		}
	}
	return &models.SyncResponse{
		LastSync: time.Now(),
		Data:     serverData,
	}, nil
}

func (s *Service) GetDataByName(ctx context.Context, userID string, name string) (*models.SecretItemData, error) {
	return s.repo.GetSecretDataByName(ctx, name, userID)
}

func (s *Service) DeleteDataByName(ctx context.Context, userID string, name string) error {
	return s.repo.DeleteSecretDataByName(ctx, name, userID)
}
