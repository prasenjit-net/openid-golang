package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/storage"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	config     *config.Config
	storage    storage.Storage
	jwtManager *crypto.JWTManager
}

// NewHandlers creates a new handlers instance
func NewHandlers(store storage.Storage, jwtManager *crypto.JWTManager, cfg *config.Config) *Handlers {
	return &Handlers{
		config:     cfg,
		storage:    store,
		jwtManager: jwtManager,
	}
}

// GetStorage returns the storage instance
func (h *Handlers) GetStorage() storage.Storage {
	return h.storage
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, err, description string) {
	writeJSON(w, status, map[string]string{
		"error":             err,
		"error_description": description,
	})
}
