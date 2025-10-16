// Package interfaces определяет контракты для слоев приложения.
package interfaces

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/models"
	"time"
)

// Service определяет бизнес-логику приложения GophKeeper.
// Обеспечивает аутентификацию, авторизацию и управление данными.
type Service interface {
	// Register регистрирует нового пользователя в системе.
	// Проверяет уникальность логина и возвращает JWT токен.
	Register(ctx context.Context, login string, password string) (*models.AuthResponse, error)

	// Login аутентифицирует пользователя и возвращает JWT токен.
	// Проверяет соответствие логина и пароля.
	Login(ctx context.Context, login string, password string) (*models.AuthResponse, error)

	// ValidateToken проверяет валидность JWT токена.
	// Возвращает ID пользователя если токен валиден.
	ValidateToken(tokenString string) (string, error)

	// GetUserByLogin возвращает пользователя по логину.
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	// SaveData сохраняет новые секретные данные для пользователя.
	SaveData(ctx context.Context, userID string, data *models.SecretItemData) error

	// GetData возвращает секретные данные по ID.
	GetData(ctx context.Context, userID string, dataID string) (*models.SecretItemData, error)

	// GetDataByName возвращает секретные данные по имени.
	GetDataByName(ctx context.Context, userID string, name string) (*models.SecretItemData, error)

	// GetUserData возвращает все секретные данные пользователя.
	GetUserData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error)

	// UpdateData обновляет существующие секретные данные.
	UpdateData(ctx context.Context, userID string, data *models.SecretItemData) error

	// DeleteData удаляет секретные данные по ID.
	DeleteData(ctx context.Context, userID string, dataID string) error

	// DeleteDataByName удаляет секретные данные по имени.
	DeleteDataByName(ctx context.Context, userID string, name string) error

	// SyncData выполняет двустороннюю синхронизацию данных.
	SyncData(ctx context.Context, userID string, req *models.SyncRequest) (*models.SyncResponse, error)
}
