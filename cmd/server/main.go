package main

import (
	"context"
	"github.com/chestorix/gophkeeper/cmd/internal/api"
	"github.com/chestorix/gophkeeper/cmd/internal/config"
	"github.com/chestorix/gophkeeper/cmd/internal/repository"
	"github.com/chestorix/gophkeeper/cmd/internal/service"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg := config.Load()
	storage, err := repository.NewPostgres(cfg.DBURI)
	if err != nil {
		logger.Fatal(err)
	}
	service := service.NewService(storage, logger)
	server := api.NewServer(cfg, &service, logger)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
