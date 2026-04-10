package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

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
	Kty     string   `json:"kty"`
	Use     string   `json:"use"`
	Kid     string   `json:"kid"`
	Alg     string   `json:"alg"`
	N       string   `json:"n"`
	E       string   `json:"e"`
	X5c     []string `json:"x5c,omitempty"`      // Certificate chain (base64 DER, not base64url)
	X5tS256 string   `json:"x5t#S256,omitempty"` // SHA-256 thumbprint of leaf cert
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

// PublicKeyToJWKWithCert builds a JWK entry including x5c and x5t#S256 from a PEM cert.
func PublicKeyToJWKWithCert(publicKey *rsa.PublicKey, keyID, certPEM string) (JWK, error) {
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

	if certPEM != "" {
		block, _ := pem.Decode([]byte(certPEM))
		if block != nil {
			// x5c: standard base64 (not base64url) of DER bytes — RFC 7517 §4.7
			jwk.X5c = []string{base64.StdEncoding.EncodeToString(block.Bytes)}
			// x5t#S256: base64url of SHA-256 of DER bytes — RFC 7517 §4.9
			sum := sha256.Sum256(block.Bytes)
			jwk.X5tS256 = base64.RawURLEncoding.EncodeToString(sum[:])
		}
	}

	return jwk, nil
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

// SigningKeyMaterial holds all PEM-encoded artifacts for a newly generated signing key.
type SigningKeyMaterial struct {
	PrivateKeyPEM string
	PublicKeyPEM  string
	CertPEM       string
	KID           string // base64url(SHA-256(DER cert)) — RFC 7517 x5t#S256
	NotBefore     time.Time
	NotAfter      time.Time
}

// GenerateSigningKeyWithCert generates a 2048-bit RSA key pair and a self-signed X.509
// certificate valid for validityDays days. The KID is derived as the RFC 7517 x5t#S256
// thumbprint (base64url of SHA-256 of the DER-encoded certificate).
func GenerateSigningKeyWithCert(validityDays int) (*SigningKeyMaterial, error) {
	if validityDays <= 0 {
		validityDays = 90
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	now := time.Now().UTC()
	notAfter := now.AddDate(0, 0, validityDays)

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "openid-server",
			Organization: []string{"OpenID Server"},
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// KID = base64url(SHA-256(DER cert)) — same as x5t#S256
	sum := sha256.Sum256(certDER)
	kid := base64.RawURLEncoding.EncodeToString(sum[:])

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return &SigningKeyMaterial{
		PrivateKeyPEM: string(privateKeyPEM),
		PublicKeyPEM:  string(publicKeyPEM),
		CertPEM:       string(certPEM),
		KID:           kid,
		NotBefore:     now,
		NotAfter:      notAfter,
	}, nil
}

// CertThumbprintS256 returns the RFC 7517 x5t#S256 thumbprint for a PEM-encoded certificate.
func CertThumbprintS256(certPEM string) (string, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}
	sum := sha256.Sum256(block.Bytes)
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

// ParseCertFromPEM parses a PEM-encoded X.509 certificate.
func ParseCertFromPEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ParsePublicKeyFromPEM parses a PEM-encoded RSA public key (PKIX format).
func ParsePublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return key, nil
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
