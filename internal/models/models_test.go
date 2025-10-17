package models

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	user := NewUser("testuser", "test@example.com", "hashed_password")
	
	if user.ID == "" {
		t.Error("User ID should not be empty")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestAuthorizationCodeExpiry(t *testing.T) {
	code := &AuthorizationCode{
		Code:      "test-code",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
	}
	
	if !code.IsExpired() {
		t.Error("Authorization code should be expired")
	}
	
	code.ExpiresAt = time.Now().Add(10 * time.Minute) // Expires in 10 minutes
	if code.IsExpired() {
		t.Error("Authorization code should not be expired")
	}
}

func TestTokenExpiry(t *testing.T) {
	token := &Token{
		ID:        "test-token",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
	}
	
	if !token.IsExpired() {
		t.Error("Token should be expired")
	}
	
	token.ExpiresAt = time.Now().Add(60 * time.Minute) // Expires in 60 minutes
	if token.IsExpired() {
		t.Error("Token should not be expired")
	}
}

func TestNewClient(t *testing.T) {
	redirectURIs := []string{"http://localhost:3000/callback"}
	client := NewClient("Test Client", redirectURIs)
	
	if client.ID == "" {
		t.Error("Client ID should not be empty")
	}
	if client.Secret == "" {
		t.Error("Client secret should not be empty")
	}
	if client.Name != "Test Client" {
		t.Errorf("Expected name 'Test Client', got '%s'", client.Name)
	}
	if len(client.RedirectURIs) != 1 {
		t.Errorf("Expected 1 redirect URI, got %d", len(client.RedirectURIs))
	}
}
