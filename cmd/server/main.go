package main

import (
	"context"
	"github.com/chestorix/gophkeeper/internal/api"
	"github.com/chestorix/gophkeeper/internal/config"
	"github.com/chestorix/gophkeeper/internal/repository"
	"github.com/chestorix/gophkeeper/internal/service"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg := config.Load()
	storage, err := repository.NewPostgres(cfg.DBURI)
	if err != nil {
		logger.Fatal("Failed to connect database", err)
	}
	serv := service.NewService(storage, logger, cfg.JWTSecret)
	server := api.NewServer(cfg, &serv, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error: ", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Info("Server exiting")
}
