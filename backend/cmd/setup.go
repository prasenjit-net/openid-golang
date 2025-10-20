package cmd

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

const (
	yesAnswer = "yes"
)

var (
	demoMode bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the OpenID Connect server",
	Long:  `Interactive setup wizard to configure the OpenID Connect server, generate keys, and create initial user.`,
	Run:   runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&demoMode, "demo", false, "Create demo user and client for testing")
}

func runSetup(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 OpenID Connect Server Setup")
	fmt.Println("================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Setup configuration
	if err := setupConfiguration(reader); err != nil {
		fmt.Printf("❌ Configuration setup failed: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Generate JWT keys
	if err := generateKeys(reader); err != nil {
		fmt.Printf("❌ Key generation failed: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Initialize database
	if demoMode {
		fmt.Println("\n📦 Demo Mode: Creating demo data...")
		if err := initializeDemoData(); err != nil {
			fmt.Printf("❌ Demo data creation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := initializeDatabase(reader); err != nil {
			fmt.Printf("❌ Database initialization failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("\n✅ Setup completed successfully!")
	fmt.Println("\nYou can now start the server with:")
	fmt.Println("  ./openid-server serve")
	fmt.Println("or simply:")
	fmt.Println("  ./openid-server")
}

func setupConfiguration(reader *bufio.Reader) error {
	configPath := "config/config.toml"

	// Ensure config directory exists
	if err := os.MkdirAll("config", 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config.toml exists
	if fileExists(configPath) {
		fmt.Printf("Configuration file already exists at %s\n", configPath)
		fmt.Print("Do you want to reconfigure? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != yesAnswer {
			fmt.Println("✓ Using existing configuration")
			return nil
		}
	}

	// Get configuration values
	host := promptWithDefault(reader, "Server host", "0.0.0.0")
	port := promptWithDefault(reader, "Server port", "8080")

	storageType := promptWithDefault(reader, "Storage type (json/mongodb)", config.StorageTypeJSON)
	storageType = strings.ToLower(storageType)

	jsonFilePath, mongoURI := promptStorageConfig(reader, storageType)

	// Determine issuer URL
	issuer := fmt.Sprintf("http://%s:%s", host, port)
	if host == "0.0.0.0" {
		issuer = fmt.Sprintf("http://localhost:%s", port)
	}
	issuer = promptWithDefault(reader, "Issuer URL", issuer)

	// Build and write configuration
	configContent := buildConfigContent(host, port, storageType, jsonFilePath, mongoURI, issuer)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.toml file: %w", err)
	}

	fmt.Printf("✓ Configuration saved to %s\n", configPath)
	return nil
}

func generateKeys(reader *bufio.Reader) error {
	keyDir := "config/keys"

	// Check if keys already exist
	privateKeyPath := filepath.Join(keyDir, "private.key")
	publicKeyPath := filepath.Join(keyDir, "public.key")

	if fileExists(privateKeyPath) && fileExists(publicKeyPath) {
		fmt.Println("JWT keys already exist")
		fmt.Print("Do you want to regenerate them? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != yesAnswer {
			fmt.Println("✓ Using existing keys")
			return nil
		}
	}

	// Create keys directory
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Generate RSA key pair
	fmt.Println("Generating RSA key pair (4096 bits)...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Save private key
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if encodeErr := pem.Encode(privateKeyFile, privateKeyPEM); encodeErr != nil {
		return fmt.Errorf("failed to write private key: %w", encodeErr)
	}

	// Save public key
	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if encodeErr := pem.Encode(publicKeyFile, publicKeyPEM); encodeErr != nil {
		return fmt.Errorf("failed to write public key: %w", encodeErr)
	}

	fmt.Printf("✓ Keys generated successfully\n")
	fmt.Printf("  Private key: %s\n", privateKeyPath)
	fmt.Printf("  Public key: %s\n", publicKeyPath)

	return nil
}

func initializeDatabase(reader *bufio.Reader) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Create initial admin user
	fmt.Println("\n👤 Create Initial Admin User")
	fmt.Println("----------------------------")

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Full Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleAdmin,
	}

	if createErr := store.CreateUser(user); createErr != nil {
		return fmt.Errorf("failed to create user: %w", createErr)
	}

	fmt.Printf("✓ Admin user '%s' created successfully\n", username)

	// Create admin-ui client
	adminClient := models.NewAdminUIClient(cfg.Issuer)
	if createErr := store.CreateClient(adminClient); createErr != nil {
		return fmt.Errorf("failed to create admin-ui client: %w", createErr)
	}

	fmt.Println("✓ Admin UI client created")

	return nil
}

func initializeDemoData() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Create demo admin user
	demoPassword := "demo123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := &models.User{
		ID:           uuid.New().String(),
		Username:     "admin",
		Email:        "admin@example.com",
		Name:         "Demo Admin",
		PasswordHash: string(hashedPassword),
		Role:         models.RoleAdmin,
	}

	if createErr := store.CreateUser(adminUser); createErr != nil {
		return fmt.Errorf("failed to create admin user: %w", createErr)
	}

	fmt.Printf("✓ Demo admin user created\n")
	fmt.Printf("  Username: admin\n")
	fmt.Printf("  Password: %s\n", demoPassword)
	fmt.Printf("  Email: admin@example.com\n")

	// Create demo regular user
	userPassword := "user123"
	hashedUserPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	regularUser := &models.User{
		ID:           uuid.New().String(),
		Username:     "user",
		Email:        "user@example.com",
		Name:         "Demo User",
		PasswordHash: string(hashedUserPassword),
		Role:         models.RoleUser,
	}

	if createErr := store.CreateUser(regularUser); createErr != nil {
		return fmt.Errorf("failed to create regular user: %w", createErr)
	}

	fmt.Printf("✓ Demo regular user created\n")
	fmt.Printf("  Username: user\n")
	fmt.Printf("  Password: %s\n", userPassword)
	fmt.Printf("  Email: user@example.com\n")

	// Create admin-ui client
	adminClient := models.NewAdminUIClient(cfg.Issuer)
	if createErr := store.CreateClient(adminClient); createErr != nil {
		return fmt.Errorf("failed to create admin-ui client: %w", createErr)
	}

	fmt.Println("✓ Admin UI client created")

	// Create demo test client
	clientSecret, err := crypto.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate client secret: %w", err)
	}
	testClient := &models.Client{
		ID:            "demo-client",
		Secret:        clientSecret,
		Name:          "Demo Test Client",
		RedirectURIs:  []string{"http://localhost:3000/callback"},
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"code"},
		Scope:         "openid profile email",
	}

	if createErr := store.CreateClient(testClient); createErr != nil {
		return fmt.Errorf("failed to create test client: %w", createErr)
	}

	fmt.Printf("✓ Demo test client created\n")
	fmt.Printf("  Client ID: demo-client\n")
	fmt.Printf("  Client Secret: %s\n", clientSecret)
	fmt.Printf("  Redirect URI: http://localhost:3000/callback\n")

	return nil
}

// Helper functions

func promptWithDefault(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s (default: %s): ", prompt, defaultValue)
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func promptStorageConfig(reader *bufio.Reader, storageType string) (jsonFilePath, mongoURI string) {
	if storageType == config.StorageTypeJSON {
		jsonFilePath = promptWithDefault(reader, "JSON file path", config.DefaultJSONFile)
	} else if storageType == config.StorageTypeMongoDB {
		mongoURI = promptWithDefault(reader, "MongoDB connection URI", "mongodb://localhost:27017/openid")
	}
	return jsonFilePath, mongoURI
}

func buildConfigContent(host, port, storageType, jsonFilePath, mongoURI, issuer string) string {
	configContent := fmt.Sprintf(`# OpenID Connect Server Configuration
issuer = "%s"

[server]
host = "%s"
port = %s

[storage]
type = "%s"
`, issuer, host, port, storageType)

	if storageType == config.StorageTypeJSON {
		configContent += fmt.Sprintf(`json_file_path = "%s"
`, jsonFilePath)
	} else if storageType == config.StorageTypeMongoDB {
		configContent += fmt.Sprintf(`mongo_uri = "%s"
`, mongoURI)
	}

	configContent += `
[jwt]
private_key_path = "config/keys/private.key"
public_key_path = "config/keys/public.key"
expiry_minutes = 60
`
	return configContent
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
