package api

import (
	"fmt"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Handler struct {
	service interfaces.Service
	logger  *logrus.Logger
	dbURL   string
}

func NewHandler(service interfaces.Service, logger *logrus.Logger, dbURL string) *Handler {
	return &Handler{service: service,
		logger: logger,
		dbURL:  dbURL,
	}
}

func (h *Handler) GetTest(w http.ResponseWriter, r *http.Request) {
	test := h.service.Test()
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, test)
}
