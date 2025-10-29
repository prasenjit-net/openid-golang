package storage

import (
	"strings"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/models"
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
	UpdateAuthorizationCode(code *models.AuthorizationCode) error
	DeleteAuthorizationCode(code string) error

	// Token operations
	CreateToken(token *models.Token) error
	GetTokenByAccessToken(accessToken string) (*models.Token, error)
	GetTokenByRefreshToken(refreshToken string) (*models.Token, error)
	GetTokensByAuthCode(authCodeID string) ([]*models.Token, error)
	DeleteToken(accessToken string) error
	RevokeTokensByAuthCode(authCodeID string) error

	// Session operations
	CreateSession(session *models.Session) error
	GetSession(id string) (*models.Session, error)
	DeleteSession(id string) error

	// AuthSession operations (OpenID Connect authorization sessions)
	CreateAuthSession(session *models.AuthSession) error
	GetAuthSession(id string) (*models.AuthSession, error)
	UpdateAuthSession(session *models.AuthSession) error
	DeleteAuthSession(id string) error

	// UserSession operations (authenticated user sessions)
	CreateUserSession(session *models.UserSession) error
	GetUserSession(id string) (*models.UserSession, error)
	GetUserSessionByUserID(userID string) (*models.UserSession, error)
	UpdateUserSession(session *models.UserSession) error
	DeleteUserSession(id string) error
	CleanupExpiredSessions() error

	// Consent operations (user consent tracking)
	CreateConsent(consent *models.Consent) error
	GetConsent(userID, clientID string) (*models.Consent, error)
	UpdateConsent(consent *models.Consent) error
	DeleteConsent(userID, clientID string) error
	DeleteConsentsForUser(userID string) error

	// InitialAccessToken operations (for dynamic client registration)
	CreateInitialAccessToken(token *models.InitialAccessToken) error
	GetInitialAccessToken(token string) (*models.InitialAccessToken, error)
	UpdateInitialAccessToken(token *models.InitialAccessToken) error
	DeleteInitialAccessToken(token string) error
	GetAllInitialAccessTokens() ([]*models.InitialAccessToken, error)

	// SigningKey operations (for key rotation)
	CreateSigningKey(key *models.SigningKey) error
	GetSigningKey(id string) (*models.SigningKey, error)
	GetSigningKeyByKID(kid string) (*models.SigningKey, error)
	GetAllSigningKeys() ([]*models.SigningKey, error)
	GetActiveSigningKey() (*models.SigningKey, error)
	UpdateSigningKey(key *models.SigningKey) error
	DeleteSigningKey(id string) error

	// Statistics operations
	GetActiveTokensCount() int
	GetRecentUserSessionsCount() int

	Close() error
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(cfg *configstore.ConfigData) (Storage, error) {
	switch cfg.Storage.Type {
	case "mongodb":
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
	case "json":
		return NewJSONStorage(cfg.Storage.JSONFilePath)
	default:
		// Default to JSON storage for backward compatibility
		return NewJSONStorage("data.json")
	}
}
