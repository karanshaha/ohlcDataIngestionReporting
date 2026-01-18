package http

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// @title Yield Engine API
// @version 2.0.0
// @description OHLC data ingestion API
// @host localhost:8080
// @BasePath /
// @schemes http https
func SetupRouter(handlers *Handlers) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "*"}, // dev + prod
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// simple logger
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Capture status after response
			statusCode := http.StatusOK
			wrapper := &statusResponseWriter{ResponseWriter: w, statusCode: &statusCode}

			next.ServeHTTP(wrapper, r)

			log.Printf("[%s] %s %s %s %d %db %.2fs batch=%d workers=%d",
				time.Now().Format("15:04:05"),
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				wrapper.statusCode,
				wrapper.size,
				time.Since(start).Seconds(),
				handlers.BatchSize,
				handlers.WorkerCount,
			)
		})
	})

	// 3. Health checks
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Routes
	r.Post("/data", handlers.UploadData)
	r.Get("/data", handlers.ListData)

	return r
}

// Helper to capture status + size (replaces GetWrapResponseWriter)
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode *int
	size       int
}

func (srw *statusResponseWriter) WriteHeader(code int) {
	*srw.statusCode = code
	srw.ResponseWriter.WriteHeader(code)
}

func (srw *statusResponseWriter) Write(b []byte) (int, error) {
	srw.size += len(b)
	return srw.ResponseWriter.Write(b)
}
