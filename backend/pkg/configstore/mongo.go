package configstore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConfigStore implements ConfigStore using MongoDB
type MongoConfigStore struct {
	client     *mongo.Client
	database   string
	collection string
	mu         sync.RWMutex
	config     *ConfigData
}

// NewMongoConfigStore creates a new MongoDB-based config store
func NewMongoConfigStore(mongoURI, database string) (*MongoConfigStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoConfigStore{
		client:     client,
		database:   database,
		collection: "config",
	}, nil
}

// Initialize ensures the config collection exists
func (s *MongoConfigStore) Initialize(ctx context.Context) error {
	// MongoDB creates collections automatically, nothing to do
	return nil
}

// IsInitialized returns true if a config document exists
func (s *MongoConfigStore) IsInitialized(ctx context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coll := s.client.Database(s.database).Collection(s.collection)

	// Check if any config document exists
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, fmt.Errorf("failed to count config documents: %w", err)
	}

	if count == 0 {
		return false, nil
	}

	// Try to get the config to validate it
	var config ConfigData
	err = coll.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return false, fmt.Errorf("failed to read config: %w", err)
	}

	// Basic validation
	if config.Issuer == "" || config.JWT.PrivateKey == "" {
		return false, nil
	}

	return true, nil
}

// GetConfig retrieves the current configuration
func (s *MongoConfigStore) GetConfig(ctx context.Context) (*ConfigData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coll := s.client.Database(s.database).Collection(s.collection)

	var config ConfigData
	err := coll.FindOne(ctx, bson.M{}).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("config not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Cache the config
	s.config = &config

	return &config, nil
}

// SaveConfig saves the configuration to MongoDB
func (s *MongoConfigStore) SaveConfig(ctx context.Context, config *ConfigData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update timestamp
	config.UpdatedAt = time.Now()

	coll := s.client.Database(s.database).Collection(s.collection)

	// Delete existing config (there should only be one)
	_, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete old config: %w", err)
	}

	// Insert new config
	_, err = coll.InsertOne(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update cache
	s.config = config

	return nil
}

// UpdateConfig updates specific fields in the configuration
func (s *MongoConfigStore) UpdateConfig(ctx context.Context, updates map[string]interface{}) error {
	// Get current config
	config, err := s.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "issuer":
			if v, ok := value.(string); ok {
				config.Issuer = v
			}
		case "server.host":
			if v, ok := value.(string); ok {
				config.Server.Host = v
			}
		case "server.port":
			if v, ok := value.(float64); ok {
				config.Server.Port = int(v)
			} else if v, ok := value.(int); ok {
				config.Server.Port = v
			}
		case "jwt.expiry_minutes":
			if v, ok := value.(float64); ok {
				config.JWT.ExpiryMinutes = int(v)
			} else if v, ok := value.(int); ok {
				config.JWT.ExpiryMinutes = v
			}
		case "jwt.refresh_enabled":
			if v, ok := value.(bool); ok {
				config.JWT.RefreshEnabled = v
			}
		default:
			return fmt.Errorf("unknown config field: %s", key)
		}
	}

	// Save updated config
	return s.SaveConfig(ctx, config)
}

// Close closes the MongoDB connection
func (s *MongoConfigStore) Close() error {
	if s.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.client.Disconnect(ctx)
	}
	return nil
}
