package handlers

import (
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
