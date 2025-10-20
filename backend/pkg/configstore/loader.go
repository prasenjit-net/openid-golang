package configstore

import (
	"context"
	"fmt"
	"os"
)

// LoaderConfig holds configuration for the config loader
type LoaderConfig struct {
	// MongoDB environment variables
	MongoURIEnv      string
	MongoDatabaseEnv string

	// JSON file path
	JSONFilePath string

	// Default values
	DefaultIssuer string
	DefaultHost   string
	DefaultPort   int
}

// DefaultLoaderConfig returns default loader configuration
func DefaultLoaderConfig() LoaderConfig {
	return LoaderConfig{
		MongoURIEnv:      "MONGODB_URI",
		MongoDatabaseEnv: "MONGODB_DATABASE",
		JSONFilePath:     "data/config.json",
		DefaultHost:      "0.0.0.0",
		DefaultPort:      8080,
	}
}

// AutoLoadConfigStore automatically detects and loads the appropriate config store
// Priority: 1) MongoDB (from env), 2) JSON file, 3) New JSON file
func AutoLoadConfigStore(ctx context.Context, cfg LoaderConfig) (ConfigStore, bool, error) {
	// Step 1: Check for MongoDB configuration in environment
	mongoURI := os.Getenv(cfg.MongoURIEnv)
	if mongoURI != "" {
		mongoDatabase := os.Getenv(cfg.MongoDatabaseEnv)
		if mongoDatabase == "" {
			mongoDatabase = "openid"
		}

		store, err := NewMongoConfigStore(mongoURI, mongoDatabase)
		if err != nil {
			return nil, false, fmt.Errorf("failed to connect to MongoDB: %w", err)
		}

		// Check if initialized
		initialized, err := store.IsInitialized(ctx)
		if err != nil {
			store.Close()
			return nil, false, fmt.Errorf("failed to check MongoDB initialization: %w", err)
		}

		return store, initialized, nil
	}

	// Step 2: Check for JSON config file
	jsonStore := NewJSONConfigStore(cfg.JSONFilePath)

	// Initialize directory if needed
	if err := jsonStore.Initialize(ctx); err != nil {
		return nil, false, fmt.Errorf("failed to initialize JSON store: %w", err)
	}

	// Check if initialized
	initialized, err := jsonStore.IsInitialized(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check JSON initialization: %w", err)
	}

	return jsonStore, initialized, nil
}

// InitializeMinimalConfig initializes the config store with minimal required data
func InitializeMinimalConfig(ctx context.Context, store ConfigStore, issuer string) error {
	// Generate JWT keys
	privateKey, publicKey, err := GenerateJWTKeyPair(4096)
	if err != nil {
		return fmt.Errorf("failed to generate JWT keys: %w", err)
	}

	// Create minimal config
	config := DefaultConfig()
	config.Issuer = issuer
	config.JWT.PrivateKey = privateKey
	config.JWT.PublicKey = publicKey

	// Detect storage backend based on config store type
	// If using MongoDB config store, also use MongoDB for data storage
	if mongoStore, ok := store.(*MongoConfigStore); ok {
		config.Storage.Type = "mongodb"
		config.Storage.MongoURI = mongoStore.mongoURI
		config.Storage.MongoDatabase = mongoStore.database
		// Clear JSON file path when using MongoDB
		config.Storage.JSONFilePath = ""
	}
	// Otherwise, DefaultConfig already sets it to JSON

	// Save to store
	if err := store.SaveConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetOrCreateConfig loads config or creates default if not initialized
func GetOrCreateConfig(ctx context.Context, store ConfigStore, issuer string) (*ConfigData, error) {
	// Check if initialized
	initialized, err := store.IsInitialized(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check initialization: %w", err)
	}

	if !initialized {
		// Initialize with minimal config
		if initErr := InitializeMinimalConfig(ctx, store, issuer); initErr != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", initErr)
		}
	}

	// Load config
	config, err := store.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return config, nil
}
