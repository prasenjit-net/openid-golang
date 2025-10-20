package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

const (
	// StorageTypeJSON represents JSON file storage type
	StorageTypeJSON = "json"
	// StorageTypeMongoDB represents MongoDB storage type
	StorageTypeMongoDB = "mongodb"
	// DefaultJSONFile is the default JSON storage file
	DefaultJSONFile = "data/openid.json"
)

// Config holds the application configuration
type Config struct {
	Server  ServerConfig  `toml:"server"`
	Storage StorageConfig `toml:"storage"`
	JWT     JWTConfig     `toml:"jwt"`
	Issuer  string        `toml:"issuer"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type         string `toml:"type"`           // mongodb or json
	MongoURI     string `toml:"mongo_uri"`      // MongoDB connection URI
	JSONFilePath string `toml:"json_file_path"` // Path to JSON file for json storage
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	PrivateKeyPath string `toml:"private_key_path"`
	PublicKeyPath  string `toml:"public_key_path"`
	ExpiryMinutes  int    `toml:"expiry_minutes"`
}

// Load loads configuration from config.toml or environment variables
func Load() (*Config, error) {
	// Try to load from config.toml first
	if _, err := os.Stat("config.toml"); err == nil {
		return LoadFromTOML("config.toml")
	}

	// Fallback to environment variables
	return loadFromEnv()
}

// LoadFromTOML loads configuration from a TOML file
func LoadFromTOML(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for any missing values
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = StorageTypeJSON
	}
	if cfg.Storage.JSONFilePath == "" {
		cfg.Storage.JSONFilePath = DefaultJSONFile
	}
	if cfg.JWT.ExpiryMinutes == 0 {
		cfg.JWT.ExpiryMinutes = 60
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadFromEnv loads configuration from environment variables (backward compatibility)
func loadFromEnv() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Storage: StorageConfig{
			Type:         getEnv("STORAGE_TYPE", StorageTypeJSON),
			MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017/openid"),
			JSONFilePath: getEnv("JSON_FILE_PATH", DefaultJSONFile),
		},
		JWT: JWTConfig{
			PrivateKeyPath: getEnv("JWT_PRIVATE_KEY", "config/keys/private.key"),
			PublicKeyPath:  getEnv("JWT_PUBLIC_KEY", "config/keys/public.key"),
			ExpiryMinutes:  getEnvAsInt("JWT_EXPIRY_MINUTES", 60),
		},
		Issuer: getEnv("ISSUER", "http://localhost:8080"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	if c.Issuer == "" {
		return fmt.Errorf("issuer cannot be empty")
	}

	if c.Storage.Type != StorageTypeMongoDB && c.Storage.Type != StorageTypeJSON {
		return fmt.Errorf("unsupported storage type: %s (must be 'mongodb' or 'json')", c.Storage.Type)
	}

	if c.Storage.Type == StorageTypeMongoDB && c.Storage.MongoURI == "" {
		return fmt.Errorf("mongo_uri is required when storage type is 'mongodb'")
	}

	if c.Storage.Type == StorageTypeJSON && c.Storage.JSONFilePath == "" {
		return fmt.Errorf("json_file_path is required when storage type is 'json'")
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// SaveToTOML saves configuration to a TOML file
func (c *Config) SaveToTOML(filePath string) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
