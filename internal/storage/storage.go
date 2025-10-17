package storage

import (
	"fmt"

	"github.com/prasenjit/openid-golang/internal/config"
	"github.com/prasenjit/openid-golang/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	// User operations
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)

	// Client operations
	CreateClient(client *models.Client) error
	GetClientByID(id string) (*models.Client, error)
	ValidateClient(clientID, clientSecret string) (*models.Client, error)

	// Authorization code operations
	CreateAuthorizationCode(code *models.AuthorizationCode) error
	GetAuthorizationCode(code string) (*models.AuthorizationCode, error)
	DeleteAuthorizationCode(code string) error

	// Token operations
	CreateToken(token *models.Token) error
	GetTokenByAccessToken(accessToken string) (*models.Token, error)
	GetTokenByRefreshToken(refreshToken string) (*models.Token, error)
	DeleteToken(accessToken string) error

	// Session operations
	CreateSession(session *models.Session) error
	GetSession(id string) (*models.Session, error)
	DeleteSession(id string) error

	Close() error
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Database.Type {
	case "sqlite":
		return NewSQLiteStorage(cfg.Database.Connection)
	case "postgres":
		return nil, fmt.Errorf("postgres storage not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Database.Type)
	}
}
