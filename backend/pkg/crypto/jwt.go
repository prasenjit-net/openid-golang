package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiry     time.Duration
	keyID      string // Key ID for JWT header
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(privateKeyPath, publicKeyPath, issuer string, expiryMinutes int) (*JWTManager, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		expiry:     time.Duration(expiryMinutes) * time.Minute,
		keyID:      "default", // Use "default" to match JWKS endpoint
	}, nil
}

// NewJWTManagerFromPEM creates a new JWT manager from PEM-encoded key strings
func NewJWTManagerFromPEM(privateKeyPEM, publicKeyPEM, issuer string, expiryMinutes int) (*JWTManager, error) {
	privateKey, err := parsePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := parsePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		expiry:     time.Duration(expiryMinutes) * time.Minute,
		keyID:      "default", // Use "default" to match JWKS endpoint
	}, nil
}

// NewJWTManagerForTesting creates a new JWT manager with generated keys for testing
func NewJWTManagerForTesting(issuer string, expiryMinutes int) (*JWTManager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     issuer,
		expiry:     time.Duration(expiryMinutes) * time.Minute,
		keyID:      "default", // Use "default" to match JWKS endpoint
	}, nil
}

// IDTokenClaims represents OpenID Connect ID Token claims
type IDTokenClaims struct {
	jwt.RegisteredClaims
	Name       string   `json:"name,omitempty"`
	GivenName  string   `json:"given_name,omitempty"`
	FamilyName string   `json:"family_name,omitempty"`
	Email      string   `json:"email,omitempty"`
	Picture    string   `json:"picture,omitempty"`
	Nonce      string   `json:"nonce,omitempty"`
	AuthTime   *int64   `json:"auth_time,omitempty"`
	ACR        string   `json:"acr,omitempty"`
	AMR        []string `json:"amr,omitempty"`
	AtHash     string   `json:"at_hash,omitempty"` // Access token hash for implicit/hybrid flows
	CHash      string   `json:"c_hash,omitempty"`  // Authorization code hash for hybrid flows
}

// GenerateIDToken generates an OpenID Connect ID token
func (jm *JWTManager) GenerateIDToken(user *models.User, clientID, nonce string) (string, error) {
	now := time.Now()
	claims := IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Name:       user.Name,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Email:      user.Email,
		Picture:    user.Picture,
		Nonce:      nonce,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = jm.keyID // Add key ID to header
	return token.SignedString(jm.privateKey)
}

// GenerateIDTokenWithClaims generates an OpenID Connect ID token with additional OIDC claims
// accessToken and authCode are optional - if provided, at_hash and c_hash will be included
func (jm *JWTManager) GenerateIDTokenWithClaims(user *models.User, clientID, nonce string, authTime time.Time, acr string, amr []string, accessToken, authCode string) (string, error) {
	now := time.Now()
	authTimeUnix := authTime.Unix()

	claims := IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Name:       user.Name,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Email:      user.Email,
		Picture:    user.Picture,
		Nonce:      nonce,
		AuthTime:   &authTimeUnix,
		ACR:        acr,
		AMR:        amr,
	}

	// Include at_hash if access token is provided (implicit/hybrid flows)
	if accessToken != "" {
		claims.AtHash = CalculateTokenHash(accessToken)
	}

	// Include c_hash if authorization code is provided (hybrid flows)
	if authCode != "" {
		claims.CHash = CalculateTokenHash(authCode)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = jm.keyID // Add key ID to header
	return token.SignedString(jm.privateKey)
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(tokenString string) (*IDTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &IDTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*IDTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// AccessTokenClaims represents OAuth 2.0 Access Token claims
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

// GenerateAccessToken generates an OAuth 2.0 access token
func (jm *JWTManager) GenerateAccessToken(user *models.User, clientID, scope string) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Scope: scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jm.privateKey)
}

// GetPublicKey returns the public key
func (jm *JWTManager) GetPublicKey() *rsa.PublicKey {
	return jm.publicKey
}

// loadPrivateKey loads an RSA private key from a PEM file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	return key, nil
}

// loadPublicKey loads an RSA public key from a PEM file
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
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

// parsePrivateKeyPEM parses an RSA private key from a PEM-encoded string
func parsePrivateKeyPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	return key, nil
}

// parsePublicKeyPEM parses an RSA public key from a PEM-encoded string
func parsePublicKeyPEM(pemData string) (*rsa.PublicKey, error) {
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
