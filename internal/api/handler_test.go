package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService реализует interfaces.Service для тестирования хендлеров
type MockService struct {
	mock.Mock
}

func (m *MockService) Register(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockService) Login(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockService) ValidateToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockService) SaveData(ctx context.Context, userID string, data *models.SecretItemData) error {
	args := m.Called(ctx, userID, data)
	return args.Error(0)
}

func (m *MockService) GetData(ctx context.Context, userID string, dataID string) (*models.SecretItemData, error) {
	args := m.Called(ctx, userID, dataID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretItemData), args.Error(1)
}

func (m *MockService) GetDataByName(ctx context.Context, userID string, name string) (*models.SecretItemData, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretItemData), args.Error(1)
}

func (m *MockService) GetUserData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error) {
	args := m.Called(ctx, userID, lastSync)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SecretItemData), args.Error(1)
}

func (m *MockService) UpdateData(ctx context.Context, userID string, data *models.SecretItemData) error {
	args := m.Called(ctx, userID, data)
	return args.Error(0)
}

func (m *MockService) DeleteData(ctx context.Context, userID string, dataID string) error {
	args := m.Called(ctx, userID, dataID)
	return args.Error(0)
}

func (m *MockService) DeleteDataByName(ctx context.Context, userID string, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockService) SyncData(ctx context.Context, userID string, req *models.SyncRequest) (*models.SyncResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SyncResponse), args.Error(1)
}

func TestHandler_Register(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name: "Successful registration",
			requestBody: models.AuthRequest{
				Login:    "testuser",
				Password: "testpass",
			},
			setupMock: func(ms *MockService) {
				ms.On("Register", mock.Anything, "testuser", "testpass").Return(&models.AuthResponse{
					Token: "test-token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"login": 123, // неверный тип
			},
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty login and password",
			requestBody:    models.AuthRequest{Login: "", Password: ""},
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Registration failed - user already exists",
			requestBody: models.AuthRequest{
				Login:    "existinguser",
				Password: "testpass",
			},
			setupMock: func(ms *MockService) {
				ms.On("Register", mock.Anything, "existinguser", "testpass").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name: "Successful login",
			requestBody: models.AuthRequest{
				Login:    "testuser",
				Password: "testpass",
			},
			setupMock: func(ms *MockService) {
				ms.On("Login", mock.Anything, "testuser", "testpass").Return(&models.AuthResponse{
					Token: "test-token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request body",
			requestBody:    "invalid json",
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Login failed - invalid credentials",
			requestBody: models.AuthRequest{
				Login:    "wronguser",
				Password: "wrongpass",
			},
			setupMock: func(ms *MockService) {
				ms.On("Login", mock.Anything, "wronguser", "wrongpass").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ListData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		lastSyncParam  string
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:          "Successful list data without lastSync",
			lastSyncParam: "",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetUserData", mock.Anything, "user123", time.Time{}).Return([]models.SecretItemData{
					{ID: "1", Name: "Test Data 1"},
					{ID: "2", Name: "Test Data 2"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Successful list data with lastSync",
			lastSyncParam: "2023-01-01T00:00:00Z",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
				ms.On("GetUserData", mock.Anything, "user123", expectedTime).Return([]models.SecretItemData{
					{ID: "1", Name: "Test Data 1"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Invalid lastSync format",
			lastSyncParam: "invalid-date",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "No authorization header",
			lastSyncParam:  "",
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "Service error",
			lastSyncParam: "",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetUserData", mock.Anything, "user123", time.Time{}).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			url := "/api/data"
			if tt.lastSyncParam != "" {
				url += "?last_sync=" + tt.lastSyncParam
			}

			req := httptest.NewRequest("GET", url, nil)
			if tt.name != "No authorization header" {
				req.Header.Set("Authorization", "Bearer valid-token")
			}

			// Создаем роутер и добавляем middleware аутентификации
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Get("/api/data", handler.ListData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_SaveData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	testData := models.SecretItemData{
		Type:     models.LoginPassword,
		Name:     "Test Data",
		Metadata: "Test metadata",
		Data:     []byte("test data"),
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "Successful save data",
			requestBody: testData,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("SaveData", mock.Anything, "user123", mock.MatchedBy(func(data *models.SecretItemData) bool {
					return data.Name == "Test Data" && data.Type == models.LoginPassword
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid request body",
			requestBody:    "invalid json",
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Save data failed",
			requestBody: testData,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("SaveData", mock.Anything, "user123", mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/data", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с middleware
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Post("/api/data", handler.SaveData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	testData := &models.SecretItemData{
		ID:   "data123",
		Name: "Test Data",
		Type: models.LoginPassword,
	}

	tests := []struct {
		name           string
		dataID         string
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:   "Successful get data by ID",
			dataID: "data123",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetData", mock.Anything, "user123", "data123").Return(testData, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Data not found",
			dataID: "nonexistent",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetData", mock.Anything, "user123", "nonexistent").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest("GET", "/api/data/"+tt.dataID, nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с chi для параметров URL
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Get("/api/data/{id}", handler.GetData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.SecretItemData
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "data123", response.ID)
				assert.Equal(t, "Test Data", response.Name)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	testData := models.SecretItemData{
		Name:     "Updated Data",
		Type:     models.LoginPassword,
		Metadata: "Updated metadata",
	}

	tests := []struct {
		name           string
		dataID         string
		requestBody    interface{}
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "Successful update data",
			dataID:      "data123",
			requestBody: testData,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("UpdateData", mock.Anything, "user123", mock.MatchedBy(func(data *models.SecretItemData) bool {
					return data.ID == "data123" && data.Name == "Updated Data"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request body",
			dataID:         "data123",
			requestBody:    "invalid json",
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Update data failed",
			dataID:      "data123",
			requestBody: testData,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("UpdateData", mock.Anything, "user123", mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/api/data/"+tt.dataID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с chi для параметров URL
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Put("/api/data/{id}", handler.UpdateData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		dataID         string
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:   "Successful delete data",
			dataID: "data123",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("DeleteData", mock.Anything, "user123", "data123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "Delete data failed",
			dataID: "data123",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("DeleteData", mock.Anything, "user123", "data123").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest("DELETE", "/api/data/"+tt.dataID, nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с chi для параметров URL
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Delete("/api/data/{id}", handler.DeleteData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_SyncData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	syncRequest := models.SyncRequest{
		LastSync: time.Now().Add(-1 * time.Hour),
		Data: []models.SecretItemData{
			{ID: "client1", Name: "Client Data"},
		},
	}

	syncResponse := &models.SyncResponse{
		LastSync: time.Now(),
		Data: []models.SecretItemData{
			{ID: "server1", Name: "Server Data"},
		},
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "Successful sync data",
			requestBody: syncRequest,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("SyncData", mock.Anything, "user123", mock.MatchedBy(func(req *models.SyncRequest) bool {
					return len(req.Data) == 1 && req.Data[0].Name == "Client Data"
				})).Return(syncResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request body",
			requestBody:    "invalid json",
			setupMock:      func(ms *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Sync data failed",
			requestBody: syncRequest,
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("SyncData", mock.Anything, "user123", mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/sync", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с middleware
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Post("/api/sync", handler.SyncData)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.SyncResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response.Data, 1)
				assert.Equal(t, "Server Data", response.Data[0].Name)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Health(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockService := &MockService{}
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestHandler_GetDataByName(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	testData := &models.SecretItemData{
		ID:   "data123",
		Name: "Test Data",
		Type: models.LoginPassword,
	}

	tests := []struct {
		name           string
		dataName       string
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:     "Successful get data by name",
			dataName: "test-data",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetDataByName", mock.Anything, "user123", "test-data").Return(testData, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Data not found by name",
			dataName: "nonexistent",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("GetDataByName", mock.Anything, "user123", "nonexistent").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest("GET", "/api/data/name/"+tt.dataName, nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с chi для параметров URL
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Get("/api/data/name/{name}", handler.GetDataByName)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteDataByName(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name           string
		dataName       string
		setupMock      func(*MockService)
		expectedStatus int
	}{
		{
			name:     "Successful delete data by name",
			dataName: "test-data",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("DeleteDataByName", mock.Anything, "user123", "test-data").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "Delete data by name failed",
			dataName: "test-data",
			setupMock: func(ms *MockService) {
				ms.On("ValidateToken", "valid-token").Return("user123", nil)
				ms.On("DeleteDataByName", mock.Anything, "user123", "test-data").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			tt.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest("DELETE", "/api/data/name/"+tt.dataName, nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			// Создаем роутер с chi для параметров URL
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := r.Header.Get("Authorization")
					if token == "" {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if strings.HasPrefix(token, "Bearer ") {
						token = strings.TrimPrefix(token, "Bearer ")
					}
					userID, err := mockService.ValidateToken(token)
					if err != nil {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					ctx := context.WithValue(r.Context(), "userID", userID)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Delete("/api/data/name/{name}", handler.DeleteDataByName)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_getUserIDFromContext(t *testing.T) {
	logger := logrus.New()
	handler := NewHandler(nil, logger)

	tests := []struct {
		name          string
		contextValue  interface{}
		expectedError bool
		expectedID    string
	}{
		{
			name:          "Valid user ID in context",
			contextValue:  "user123",
			expectedError: false,
			expectedID:    "user123",
		},
		{
			name:          "No user ID in context",
			contextValue:  nil,
			expectedError: true,
			expectedID:    "",
		},
		{
			name:          "Invalid user ID type in context",
			contextValue:  123, // не строка
			expectedError: true,
			expectedID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.contextValue != nil {
				ctx := context.WithValue(req.Context(), "userID", tt.contextValue)
				req = req.WithContext(ctx)
			}

			userID, err := handler.getUserIDFromContext(req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
