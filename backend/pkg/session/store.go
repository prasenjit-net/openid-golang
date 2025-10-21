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
	// Store as a generic session in the storage backend
	genericSession := &models.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
	return s.storage.CreateSession(genericSession)
}

// GetAuthSession retrieves an authorization session by ID
func (s *sessionStore) GetAuthSession(id string) (*models.AuthSession, error) {
	session, err := s.storage.GetSession(id)
	if err != nil {
		return nil, err
	}

	// For now, return a basic mapping
	// TODO: Extend storage to support full AuthSession
	authSession := &models.AuthSession{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
	return authSession, nil
}

// UpdateAuthSession updates an authorization session
func (s *sessionStore) UpdateAuthSession(session *models.AuthSession) error {
	// For now, use create/update logic
	return s.CreateAuthSession(session)
}

// DeleteAuthSession deletes an authorization session
func (s *sessionStore) DeleteAuthSession(id string) error {
	return s.storage.DeleteSession(id)
}

// CreateUserSession creates a new user session
func (s *sessionStore) CreateUserSession(session *models.UserSession) error {
	genericSession := &models.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
	return s.storage.CreateSession(genericSession)
}

// GetUserSession retrieves a user session by ID
func (s *sessionStore) GetUserSession(id string) (*models.UserSession, error) {
	session, err := s.storage.GetSession(id)
	if err != nil {
		return nil, err
	}

	userSession := &models.UserSession{
		ID:             session.ID,
		UserID:         session.UserID,
		AuthTime:       session.CreatedAt,
		LastActivityAt: session.CreatedAt,
		ExpiresAt:      session.ExpiresAt,
		CreatedAt:      session.CreatedAt,
	}
	return userSession, nil
}

// UpdateUserSession updates a user session
func (s *sessionStore) UpdateUserSession(session *models.UserSession) error {
	return s.CreateUserSession(session)
}

// DeleteUserSession deletes a user session
func (s *sessionStore) DeleteUserSession(id string) error {
	return s.storage.DeleteSession(id)
}

// DeleteExpiredUserSessions removes all expired user sessions
func (s *sessionStore) DeleteExpiredUserSessions() error {
	// This requires iterating through sessions
	// For now, return nil (individual checks on retrieval)
	return nil
}

// CleanupExpiredSessions removes all expired sessions
func (s *sessionStore) CleanupExpiredSessions() error {
	return s.DeleteExpiredUserSessions()
}
