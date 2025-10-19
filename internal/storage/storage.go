package storage

import (
	"strings"

	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/models"
)

const (
	// StorageTypeJSON represents JSON file storage type
	StorageTypeJSON = "json"
	// StorageTypeMongoDB represents MongoDB storage type
	StorageTypeMongoDB = "mongodb"
	// DefaultJSONFile is the default JSON storage file
	DefaultJSONFile = "data.json"
)

// Storage defines the interface for data persistence
type Storage interface {
	// User operations
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error

	// Client operations
	CreateClient(client *models.Client) error
	GetClientByID(id string) (*models.Client, error)
	GetAllClients() ([]*models.Client, error)
	UpdateClient(client *models.Client) error
	DeleteClient(id string) error
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
	switch cfg.Storage.Type {
	case StorageTypeMongoDB:
		// Parse database name from MongoDB URI
		// Expected format: mongodb://host:port/database
		uri := cfg.Storage.MongoURI
		dbName := "openid" // default
		if idx := strings.LastIndex(uri, "/"); idx != -1 && idx < len(uri)-1 {
			dbName = uri[idx+1:]
			// Remove query parameters if present
			if qIdx := strings.Index(dbName, "?"); qIdx != -1 {
				dbName = dbName[:qIdx]
			}
		}
		return NewMongoDBStorage(uri, dbName)
	case StorageTypeJSON:
		return NewJSONStorage(cfg.Storage.JSONFilePath)
	default:
		// Default to JSON storage for backward compatibility
		return NewJSONStorage(DefaultJSONFile)
	}
}
