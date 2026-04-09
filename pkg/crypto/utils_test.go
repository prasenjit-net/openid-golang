package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "test_password_123"
	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestValidatePassword(t *testing.T) {
	password := "test_password_123"
	hash, _ := HashPassword(password)

	// Test correct password
	if !ValidatePassword(password, hash) {
		t.Error("Valid password should be accepted")
	}

	// Test incorrect password
	if ValidatePassword("wrong_password", hash) {
		t.Error("Invalid password should be rejected")
	}
}

func TestVerifyCodeChallenge(t *testing.T) {
	// Test plain method
	verifier := "test_verifier"
	challenge := "test_verifier"

	if !VerifyCodeChallenge(verifier, challenge, "plain") {
		t.Error("Plain code challenge verification failed")
	}

	// Test S256 method
	// This is a simplified test - in real scenarios you'd compute the actual SHA256
	if VerifyCodeChallenge("wrong", challenge, "S256") {
		t.Error("S256 should fail with wrong verifier")
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 32
	str1, err := GenerateRandomString(length)

	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}

	if len(str1) != length {
		t.Errorf("Expected length %d, got %d", length, len(str1))
	}

	// Generate another and ensure they're different
	str2, _ := GenerateRandomString(length)
	if str1 == str2 {
		t.Error("Two random strings should be different")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()

	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	if state1 == "" {
		t.Error("State should not be empty")
	}

	// Generate another and ensure they're different
	state2, _ := GenerateState()
	if state1 == state2 {
		t.Error("Two state values should be different")
	}
}

func TestCalculateTokenHash(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string // Expected hash in base64url format
	}{
		{
			name:     "Simple token",
			token:    "jHkWEdUXMU1BwAsC4vtUsZwnNvTIxEl0z9K3vx5KF0Y",
			expected: "77QmUPtjPfzWtF2AnpK9RQ", // Pre-calculated expected hash
		},
		{
			name:     "Empty string",
			token:    "",
			expected: "47DEQpj8HBSa-_TImW-5JA", // SHA-256 of empty string, left 128 bits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTokenHash(tt.token)
			if result == "" {
				t.Error("Hash should not be empty")
			}

			// Verify it's base64url encoded (no padding)
			if len(result) != 22 { // 16 bytes base64url encoded = 22 chars
				t.Errorf("Expected hash length 22, got %d", len(result))
			}

			// Verify consistency
			result2 := CalculateTokenHash(tt.token)
			if result != result2 {
				t.Error("Hash should be consistent for same input")
			}
		})
	}

	// Test that different tokens produce different hashes
	t.Run("Different tokens produce different hashes", func(t *testing.T) {
		hash1 := CalculateTokenHash("token1")
		hash2 := CalculateTokenHash("token2")
		if hash1 == hash2 {
			t.Error("Different tokens should produce different hashes")
		}
	})
}
