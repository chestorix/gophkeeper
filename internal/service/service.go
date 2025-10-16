// Package service реализует бизнес-логику приложения GophKeeper.
// Обеспечивает аутентификацию, управление пользователями и секретными данными.
package service

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/errors"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// Service предоставляет методы для работы с пользователями и данными.
// Реализует интерфейс interfaces.Service.
type Service struct {
	repo      interfaces.Repository
	logger    *logrus.Logger
	jwtSecret string
}

// Claims содержит кастомные claims для JWT токенов.
// Включает ID пользователя и логин для упрощения доступа.
type Claims struct {
	UserID string `json:"user_id"` // ID пользователя
	Login  string `json:"login"`   // Логин пользователя
	jwt.RegisteredClaims
}

// NewService создает новый экземпляр Service с заданными зависимостями.
//
// repo: репозиторий для работы с данными
// logger: логгер для записи событий
// jwtSecret: секретный ключ для подписи JWT токенов
func NewService(repo interfaces.Repository, logger *logrus.Logger, jwtSecret string) Service {
	return Service{repo: repo, logger: logger, jwtSecret: jwtSecret}
}

// hashPassword создает SHA256 хэш от пароля.
// Используется для безопасного хранения паролей в базе данных.
func (s *Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword проверяет соответствие пароля и хэша.
func (s *Service) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken создает JWT токен для пользователя с временем жизни 24 часа.
// Токен содержит ID пользователя и логин в claims.
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

// Register регистрирует нового пользователя в системе.
// Проверяет уникальность логина и возвращает JWT токен при успешной регистрации.
//
// ctx: контекст для отмены операций
// login: логин нового пользователя
// password: пароль нового пользователя
//
// Возвращает AuthResponse с токеном или ошибку если регистрация не удалась.
func (s *Service) Register(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	existingUser, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return nil, errors.ErrCreateUserFailed
	}

	user := &models.User{
		Login:        login,
		PasswordHash: passwordHash,
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

// Login аутентифицирует пользователя и возвращает JWT токен.
// Проверяет соответствие логина и пароля.
//
// ctx: контекст для отмены операций
// login: логин пользователя
// password: пароль пользователя
//
// Возвращает AuthResponse с токеном или ошибку если аутентификация не удалась.
func (s *Service) Login(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	if !s.verifyPassword(password, user.PasswordHash) {
		return nil, errors.ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.ErrGenerateTokenFailed
	}
	return &models.AuthResponse{Token: token}, nil
}

// ValidateToken проверяет валидность JWT токена и возвращает ID пользователя.
// Используется middleware для защиты маршрутов.
//
// tokenString: JWT токен из заголовка Authorization
//
// Возвращает ID пользователя если токен валиден, или ошибку.
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

// GetUserByLogin возвращает пользователя по логину.
// В основном используется для внутренних нужд сервиса.
func (s *Service) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

// SaveData сохраняет новые секретные данные для пользователя.
// Автоматически устанавливает UserID и временные метки.
//
// userID: ID пользователя-владельца
// data: данные для сохранения
func (s *Service) SaveData(ctx context.Context, userID string, data *models.SecretItemData) error {
	data.UserID = userID
	return s.repo.SaveSecretData(ctx, data)
}

// GetData возвращает секретные данные по ID.
// Проверяет что данные принадлежат указанному пользователю.
func (s *Service) GetData(ctx context.Context, userID string, dataID string) (*models.SecretItemData, error) {
	return s.repo.GetSecretDataByID(ctx, dataID, userID)
}

// GetUserData возвращает все секретные данные пользователя измененные после lastSync.
// Используется для синхронизации между клиентами.
func (s *Service) GetUserData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error) {
	return s.repo.GetUserSecretData(ctx, userID, lastSync)
}

// UpdateData обновляет существующие секретные данные.
// Проверяет принадлежность данных пользователю и увеличивает версию.
func (s *Service) UpdateData(ctx context.Context, userID string, data *models.SecretItemData) error {
	data.UserID = userID
	return s.repo.UpdateSecretData(ctx, data)
}

// DeleteData удаляет секретные данные по ID.
// Проверяет принадлежность данных пользователю.
func (s *Service) DeleteData(ctx context.Context, userID string, dataID string) error {
	return s.repo.DeleteSecretData(ctx, dataID, userID)
}

// SyncData выполняет двустороннюю синхронизацию данных.
// Получает данные с клиента и возвращает данные с сервера.
//
// Сохраняет данные от клиента и возвращает все данные пользователя
// измененные после последней синхронизации.
func (s *Service) SyncData(ctx context.Context, userID string, req *models.SyncRequest) (*models.SyncResponse, error) {
	// Получаем данные с сервера, измененные после последней синхронизации
	serverData, err := s.GetUserData(ctx, userID, req.LastSync)
	if err != nil {
		return nil, errors.ErrGetSecretDataFailed
	}

	// Обрабатываем данные от клиента с проверкой конфликтов
	for i := range req.Data {
		clientItem := &req.Data[i]
		clientItem.UserID = userID

		// Проверяем существование данных на сервере
		serverItem, err := s.repo.GetSecretDataByName(ctx, clientItem.Name, userID)
		if err != nil && err != errors.ErrDataNotFound {
			s.logger.Warnf("Failed to check existing data: %v", err)
			continue
		}

		if serverItem != nil {
			// Проверяем конфликт версий
			if !clientItem.UpdatedAt.After(serverItem.UpdatedAt) || clientItem.Version <= serverItem.Version {
				s.logger.Debugf("Skipping older client data: %s", clientItem.Name)
				continue
			}

			// Обновляем существующие данные
			clientItem.ID = serverItem.ID
			if err := s.repo.UpdateSecretData(ctx, clientItem); err != nil {
				s.logger.Warnf("Failed to update secret data: %v", err)
			}
		} else {
			// Сохраняем новые данные
			if err := s.repo.SaveSecretData(ctx, clientItem); err != nil {
				s.logger.Warnf("Failed to save client secret data: %v", err)
			}
		}
	}

	return &models.SyncResponse{
		LastSync: time.Now(),
		Data:     serverData,
	}, nil
}

// GetDataByName возвращает секретные данные по имени.
// Удобно для поиска данных по человеко-читаемому имени.
func (s *Service) GetDataByName(ctx context.Context, userID string, name string) (*models.SecretItemData, error) {
	return s.repo.GetSecretDataByName(ctx, name, userID)
}

// DeleteDataByName удаляет секретные данные по имени.
// Удобно для удаления данных без знания их ID.
func (s *Service) DeleteDataByName(ctx context.Context, userID string, name string) error {
	return s.repo.DeleteSecretDataByName(ctx, name, userID)
}
