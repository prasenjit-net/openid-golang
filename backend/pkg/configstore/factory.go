package configstore

import (
	"fmt"
)

// Config holds the initialization parameters for creating a config store
type Config struct {
	// Type of config store: "json" or "mongodb"
	Type string

	// For JSON config store
	JSONFilePath string

	// For MongoDB config store
	MongoURI      string
	MongoDatabase string
}

// NewConfigStore creates a new config store based on the provided configuration
func NewConfigStore(cfg Config) (ConfigStore, error) {
	switch cfg.Type {
	case "json", "":
		if cfg.JSONFilePath == "" {
			cfg.JSONFilePath = "data/config.json"
		}
		return NewJSONConfigStore(cfg.JSONFilePath), nil

	case "mongodb", "mongo":
		if cfg.MongoURI == "" {
			return nil, fmt.Errorf("MongoDB URI is required")
		}
		if cfg.MongoDatabase == "" {
			cfg.MongoDatabase = "openid"
		}
		return NewMongoConfigStore(cfg.MongoURI, cfg.MongoDatabase)

	default:
		return nil, fmt.Errorf("unsupported config store type: %s", cfg.Type)
	}
}

// DefaultJSONConfigStore creates a default JSON-based config store
func DefaultJSONConfigStore() ConfigStore {
	return NewJSONConfigStore("data/config.json")
}
