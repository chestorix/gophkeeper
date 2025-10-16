// Package api предоставляет HTTP handlers для REST API GophKeeper.
// Обрабатывает входящие запросы, валидирует данные и возвращает ответы.
package api

import (
	"encoding/json"
	"fmt"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Handler предоставляет методы для обработки HTTP запросов.
// Инкапсулирует бизнес-логику через интерфейс Service.
type Handler struct {
	service interfaces.Service
	logger  *logrus.Logger
}

// NewHandler создает новый экземпляр Handler с заданными зависимостями.
//
// service: сервис для выполнения бизнес-логики
// logger: логгер для записи событий
func NewHandler(service interfaces.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// getUserIDFromContext извлекает ID пользователя из контекста запроса.
// Используется middleware аутентификации для установки userID.
func (h *Handler) getUserIDFromContext(r *http.Request) (string, error) {
	userID := r.Context().Value("userID")
	if userID == nil {
		return "", fmt.Errorf("userID not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("userID is not a string")
	}

	return userIDStr, nil
}

// Register обрабатывает запрос на регистрацию нового пользователя.
// POST /api/user/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}
	resp, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Errorf("Registration failed %v", err)
		http.Error(w, fmt.Sprintf("Registration failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Encoding failed %v", err)
	}
}

// Login обрабатывает запрос на аутентификацию пользователя.
// POST /api/user/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Errorf("Login failed %v", err)
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Encoding failed %v", err)
	}
}

// ListData возвращает список секретных данных пользователя.
// GET /api/data?last_sync=2023-01-01T00:00:00Z
func (h *Handler) ListData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	lastSyncStr := r.URL.Query().Get("last_sync")
	var lastSync time.Time
	if lastSyncStr != "" {
		var err error
		lastSync, err = time.Parse(time.RFC3339, lastSyncStr)
		if err != nil {
			http.Error(w, "invalid last_sync format", http.StatusBadRequest)
			return
		}
	}

	data, err := h.service.GetUserData(r.Context(), userID, lastSync)
	if err != nil {
		h.logger.Errorf("Failed to list data: %v", err)
		http.Error(w, fmt.Sprintf("failed to list data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// UpdateData обновляет существующие секретные данные.
// PUT /api/data/{id}
func (h *Handler) UpdateData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	dataID := chi.URLParam(r, "id")

	var data models.SecretItemData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	data.ID = dataID

	if err := h.service.UpdateData(r.Context(), userID, &data); err != nil {
		h.logger.Errorf("Failed to update data: %v", err)
		http.Error(w, fmt.Sprintf("failed to update data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SaveData сохраняет новые секретные данные.
// POST /api/data
func (h *Handler) SaveData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var data models.SecretItemData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.SaveData(r.Context(), userID, &data); err != nil {
		h.logger.Errorf("Saving data failed: %v", err)
		http.Error(w, fmt.Sprintf("Saving data failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GetData возвращает секретные данные по ID.
// GET /api/data/{id}
func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dataID := chi.URLParam(r, "id")

	data, err := h.service.GetData(r.Context(), userID, dataID)
	if err != nil {
		h.logger.Errorf("Getting data failed: %v", err)
		http.Error(w, fmt.Sprintf("Getting data failed: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("Encoding failed %v", err)
	}
}

// SyncData выполняет синхронизацию данных между клиентом и сервером.
// POST /api/sync
func (h *Handler) SyncData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.service.SyncData(r.Context(), userID, &req)
	if err != nil {
		h.logger.Errorf("Syncing data failed: %v", err)
		http.Error(w, fmt.Sprintf("Syncing data failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorf("Encoding failed %v", err)
		http.Error(w, "encoding failed", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteData удаляет секретные данные по ID.
// DELETE /api/data/{id}
func (h *Handler) DeleteData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	dataID := chi.URLParam(r, "id")

	if err := h.service.DeleteData(r.Context(), userID, dataID); err != nil {
		h.logger.Errorf("Failed to delete data: %v", err)
		http.Error(w, fmt.Sprintf("failed to delete data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Health проверяет доступность сервера.
// GET /health
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetDataByName возвращает секретные данные по имени.
// GET /api/data/name/{name}
func (h *Handler) GetDataByName(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	name := chi.URLParam(r, "name")

	data, err := h.service.GetDataByName(r.Context(), userID, name)
	if err != nil {
		h.logger.Errorf("Getting data by name failed: %v", err)
		http.Error(w, fmt.Sprintf("Getting data failed: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logger.Errorf("Encoding failed %v", err)
		http.Error(w, "encoding failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteDataByName удаляет секретные данные по имени.
// DELETE /api/data/name/{name}
func (h *Handler) DeleteDataByName(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := chi.URLParam(r, "name")

	if err := h.service.DeleteDataByName(r.Context(), userID, name); err != nil {
		h.logger.Errorf("Failed to delete data by name: %v", err)
		http.Error(w, fmt.Sprintf("failed to delete data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
