// internal/interfaces/repository.go
package interfaces

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/models"
	"time"
)

type Repository interface {
	CreateUsers(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)

	SaveSecretData(ctx context.Context, data *models.SecretItemData) error
	GetSecretDataByID(ctx context.Context, id, userID string) (*models.SecretItemData, error)
	GetSecretDataByName(ctx context.Context, name, userID string) (*models.SecretItemData, error) // НОВЫЙ МЕТОД
	GetUserSecretData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error)
	UpdateSecretData(ctx context.Context, data *models.SecretItemData) error
	DeleteSecretData(ctx context.Context, id string, userID string) error
	DeleteSecretDataByName(ctx context.Context, name, userID string) error // НОВЫЙ МЕТОД

	Ping(ctx context.Context) error
}
