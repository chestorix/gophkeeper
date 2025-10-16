package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"

	"github.com/chestorix/gophkeeper/internal/errors"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository реализует interfaces.Repository для тестирования
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUsers(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) SaveSecretData(ctx context.Context, data *models.SecretItemData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRepository) GetSecretDataByID(ctx context.Context, id, userID string) (*models.SecretItemData, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretItemData), args.Error(1)
}

func (m *MockRepository) GetSecretDataByName(ctx context.Context, name, userID string) (*models.SecretItemData, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretItemData), args.Error(1)
}

func (m *MockRepository) GetUserSecretData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error) {
	args := m.Called(ctx, userID, lastSync)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SecretItemData), args.Error(1)
}

func (m *MockRepository) UpdateSecretData(ctx context.Context, data *models.SecretItemData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRepository) DeleteSecretData(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRepository) DeleteSecretDataByName(ctx context.Context, name, userID string) error {
	args := m.Called(ctx, name, userID)
	return args.Error(0)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestService_Register(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name          string
		login         string
		password      string
		setupMock     func(*MockRepository)
		expectedError error
		wantToken     bool
	}{
		{
			name:     "Successful registration",
			login:    "testuser",
			password: "testpass",
			setupMock: func(mr *MockRepository) {
				mr.On("GetUserByLogin", mock.Anything, "testuser").Return(nil, errors.ErrUserNotFound)
				mr.On("CreateUsers", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
					return user.Login == "testuser" && user.PasswordHash != ""
				})).Return(nil)
			},
			expectedError: nil,
			wantToken:     true,
		},
		{
			name:     "User already exists",
			login:    "existinguser",
			password: "testpass",
			setupMock: func(mr *MockRepository) {
				mr.On("GetUserByLogin", mock.Anything, "existinguser").Return(&models.User{}, nil)
			},
			expectedError: errors.ErrUserAlreadyExists,
			wantToken:     false,
		},
		{
			name:     "Create user failed",
			login:    "testuser",
			password: "testpass",
			setupMock: func(mr *MockRepository) {
				mr.On("GetUserByLogin", mock.Anything, "testuser").Return(nil, errors.ErrUserNotFound)
				mr.On("CreateUsers", mock.Anything, mock.Anything).Return(errors.ErrCreateUserFailed)
			},
			expectedError: errors.ErrCreateUserFailed,
			wantToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo, logger, "test-secret")
			resp, err := service.Register(context.Background(), tt.login, tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.wantToken {
					assert.NotEmpty(t, resp.Token)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Правильный пароль
	correctPassword := "correctpass"

	tests := []struct {
		name          string
		login         string
		password      string
		setupMock     func(*MockRepository)
		expectedError error
		wantToken     bool
	}{
		{
			name:     "Successful login",
			login:    "testuser",
			password: "correctpass",
			setupMock: func(mr *MockRepository) {
				// Генерируем bcrypt хэш для правильного пароля
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

				mr.On("GetUserByLogin", mock.Anything, "testuser").Return(&models.User{
					ID:           "user-id",
					Login:        "testuser",
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			expectedError: nil,
			wantToken:     true,
		},
		{
			name:     "Invalid credentials - wrong password",
			login:    "testuser",
			password: "wrongpass",
			setupMock: func(mr *MockRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
				mr.On("GetUserByLogin", mock.Anything, "testuser").Return(&models.User{
					ID:           "user-id",
					Login:        "testuser",
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			expectedError: errors.ErrInvalidCredentials,
			wantToken:     false,
		},
		{
			name:     "User not found",
			login:    "nonexistent",
			password: "anypass",
			setupMock: func(mr *MockRepository) {
				mr.On("GetUserByLogin", mock.Anything, "nonexistent").Return(nil, errors.ErrUserNotFound)
			},
			expectedError: errors.ErrInvalidCredentials,
			wantToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo, logger, "test-secret")
			resp, err := service.Login(context.Background(), tt.login, tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.wantToken {
					assert.NotEmpty(t, resp.Token)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	logger := logrus.New()
	service := NewService(nil, logger, "test-secret")

	// Создаем валидный токен
	user := &models.User{ID: "user123", Login: "testuser"}
	token, err := service.generateToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		token         string
		expectedError error
		wantUserID    string
	}{
		{
			name:          "Valid token",
			token:         token,
			expectedError: nil,
			wantUserID:    "user123",
		},
		{
			name:          "Invalid token",
			token:         "invalid-token",
			expectedError: errors.ErrInvalidToken,
			wantUserID:    "",
		},
		{
			name:          "Empty token",
			token:         "",
			expectedError: errors.ErrInvalidToken,
			wantUserID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := service.ValidateToken(tt.token)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID, userID)
			}
		})
	}
}

func TestService_SaveData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, logger, "test-secret")

	data := &models.SecretItemData{
		Type:     models.LoginPassword,
		Name:     "Test Data",
		Metadata: "Test metadata",
		Data:     []byte("test data"),
	}

	mockRepo.On("SaveSecretData", mock.Anything, mock.MatchedBy(func(d *models.SecretItemData) bool {
		return d.UserID == "user123" && d.Name == "Test Data"
	})).Return(nil)

	err := service.SaveData(context.Background(), "user123", data)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, logger, "test-secret")

	expectedData := &models.SecretItemData{
		ID:     "data123",
		UserID: "user123",
		Type:   models.LoginPassword,
		Name:   "Test Data",
	}

	mockRepo.On("GetSecretDataByID", mock.Anything, "data123", "user123").Return(expectedData, nil)

	data, err := service.GetData(context.Background(), "user123", "data123")
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	mockRepo.AssertExpectations(t)
}

func TestService_SyncData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, logger, "test-secret")

	lastSync := time.Now().Add(-1 * time.Hour)
	clientData := []models.SecretItemData{
		{
			ID:   "client1",
			Type: models.LoginPassword,
			Name: "Client Data 1",
		},
	}

	serverData := []models.SecretItemData{
		{
			ID:   "server1",
			Type: models.TextData,
			Name: "Server Data 1",
		},
	}

	mockRepo.On("GetUserSecretData", mock.Anything, "user123", lastSync).Return(serverData, nil)
	mockRepo.On("SaveSecretData", mock.Anything, mock.AnythingOfType("*models.SecretItemData")).Return(nil).Times(len(clientData))

	req := &models.SyncRequest{
		LastSync: lastSync,
		Data:     clientData,
	}

	resp, err := service.SyncData(context.Background(), "user123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "server1", resp.Data[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, logger, "test-secret")

	mockRepo.On("DeleteSecretData", mock.Anything, "data123", "user123").Return(nil)

	err := service.DeleteData(context.Background(), "user123", "data123")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_hashPassword(t *testing.T) {
	service := NewService(nil, logrus.New(), "test-secret")

	password := "testpassword123"
	hash1, _ := service.hashPassword(password)
	hash2, _ := service.hashPassword(password)

	// Хэш должен быть одинаковым для одного пароля
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, password, hash1)
	assert.Len(t, hash1, 64) // SHA256 produces 64 character hex string
}

func TestService_GenerateToken(t *testing.T) {
	service := NewService(nil, logrus.New(), "test-secret")

	user := &models.User{
		ID:    "test-user-id",
		Login: "testuser",
	}

	token, err := service.generateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем что токен можно валидировать
	userID, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "test-user-id", userID)
}
