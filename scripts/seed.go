package main

import (
	"log"

	"github.com/prasenjit/openid-golang/internal/config"
	"github.com/prasenjit/openid-golang/internal/crypto"
	"github.com/prasenjit/openid-golang/internal/models"
	"github.com/prasenjit/openid-golang/internal/storage"
)

func main() {
	log.Println("Seeding database with test data...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create test user
	passwordHash, _ := crypto.HashPassword("password123")
	user := models.NewUser("testuser", "test@example.com", passwordHash)
	user.Name = "Test User"
	user.GivenName = "Test"
	user.FamilyName = "User"

	if err := store.CreateUser(user); err != nil {
		log.Printf("Warning: Failed to create user (may already exist): %v", err)
	} else {
		log.Printf("✓ Created test user: %s (password: password123)", user.Username)
	}

	// Create test client
	client := models.NewClient("Test Client", []string{"http://localhost:3000/callback"})
	if err := store.CreateClient(client); err != nil {
		log.Printf("Warning: Failed to create client (may already exist): %v", err)
	} else {
		log.Printf("✓ Created test client:")
		log.Printf("  Client ID: %s", client.ID)
		log.Printf("  Client Secret: %s", client.Secret)
		log.Printf("  Redirect URIs: %v", client.RedirectURIs)
	}

	log.Println("\nSeeding complete!")
}
