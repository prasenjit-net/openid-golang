package models

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	user := NewUser("testuser", "test@example.com", "hashed_password", RoleUser)

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

func TestUserRoles(t *testing.T) {
	adminUser := NewAdminUser("admin", "admin@example.com", "hashed_password")
	if !adminUser.IsAdmin() {
		t.Error("Admin user should have admin role")
	}
	if adminUser.Role != RoleAdmin {
		t.Errorf("Expected role 'admin', got '%s'", adminUser.Role)
	}
	if !adminUser.HasRole(RoleAdmin) {
		t.Error("Admin user should have admin role")
	}

	regularUser := NewRegularUser("user", "user@example.com", "hashed_password")
	if regularUser.IsAdmin() {
		t.Error("Regular user should not have admin role")
	}
	if regularUser.Role != RoleUser {
		t.Errorf("Expected role 'user', got '%s'", regularUser.Role)
	}
	if !regularUser.HasRole(RoleUser) {
		t.Error("Regular user should have user role")
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
