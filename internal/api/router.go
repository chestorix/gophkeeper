package api

import (
	mw "github.com/chestorix/gophkeeper/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type Router struct {
	chi.Router
	logger *logrus.Logger
}

func NewRouter(logger *logrus.Logger) *Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &Router{
		Router: r,
		logger: logger,
	}
}

func (r *Router) SetupRoutes(handler *Handler) {
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.Register)
		r.Post("/api/user/login", handler.Login)
		r.Get("/health", handler.Health)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth(handler.service))
		r.Post("/api/data", handler.SaveData)
		r.Get("/api/data", handler.ListData)
		r.Get("/api/data/{id}", handler.GetData)
		r.Get("/api/data/name/{name}", handler.GetDataByName) // НОВЫЙ ЭНДПОИНТ
		r.Put("/api/data/{id}", handler.UpdateData)
		r.Delete("/api/data/{id}", handler.DeleteData)
		r.Delete("/api/data/name/{name}", handler.DeleteDataByName) // НОВЫЙ ЭНДПОИНТ
		r.Post("/api/sync", handler.SyncData)
	})
}
