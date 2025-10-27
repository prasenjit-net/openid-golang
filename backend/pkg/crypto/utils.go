package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidatePassword validates a password against a hash
func ValidatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// VerifyCodeChallenge verifies a PKCE code challenge
func VerifyCodeChallenge(codeVerifier, codeChallenge, method string) bool {
	if method == "plain" {
		return codeVerifier == codeChallenge
	}

	if method == "S256" {
		hash := sha256.Sum256([]byte(codeVerifier))
		computed := base64.RawURLEncoding.EncodeToString(hash[:])
		return computed == codeChallenge
	}

	return false
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// PublicKeyToJWKS converts an RSA public key to JWKS format
func PublicKeyToJWKS(publicKey *rsa.PublicKey, keyID string) (*JWKS, error) {
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: keyID,
		Alg: "RS256",
		N:   n,
		E:   e,
	}

	return &JWKS{Keys: []JWK{jwk}}, nil
}

// MarshalJWKS marshals JWKS to JSON
func MarshalJWKS(jwks *JWKS) ([]byte, error) {
	return json.Marshal(jwks)
}

// GenerateState generates a random state parameter for OAuth2
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CalculateTokenHash calculates the hash of a token for at_hash or c_hash claims
// as specified in OIDC Core spec sections 3.2.2.9 (at_hash) and 3.3.2.11 (c_hash)
// For RS256, uses SHA-256 and takes the left-most 128 bits (16 bytes), base64url encoded
func CalculateTokenHash(token string) string {
	// Hash the token using SHA-256 (for RS256 algorithm)
	hash := sha256.Sum256([]byte(token))

	// Take the left-most half (128 bits = 16 bytes for SHA-256)
	leftHalf := hash[:len(hash)/2]

	// Base64url encode without padding
	return base64.RawURLEncoding.EncodeToString(leftHalf)
}

// GenerateRSAKeyPair generates a new RSA key pair for signing
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// EncodePrivateKeyToPEM encodes an RSA private key to PEM format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if privateKeyPEM == nil {
		return "", fmt.Errorf("failed to encode private key to PEM")
	}
	return string(privateKeyPEM), nil
}

// EncodePublicKeyToPEM encodes an RSA public key to PEM format
func EncodePublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if publicKeyPEM == nil {
		return "", fmt.Errorf("failed to encode public key to PEM")
	}
	return string(publicKeyPEM), nil
}
