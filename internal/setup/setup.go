package setup

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

	"golang.org/x/crypto/bcrypt"

	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/models"
	"github.com/prasenjit-net/openid-golang/internal/storage"
)

// Run executes the interactive setup wizard
func Run() error {
	fmt.Println("🚀 OpenID Connect Server Setup Wizard")
	fmt.Println("=====================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Generate RSA keys
	fmt.Println("Step 1: Generate RSA Keys")
	fmt.Println("-------------------------")
	if err := setupKeys(reader); err != nil {
		return fmt.Errorf("failed to setup keys: %w", err)
	}

	// Step 2: Create configuration
	fmt.Println("\nStep 2: Server Configuration")
	fmt.Println("-----------------------------")
	if err := setupConfig(reader); err != nil {
		return fmt.Errorf("failed to setup config: %w", err)
	}

	// Step 3: Initialize database
	fmt.Println("\nStep 3: Initialize Database")
	fmt.Println("----------------------------")
	if err := initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Step 4: Create admin user
	fmt.Println("\nStep 4: Create Admin User")
	fmt.Println("-------------------------")
	if err := createAdminUser(reader); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Step 5: Create first OAuth client
	fmt.Println("\nStep 5: Create OAuth Client (Optional)")
	fmt.Println("---------------------------------------")
	if err := createOAuthClient(reader); err != nil {
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}

	fmt.Println("\n✅ Setup Complete!")
	fmt.Println("\nYou can now start the server with:")
	fmt.Println("  ./openid-server")
	fmt.Println("\nAccess the admin UI at:")
	fmt.Println("  http://localhost:8080/")
	fmt.Println()

	return nil
}

func setupKeys(reader *bufio.Reader) error {
	keyDir := "config/keys"
	privateKeyPath := filepath.Join(keyDir, "private.key")
	publicKeyPath := filepath.Join(keyDir, "public.key")

	// Check if keys already exist
	if fileExists(privateKeyPath) && fileExists(publicKeyPath) {
		fmt.Printf("RSA keys already exist at %s\n", keyDir)
		fmt.Print("Do you want to regenerate them? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("✓ Using existing keys")
			return nil
		}
	}

	// Create directory
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Generate RSA key pair
	fmt.Println("Generating 4096-bit RSA key pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Save private key
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Save public key
	publicKeyFile, err := os.Create(publicKeyPath)
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
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	fmt.Printf("✓ RSA keys generated at %s\n", keyDir)
	return nil
}

func setupConfig(reader *bufio.Reader) error {
	envPath := ".env"

	// Check if .env exists
	if fileExists(envPath) {
		fmt.Printf("Configuration file already exists at %s\n", envPath)
		fmt.Print("Do you want to reconfigure? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("✓ Using existing configuration")
			return nil
		}
	}

	// Get server host
	fmt.Print("Server host (default: 0.0.0.0): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "0.0.0.0"
	}

	// Get server port
	fmt.Print("Server port (default: 8080): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "8080"
	}

	// Get database type
	fmt.Print("Database type (sqlite/postgres, default: sqlite): ")
	dbType, _ := reader.ReadString('\n')
	dbType = strings.TrimSpace(strings.ToLower(dbType))
	if dbType == "" {
		dbType = "sqlite"
	}

	// Get database connection
	var dbConnection string
	if dbType == "sqlite" {
		fmt.Print("Database file path (default: ./openid.db): ")
		dbConnection, _ = reader.ReadString('\n')
		dbConnection = strings.TrimSpace(dbConnection)
		if dbConnection == "" {
			dbConnection = "./openid.db"
		}
	} else {
		fmt.Print("Database connection string: ")
		dbConnection, _ = reader.ReadString('\n')
		dbConnection = strings.TrimSpace(dbConnection)
	}

	// Get issuer URL
	issuer := fmt.Sprintf("http://%s:%s", host, port)
	if host == "0.0.0.0" {
		issuer = fmt.Sprintf("http://localhost:%s", port)
	}
	fmt.Printf("Issuer URL (default: %s): ", issuer)
	customIssuer, _ := reader.ReadString('\n')
	customIssuer = strings.TrimSpace(customIssuer)
	if customIssuer != "" {
		issuer = customIssuer
	}

	// Create .env file
	envContent := fmt.Sprintf(`# OpenID Connect Server Configuration
SERVER_HOST=%s
SERVER_PORT=%s
DB_TYPE=%s
DB_CONNECTION=%s
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
JWT_EXPIRY_MINUTES=60
ISSUER=%s
`, host, port, dbType, dbConnection, issuer)

	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	fmt.Printf("✓ Configuration saved to %s\n", envPath)
	return nil
}

func initializeDatabase() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	fmt.Println("✓ Database initialized successfully")
	return nil
}

func createAdminUser(reader *bufio.Reader) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Get username
	fmt.Print("Admin username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Get email
	fmt.Print("Admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Get password
	fmt.Print("Admin password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := store.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("✓ Admin user '%s' created successfully\n", username)
	return nil
}

func createOAuthClient(reader *bufio.Reader) error {
	fmt.Print("Do you want to create an OAuth client now? (y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("✓ Skipped OAuth client creation")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Get client name
	fmt.Print("Client name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("client name cannot be empty")
	}

	// Get redirect URI
	fmt.Print("Redirect URI: ")
	redirectURI, _ := reader.ReadString('\n')
	redirectURI = strings.TrimSpace(redirectURI)
	if redirectURI == "" {
		return fmt.Errorf("redirect URI cannot be empty")
	}

	// Generate client credentials
	clientID := generateRandomString(32)
	clientSecret := generateRandomString(64)

	// Create client
	client := &models.Client{
		ID:            clientID,
		Secret:        clientSecret,
		Name:          name,
		RedirectURIs:  []string{redirectURI},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
	}

	if err := store.CreateClient(client); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	fmt.Println("\n✓ OAuth client created successfully!")
	fmt.Printf("\nClient ID:     %s\n", clientID)
	fmt.Printf("Client Secret: %s\n", clientSecret)
	fmt.Println("\n⚠️  Please save these credentials securely - the secret won't be shown again!")
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
