package configstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONConfigStore implements ConfigStore using a JSON file
type JSONConfigStore struct {
	filePath string
	mu       sync.RWMutex
	config   *ConfigData
}

// NewJSONConfigStore creates a new JSON-based config store
func NewJSONConfigStore(filePath string) *JSONConfigStore {
	return &JSONConfigStore{
		filePath: filePath,
	}
}

// Initialize checks if config file exists and creates directory if needed
func (s *JSONConfigStore) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// IsInitialized returns true if the config file exists and is valid
func (s *JSONConfigStore) IsInitialized(ctx context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check config file: %w", err)
	}

	// Try to read and parse the file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigData
	if err := json.Unmarshal(data, &config); err != nil {
		return false, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Basic validation
	if config.Issuer == "" || config.JWT.PrivateKey == "" {
		return false, nil
	}

	return true, nil
}

// GetConfig retrieves the current configuration
func (s *JSONConfigStore) GetConfig(ctx context.Context) (*ConfigData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", s.filePath)
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigData
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Cache the config
	s.config = &config

	return &config, nil
}

// SaveConfig saves the configuration to file
func (s *JSONConfigStore) SaveConfig(ctx context.Context, config *ConfigData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// Update cache
	s.config = config

	return nil
}

// UpdateConfig updates specific fields in the configuration
func (s *JSONConfigStore) UpdateConfig(ctx context.Context, updates map[string]interface{}) error {
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
		// Add more fields as needed
		default:
			return fmt.Errorf("unknown config field: %s", key)
		}
	}

	// Save updated config
	return s.SaveConfig(ctx, config)
}

// Close closes the config store (no-op for JSON)
func (s *JSONConfigStore) Close() error {
	return nil
}
