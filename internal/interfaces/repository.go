// Package interfaces определяет контракты для слоев приложения.
// Содержит интерфейсы репозитория и сервиса для обеспечения слабой связанности.
package interfaces

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/models"
	"time"
)

// Repository определяет контракт для работы с хранилищем данных.
// Реализации должны обеспечивать сохранность и целостность данных.
type Repository interface {
	// CreateUsers создает нового пользователя в системе.
	// Должен возвращать ошибку если пользователь с таким логином уже существует.
	CreateUsers(ctx context.Context, user *models.User) error

	// GetUserByLogin возвращает пользователя по логину.
	// Должен возвращать ErrUserNotFound если пользователь не существует.
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	// GetUserByID возвращает пользователя по идентификатору.
	// Должен возвращать ErrUserNotFound если пользователь не существует.
	GetUserByID(ctx context.Context, id string) (*models.User, error)

	// SaveSecretData сохраняет секретные данные пользователя.
	// Генерирует уникальный ID и временные метки если они не установлены.
	SaveSecretData(ctx context.Context, data *models.SecretItemData) error

	// GetSecretDataByID возвращает секретные данные по ID.
	// Проверяет что данные принадлежат указанному пользователю.
	GetSecretDataByID(ctx context.Context, id, userID string) (*models.SecretItemData, error)

	// GetSecretDataByName возвращает секретные данные по имени.
	// Удобно для поиска данных по человеко-читаемому имени.
	GetSecretDataByName(ctx context.Context, name, userID string) (*models.SecretItemData, error)

	// GetUserSecretData возвращает все секретные данные пользователя измененные после указанного времени.
	// Используется для синхронизации между клиентами.
	GetUserSecretData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error)

	// UpdateSecretData обновляет существующие секретные данные.
	// Автоматически увеличивает версию и обновляет временные метки.
	UpdateSecretData(ctx context.Context, data *models.SecretItemData) error

	// DeleteSecretData удаляет секретные данные по ID.
	// Проверяет принадлежность данных пользователю.
	DeleteSecretData(ctx context.Context, id string, userID string) error

	// DeleteSecretDataByName удаляет секретные данные по имени.
	// Удобно для удаления данных без знания их ID.
	DeleteSecretDataByName(ctx context.Context, name, userID string) error

	// Ping проверяет соединение с хранилищем.
	// Используется для health checks.
	Ping(ctx context.Context) error
}
