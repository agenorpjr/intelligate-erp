package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/intelligate/erp/backend/internal/config"
)

type HealthResponse struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

func HealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:      "ok",
			Version:     cfg.AppVersion,
			Environment: cfg.AppEnv,
			Timestamp:   time.Now().UTC(),
		}
		RespondJSON(w, http.StatusOK, response)
	}
}

func RegisterHealthRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/health", HealthHandler(cfg))
}
