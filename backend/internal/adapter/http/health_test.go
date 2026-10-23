package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/intelligate/erp/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	cfg := &config.Config{
		AppVersion:  "0.1.0-test",
		AppEnv:      "test",
		AppPort:     "8080",
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/intelligate?sslmode=disable",
		HTTPTimeout: 30 * time.Second,
	}

	router := chi.NewRouter()
	RegisterHealthRoutes(router, cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "0.1.0-test", response.Version)
	assert.Equal(t, "test", response.Environment)
	assert.WithinDuration(t, time.Now().UTC(), response.Timestamp, 5*time.Second)
}

func TestHealthResponseStructure(t *testing.T) {
	cfg := &config.Config{
		AppVersion:  "1.0.0",
		AppEnv:      "production",
		AppPort:     "8080",
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/intelligate?sslmode=disable",
		HTTPTimeout: 30 * time.Second,
	}

	router := chi.NewRouter()
	RegisterHealthRoutes(router, cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verify all required fields exist
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "version")
	assert.Contains(t, response, "environment")
	assert.Contains(t, response, "timestamp")

	// Verify types
	assert.IsType(t, "", response["status"])
	assert.IsType(t, "", response["version"])
	assert.IsType(t, "", response["environment"])
	assert.IsType(t, "", response["timestamp"]) // JSON serializes time as string
}
