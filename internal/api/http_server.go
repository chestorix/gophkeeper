package api

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/config"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	server *http.Server
	logger *logrus.Logger
}

func NewServer(cfg *config.ServerConfig, service interfaces.Service, logger *logrus.Logger) *Server {
	handler := NewHandler(service, logger)
	router := NewRouter(logger)
	router.SetupRoutes(handler)

	return &Server{
		server: &http.Server{
			Addr:    cfg.RunAddress,
			Handler: router,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting server...")
	s.logger.Infoln("Server listening on address: ", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
