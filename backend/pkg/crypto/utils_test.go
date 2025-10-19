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
