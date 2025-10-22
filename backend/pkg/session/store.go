package session

import (
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

// Store provides session management operations
type Store interface {
	// AuthSession operations
	CreateAuthSession(session *models.AuthSession) error
	GetAuthSession(id string) (*models.AuthSession, error)
	UpdateAuthSession(session *models.AuthSession) error
	DeleteAuthSession(id string) error

	// UserSession operations
	CreateUserSession(session *models.UserSession) error
	GetUserSession(id string) (*models.UserSession, error)
	UpdateUserSession(session *models.UserSession) error
	DeleteUserSession(id string) error
	DeleteExpiredUserSessions() error

	// Convenience methods
	CleanupExpiredSessions() error
}

// sessionStore implements Store using the existing storage backend
type sessionStore struct {
	storage storage.Storage
}

// NewStore creates a new session store
func NewStore(storage storage.Storage) Store {
	return &sessionStore{
		storage: storage,
	}
}

// CreateAuthSession creates a new authorization session
func (s *sessionStore) CreateAuthSession(session *models.AuthSession) error {
	return s.storage.CreateAuthSession(session)
}

// GetAuthSession retrieves an authorization session by ID
func (s *sessionStore) GetAuthSession(id string) (*models.AuthSession, error) {
	return s.storage.GetAuthSession(id)
}

// UpdateAuthSession updates an authorization session
func (s *sessionStore) UpdateAuthSession(session *models.AuthSession) error {
	return s.storage.UpdateAuthSession(session)
}

// DeleteAuthSession deletes an authorization session
func (s *sessionStore) DeleteAuthSession(id string) error {
	return s.storage.DeleteAuthSession(id)
}

// CreateUserSession creates a new user session
func (s *sessionStore) CreateUserSession(session *models.UserSession) error {
	return s.storage.CreateUserSession(session)
}

// GetUserSession retrieves a user session by ID
func (s *sessionStore) GetUserSession(id string) (*models.UserSession, error) {
	return s.storage.GetUserSession(id)
}

// UpdateUserSession updates a user session
func (s *sessionStore) UpdateUserSession(session *models.UserSession) error {
	return s.storage.UpdateUserSession(session)
}

// DeleteUserSession deletes a user session
func (s *sessionStore) DeleteUserSession(id string) error {
	return s.storage.DeleteUserSession(id)
}

// DeleteExpiredUserSessions removes all expired user sessions
func (s *sessionStore) DeleteExpiredUserSessions() error {
	return s.storage.CleanupExpiredSessions()
}

// CleanupExpiredSessions removes all expired sessions
func (s *sessionStore) CleanupExpiredSessions() error {
	return s.storage.CleanupExpiredSessions()
}
