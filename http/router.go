package http

import (
	"net/http"
	"ohlcDataIngestionReporting/ohlc"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// @title Yield Engine API
// @version 2.0.0
// @description OHLC data ingestion API
// @host localhost:8080
// @BasePath /
// @schemes http https
func SetupRouter(repo ohlc.Repository) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger docs
	//r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Routes
	handlers := &Handlers{Repo: repo}
	r.Post("/data", handlers.UploadData)
	r.Get("/data", handlers.ListData)

	return r
}
