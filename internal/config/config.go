package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Issuer   string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type       string // sqlite, postgres
	Connection string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Issuer         string
	ExpiryMinutes  int
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Type:       getEnv("DB_TYPE", "sqlite"),
			Connection: getEnv("DB_CONNECTION", "./openid.db"),
		},
		JWT: JWTConfig{
			PrivateKeyPath: getEnv("JWT_PRIVATE_KEY", "config/keys/private.key"),
			PublicKeyPath:  getEnv("JWT_PUBLIC_KEY", "config/keys/public.key"),
			ExpiryMinutes:  getEnvAsInt("JWT_EXPIRY_MINUTES", 60),
		},
		Issuer: getEnv("ISSUER", "http://localhost:8080"),
	}

	cfg.JWT.Issuer = cfg.Issuer

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

	if c.Database.Type != "sqlite" && c.Database.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
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
