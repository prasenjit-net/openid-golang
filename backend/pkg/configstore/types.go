package configstore

import (
	"context"
	"time"
)

// ConfigStore defines the interface for storing and retrieving application configuration
type ConfigStore interface {
	// Initialize checks if config store exists and initializes if needed
	Initialize(ctx context.Context) error

	// IsInitialized returns true if the config store has been set up
	IsInitialized(ctx context.Context) (bool, error)

	// GetConfig retrieves the current configuration
	GetConfig(ctx context.Context) (*ConfigData, error)

	// SaveConfig saves the configuration
	SaveConfig(ctx context.Context, config *ConfigData) error

	// UpdateConfig updates specific fields in the configuration
	UpdateConfig(ctx context.Context, updates map[string]interface{}) error

	// Close closes any open connections
	Close() error
}

// ConfigData holds all application configuration
type ConfigData struct {
	// Metadata
	Version   string    `json:"version" bson:"version"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`

	// Server Configuration
	Server ServerConfig `json:"server" bson:"server"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt" bson:"jwt"`

	// Issuer URL
	Issuer string `json:"issuer" bson:"issuer"`

	// Storage Backend Configuration
	Storage StorageBackendConfig `json:"storage" bson:"storage"`

	// Dynamic Client Registration Configuration
	Registration RegistrationConfig `json:"registration" bson:"registration"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `json:"host" bson:"host"`
	Port int    `json:"port" bson:"port"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	// Keys are stored as PEM-encoded strings
	PrivateKey     string `json:"private_key" bson:"private_key"`
	PublicKey      string `json:"public_key" bson:"public_key"`
	ExpiryMinutes  int    `json:"expiry_minutes" bson:"expiry_minutes"`
	RefreshEnabled bool   `json:"refresh_enabled" bson:"refresh_enabled"`
}

// StorageBackendConfig defines which storage backend to use for data
type StorageBackendConfig struct {
	Type string `json:"type" bson:"type"` // "json" or "mongodb"

	// For JSON backend
	JSONFilePath string `json:"json_file_path,omitempty" bson:"json_file_path,omitempty"`

	// For MongoDB backend
	MongoURI      string `json:"mongo_uri,omitempty" bson:"mongo_uri,omitempty"`
	MongoDatabase string `json:"mongo_database,omitempty" bson:"mongo_database,omitempty"`
}

// RegistrationConfig holds dynamic client registration configuration
type RegistrationConfig struct {
	Enabled                   bool   `json:"enabled" bson:"enabled"`
	Endpoint                  string `json:"endpoint" bson:"endpoint"` // Custom endpoint path (default: /register)
	ServiceDocumentation      string `json:"service_documentation,omitempty" bson:"service_documentation,omitempty"`
	PolicyURI                 string `json:"policy_uri,omitempty" bson:"policy_uri,omitempty"`
	TosURI                    string `json:"tos_uri,omitempty" bson:"tos_uri,omitempty"`
	RequireInitialAccessToken bool   `json:"require_initial_access_token" bson:"require_initial_access_token"` // Require token for registration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ConfigData {
	return &ConfigData{
		Version:   "1.0",
		UpdatedAt: time.Now(),
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		JWT: JWTConfig{
			ExpiryMinutes:  60,
			RefreshEnabled: true,
		},
		Issuer: "http://localhost:8080",
		Storage: StorageBackendConfig{
			Type:         "json",
			JSONFilePath: "data/openid.json",
		},
		Registration: RegistrationConfig{
			Enabled:  false,
			Endpoint: "/register",
		},
	}
}
