package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func NewService

func (s *Service) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewService(repo interfaces.Repository, logger *logrus.Logger, jwtSecret string) Service {
	return Service{repo: repo, logger: logger, jwtSecret: jwtSecret}
}

func (s *Service) hashPassword(password string) (string) {
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

func (s *Service) Register (ctx context.Context, login,password string) (*models.AuthResponse, error) {
	existingUser,err := s.repo.GetUserByLogin(ctx, login)
	if err == nil && existingUser != nil {
		return nil,errors.ErrUserAlreadyExists
	}
	user := &models.User{
		Login: login,
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
