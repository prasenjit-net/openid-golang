package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

var (
	issuerURL      string
	adminUsername  string
	adminPassword  string
	storageType    string
	mongoURI       string
	nonInteractive bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the OpenID Connect server via CLI",
	Long: `CLI-based setup wizard to initialize the OpenID Connect server configuration.
This command initializes the config store with issuer URL and optionally creates an admin user.
Configuration is stored in data/config.json or MongoDB (if MONGODB_URI env is set).

Examples:
  # Interactive mode
  openid-server setup

  # Non-interactive mode with issuer
  openid-server setup --issuer http://localhost:8080

  # With admin user
  openid-server setup --issuer http://localhost:8080 --admin-user admin --admin-pass secret123

  # Using environment variables
  ISSUER_URL=http://localhost:8080 ADMIN_USER=admin ADMIN_PASS=secret123 openid-server setup --non-interactive
`,
	Run: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().StringVar(&issuerURL, "issuer", "", "Issuer URL (e.g., http://localhost:8080)")
	setupCmd.Flags().StringVar(&adminUsername, "admin-user", "", "Admin username (optional)")
	setupCmd.Flags().StringVar(&adminPassword, "admin-pass", "", "Admin password (optional)")
	setupCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Non-interactive mode (use flags or env vars)")
}

func runSetup(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	fmt.Println("🚀 OpenID Connect Server Setup (CLI)")
	fmt.Println("====================================")
	fmt.Println()

	// Try to auto-load config store (MongoDB env or JSON file)
	loaderCfg := configstore.LoaderConfig{
		MongoURIEnv:      "MONGODB_URI",
		MongoDatabaseEnv: "MONGODB_DATABASE",
		JSONFilePath:     "data/config.json",
	}

	configStoreInstance, initialized, err := configstore.AutoLoadConfigStore(ctx, loaderCfg)
	if err != nil {
		fmt.Printf("❌ Failed to initialize config store: %v\n", err)
		os.Exit(1)
	}
	defer configStoreInstance.Close()

	// Check if already initialized
	if initialized {
		fmt.Println("⚠️  Configuration already exists!")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to reinitialize? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Setup cancelled.")
			return
		}
	}

	// Get configuration values
	if err := gatherConfiguration(); err != nil {
		fmt.Printf("❌ Failed to gather configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := validateConfiguration(); err != nil {
		fmt.Printf("❌ Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize config store
	fmt.Println("\n� Initializing configuration...")
	if err := configstore.InitializeMinimalConfig(ctx, configStoreInstance, issuerURL); err != nil {
		fmt.Printf("❌ Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Configuration initialized with auto-generated JWT keys")

	// Create admin user if provided
	if adminUsername != "" && adminPassword != "" {
		fmt.Println("\n👤 Creating admin user...")
		if err := createAdminUserCLI(ctx, configStoreInstance); err != nil {
			fmt.Printf("⚠️  Failed to create admin user: %v\n", err)
			fmt.Println("You can create users later via the web UI or API")
		} else {
			fmt.Printf("✓ Admin user '%s' created successfully\n", adminUsername)
		}
	}

	fmt.Println("\n✅ Setup completed successfully!")
	fmt.Println("\nConfiguration stored in:", getStorageLocation(loaderCfg))
	fmt.Println("\nYou can now start the server with:")
	fmt.Println("  ./openid-server serve")
	fmt.Println("or simply:")
	fmt.Println("  ./openid-server")
	if adminUsername != "" {
		fmt.Printf("\nLogin with:\n  Username: %s\n  Password: %s\n", adminUsername, adminPassword)
	}
}

// gatherConfiguration collects configuration from flags, env vars, or interactive prompts
func gatherConfiguration() error {
	reader := bufio.NewReader(os.Stdin)

	// Get issuer URL
	if issuerURL == "" {
		issuerURL = os.Getenv("ISSUER_URL")
	}
	if issuerURL == "" && !nonInteractive {
		issuerURL = promptWithDefault(reader, "Issuer URL (e.g., http://localhost:8080)", "http://localhost:8080")
	}

	// Get admin credentials (optional)
	if adminUsername == "" {
		adminUsername = os.Getenv("ADMIN_USER")
	}
	if adminPassword == "" {
		adminPassword = os.Getenv("ADMIN_PASS")
	}

	if adminUsername == "" && adminPassword == "" && !nonInteractive {
		fmt.Println("\n👤 Admin User (Optional)")
		fmt.Println("You can create an admin user now, or skip and create users later via the web UI.")
		fmt.Print("Create admin user now? (y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			fmt.Print("Admin username: ")
			username, _ := reader.ReadString('\n')
			adminUsername = strings.TrimSpace(username)

			fmt.Print("Admin password (min 6 characters): ")
			password, _ := reader.ReadString('\n')
			adminPassword = strings.TrimSpace(password)
		}
	}

	return nil
}

// validateConfiguration validates the gathered configuration
func validateConfiguration() error {
	if issuerURL == "" {
		return fmt.Errorf("issuer URL is required")
	}

	if adminUsername != "" && adminPassword == "" {
		return fmt.Errorf("admin password is required when username is provided")
	}

	if adminPassword != "" && adminUsername == "" {
		return fmt.Errorf("admin username is required when password is provided")
	}

	if adminPassword != "" && len(adminPassword) < 6 {
		return fmt.Errorf("admin password must be at least 6 characters long")
	}

	return nil
}

// createAdminUserCLI creates an admin user using the config store
func createAdminUserCLI(ctx context.Context, configStoreInstance configstore.ConfigStore) error {
	// Load the config from store
	configData, err := configStoreInstance.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Convert to config.Config for storage initialization
	cfg := convertConfigData(configData)

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Hash password
	hashedPassword, err := crypto.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := &models.User{
		ID:           uuid.New().String(),
		Username:     adminUsername,
		PasswordHash: hashedPassword,
		Email:        adminUsername + "@local",
		Role:         "admin",
		Name:         "Administrator",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if createErr := store.CreateUser(adminUser); createErr != nil {
		return fmt.Errorf("failed to create user: %w", createErr)
	}

	return nil
}

// getStorageLocation returns a human-readable description of where config is stored
func getStorageLocation(loaderCfg configstore.LoaderConfig) string {
	if mongoURI := os.Getenv(loaderCfg.MongoURIEnv); mongoURI != "" {
		return "MongoDB (from " + loaderCfg.MongoURIEnv + " environment variable)"
	}
	return loaderCfg.JSONFilePath
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
