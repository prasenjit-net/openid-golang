package configstore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateJWTKeyPair generates an RSA key pair and returns them as PEM-encoded strings
func GenerateJWTKeyPair(bits int) (privateKeyPEM, publicKeyPEM string, err error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyPEM = string(pem.EncodeToMemory(privateKeyBlock))

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEM = string(pem.EncodeToMemory(publicKeyBlock))

	return privateKeyPEM, publicKeyPEM, nil
}

// ValidateJWTKeys validates that the provided PEM-encoded keys are valid
func ValidateJWTKeys(privateKeyPEM, publicKeyPEM string) error {
	// Parse private key
	privateBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if privateBlock == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	_, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Parse public key
	publicBlock, _ := pem.Decode([]byte(publicKeyPEM))
	if publicBlock == nil {
		return fmt.Errorf("failed to decode public key PEM")
	}

	_, err = x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	return nil
}
