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

type Handler struct {
	service interfaces.Service
	logger  *logrus.Logger
}

func NewHandler(service interfaces.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

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
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorf("Encoding failed %v", err)
		http.Error(w, "encoding failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

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
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorf("Encoding failed %v", err)
		http.Error(w, "encoding failed", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

}

func (h *Handler) ListData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	h.logger.Info(userID)
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

func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.Errorf("Failed to get userID from context: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	dataID := chi.URLParam(r, "dataID")

	data, err := h.service.GetData(r.Context(), userID, dataID)
	if err != nil {
		h.logger.Errorf("Getting data failed: %v", err)
		http.Error(w, fmt.Sprintf("Getting data failed: %v", err), http.StatusNotFound)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logger.Errorf("Encoding failed %v", err)
		http.Error(w, "encoding failed", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

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

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
