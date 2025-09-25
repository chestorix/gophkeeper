package service

import (
	"context"
	"github.com/chestorix/gophkeeper/cmd/internal/interfaces"
	"github.com/chestorix/gophkeeper/cmd/internal/models"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo   interfaces.Repository
	logger *logrus.Logger
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewService(repo interfaces.Repository, logger *logrus.Logger) Service {
	return Service{repo: repo, logger: logger}
}

func (s *Service) Test() string {
	return s.repo.Test()
}
