package handlers

import (
	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/session"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	config         *config.Config
	storage        storage.Storage
	jwtManager     *crypto.JWTManager
	sessionManager *session.Manager
}

// NewHandlers creates a new handlers instance
func NewHandlers(store storage.Storage, jwtManager *crypto.JWTManager, cfg *config.Config, sessionMgr *session.Manager) *Handlers {
	return &Handlers{
		config:         cfg,
		storage:        store,
		jwtManager:     jwtManager,
		sessionManager: sessionMgr,
	}
}

// GetStorage returns the storage instance
func (h *Handlers) GetStorage() storage.Storage {
	return h.storage
}
