package api

import (
	"encoding/json"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/sirupsen/logrus"
	"net/http"
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
		http.Error(w, "registration failed", http.StatusBadRequest)
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
