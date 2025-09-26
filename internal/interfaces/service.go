package interfaces

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/models"
	"time"
)

type Service interface {
	Register(ctx context.Context, login string, password string) (*models.AuthResponse, error)
	Login(ctx context.Context, login string, password string) (*models.AuthResponse, error)
	ValidateToken(token string) (string, error)

	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	SaveData(ctx context.Context, userID string, data *models.SecretItemData) error
	GetData(ctx context.Context, userID string, dataID string) (*models.SecretItemData, error)
	GetUserData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error)

	UpdateData(ctx context.Context, userID string, data *models.SecretItemData) error
	DeleteData(ctx context.Context, userID string, dataID string) error

	SyncData(ctx context.Context, userID string, req *models.SyncRequest) (*models.SyncResponse, error)
}
