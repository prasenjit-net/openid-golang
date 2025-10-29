package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// JSONStorage implements Storage interface using JSON file
type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
	data     *JSONData
}

// JSONUser represents a user with password hash for JSON storage
type JSONUser struct {
	*models.User
	PasswordHash string `json:"password_hash"`
}

// JSONData holds all the data
type JSONData struct {
	Users               map[string]*JSONUser                  `json:"users"`
	Clients             map[string]*models.Client             `json:"clients"`
	AuthorizationCodes  map[string]*models.AuthorizationCode  `json:"authorization_codes"`
	Tokens              map[string]*models.Token              `json:"tokens"`
	Sessions            map[string]*models.Session            `json:"sessions"`
	AuthSessions        map[string]*models.AuthSession        `json:"auth_sessions"`
	UserSessions        map[string]*models.UserSession        `json:"user_sessions"`
	Consents            map[string]*models.Consent            `json:"consents"`              // Key: userID:clientID
	InitialAccessTokens map[string]*models.InitialAccessToken `json:"initial_access_tokens"` // Key: token
	SigningKeys         map[string]*models.SigningKey         `json:"signing_keys"`          // Key: key ID
}

// NewJSONStorage creates a new JSON file storage
func NewJSONStorage(filePath string) (*JSONStorage, error) {
	storage := &JSONStorage{
		filePath: filePath,
		data: &JSONData{
			Users:               make(map[string]*JSONUser),
			Clients:             make(map[string]*models.Client),
			AuthorizationCodes:  make(map[string]*models.AuthorizationCode),
			Tokens:              make(map[string]*models.Token),
			Sessions:            make(map[string]*models.Session),
			AuthSessions:        make(map[string]*models.AuthSession),
			UserSessions:        make(map[string]*models.UserSession),
			Consents:            make(map[string]*models.Consent),
			InitialAccessTokens: make(map[string]*models.InitialAccessToken),
			SigningKeys:         make(map[string]*models.SigningKey),
		},
	}

	// Load existing data if file exists
	if _, err := os.Stat(filePath); err == nil {
		if loadErr := storage.load(); loadErr != nil {
			return nil, fmt.Errorf("failed to load existing data: %w", loadErr)
		}
	} else {
		// Create new file
		if saveErr := storage.save(); saveErr != nil {
			return nil, fmt.Errorf("failed to create data file: %w", saveErr)
		}
	}

	return storage, nil
}

func (j *JSONStorage) load() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, j.data)
}

func (j *JSONStorage) save() error {
	data, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(j.filePath, data, 0600)
}

func (j *JSONStorage) Close() error {
	return j.save()
}

// User operations
func (j *JSONStorage) CreateUser(user *models.User) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Check for duplicates
	for _, u := range j.data.Users {
		if u.Username == user.Username {
			return fmt.Errorf("username already exists")
		}
		if u.Email == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Store as JSONUser with password hash
	jsonUser := &JSONUser{
		User:         user,
		PasswordHash: user.PasswordHash,
	}
	j.data.Users[user.ID] = jsonUser
	return j.save()
}

func (j *JSONStorage) GetUserByUsername(username string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, jsonUser := range j.data.Users {
		if jsonUser.Username == username {
			// Restore password hash from JSON storage
			user := jsonUser.User
			user.PasswordHash = jsonUser.PasswordHash
			return user, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) GetUserByID(id string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	jsonUser, exists := j.data.Users[id]
	if !exists {
		return nil, nil
	}
	// Restore password hash from JSON storage
	user := jsonUser.User
	user.PasswordHash = jsonUser.PasswordHash
	return user, nil
}

func (j *JSONStorage) GetUserByEmail(email string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, jsonUser := range j.data.Users {
		if jsonUser.Email == email {
			// Restore password hash from JSON storage
			user := jsonUser.User
			user.PasswordHash = jsonUser.PasswordHash
			return user, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) GetAllUsers() ([]*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	users := make([]*models.User, 0, len(j.data.Users))
	for _, jsonUser := range j.data.Users {
		user := jsonUser.User
		user.PasswordHash = jsonUser.PasswordHash
		users = append(users, user)
	}
	return users, nil
}

func (j *JSONStorage) UpdateUser(user *models.User) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	jsonUser, exists := j.data.Users[user.ID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.UpdatedAt = time.Now()
	user.CreatedAt = jsonUser.CreatedAt // Preserve creation time

	// Update the JSON user
	jsonUser.User = user
	jsonUser.PasswordHash = user.PasswordHash

	return j.save()
}

func (j *JSONStorage) DeleteUser(id string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, exists := j.data.Users[id]; !exists {
		return fmt.Errorf("user not found")
	}

	delete(j.data.Users, id)
	return j.save()
}

// Client operations
func (j *JSONStorage) CreateClient(client *models.Client) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	client.CreatedAt = time.Now()
	j.data.Clients[client.ID] = client
	return j.save()
}

func (j *JSONStorage) GetClientByID(id string) (*models.Client, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	client, exists := j.data.Clients[id]
	if !exists {
		return nil, nil
	}
	return client, nil
}

func (j *JSONStorage) GetAllClients() ([]*models.Client, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	clients := make([]*models.Client, 0, len(j.data.Clients))
	for _, client := range j.data.Clients {
		clients = append(clients, client)
	}
	return clients, nil
}

func (j *JSONStorage) UpdateClient(client *models.Client) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	existing, exists := j.data.Clients[client.ID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	// Preserve creation time
	client.CreatedAt = existing.CreatedAt
	j.data.Clients[client.ID] = client
	return j.save()
}

func (j *JSONStorage) DeleteClient(id string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, exists := j.data.Clients[id]; !exists {
		return fmt.Errorf("client not found")
	}

	delete(j.data.Clients, id)
	return j.save()
}

func (j *JSONStorage) ValidateClient(clientID, clientSecret string) (*models.Client, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	client, exists := j.data.Clients[clientID]
	if !exists {
		return nil, nil
	}
	if client.Secret != clientSecret {
		return nil, nil
	}
	return client, nil
}

// Authorization code operations
func (j *JSONStorage) CreateAuthorizationCode(code *models.AuthorizationCode) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	code.CreatedAt = time.Now()
	j.data.AuthorizationCodes[code.Code] = code
	return j.save()
}

func (j *JSONStorage) GetAuthorizationCode(code string) (*models.AuthorizationCode, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	authCode, exists := j.data.AuthorizationCodes[code]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, nil
	}

	return authCode, nil
}

func (j *JSONStorage) UpdateAuthorizationCode(code *models.AuthorizationCode) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, exists := j.data.AuthorizationCodes[code.Code]; !exists {
		return nil
	}

	j.data.AuthorizationCodes[code.Code] = code
	return j.save()
}

func (j *JSONStorage) DeleteAuthorizationCode(code string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.AuthorizationCodes, code)
	return j.save()
}

// Token operations
func (j *JSONStorage) CreateToken(token *models.Token) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	token.CreatedAt = time.Now()
	j.data.Tokens[token.ID] = token
	return j.save()
}

func (j *JSONStorage) GetTokenByAccessToken(accessToken string) (*models.Token, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, token := range j.data.Tokens {
		if token.AccessToken == accessToken {
			// Check if expired
			if time.Now().After(token.ExpiresAt) {
				return nil, nil
			}
			return token, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) GetTokenByRefreshToken(refreshToken string) (*models.Token, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, token := range j.data.Tokens {
		if token.RefreshToken == refreshToken {
			return token, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) DeleteToken(tokenID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.Tokens, tokenID)
	return j.save()
}

func (j *JSONStorage) GetTokensByAuthCode(authCodeID string) ([]*models.Token, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var tokens []*models.Token
	for _, token := range j.data.Tokens {
		if token.AuthorizationCodeID == authCodeID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (j *JSONStorage) RevokeTokensByAuthCode(authCodeID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Find and delete all tokens associated with this authorization code
	for tokenID, token := range j.data.Tokens {
		if token.AuthorizationCodeID == authCodeID {
			delete(j.data.Tokens, tokenID)
		}
	}
	return j.save()
}

// Session operations
func (j *JSONStorage) CreateSession(session *models.Session) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	session.CreatedAt = time.Now()
	j.data.Sessions[session.ID] = session
	return j.save()
}

func (j *JSONStorage) GetSession(sessionID string) (*models.Session, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	session, exists := j.data.Sessions[sessionID]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	return session, nil
}

func (j *JSONStorage) DeleteSession(sessionID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.Sessions, sessionID)
	return j.save()
}

// AuthSession operations
func (j *JSONStorage) CreateAuthSession(session *models.AuthSession) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	j.data.AuthSessions[session.ID] = session
	return j.save()
}

func (j *JSONStorage) GetAuthSession(sessionID string) (*models.AuthSession, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	session, exists := j.data.AuthSessions[sessionID]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	return session, nil
}

func (j *JSONStorage) UpdateAuthSession(session *models.AuthSession) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data.AuthSessions[session.ID] = session
	return j.save()
}

func (j *JSONStorage) DeleteAuthSession(sessionID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.AuthSessions, sessionID)
	return j.save()
}

// UserSession operations
func (j *JSONStorage) CreateUserSession(session *models.UserSession) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.AuthTime.IsZero() {
		session.AuthTime = time.Now()
	}
	session.LastActivityAt = time.Now()
	j.data.UserSessions[session.ID] = session
	return j.save()
}

func (j *JSONStorage) GetUserSession(sessionID string) (*models.UserSession, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	session, exists := j.data.UserSessions[sessionID]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	return session, nil
}

func (j *JSONStorage) GetUserSessionByUserID(userID string) (*models.UserSession, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// Find the most recent session for the user
	var latestSession *models.UserSession
	for _, session := range j.data.UserSessions {
		if session.UserID == userID && time.Now().Before(session.ExpiresAt) {
			if latestSession == nil || session.AuthTime.After(latestSession.AuthTime) {
				latestSession = session
			}
		}
	}

	return latestSession, nil
}

func (j *JSONStorage) UpdateUserSession(session *models.UserSession) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	session.LastActivityAt = time.Now()
	j.data.UserSessions[session.ID] = session
	return j.save()
}

func (j *JSONStorage) DeleteUserSession(sessionID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.UserSessions, sessionID)
	return j.save()
}

func (j *JSONStorage) CleanupExpiredSessions() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	now := time.Now()
	deleted := 0

	// Clean up auth sessions
	for id, session := range j.data.AuthSessions {
		if now.After(session.ExpiresAt) {
			delete(j.data.AuthSessions, id)
			deleted++
		}
	}

	// Clean up user sessions
	for id, session := range j.data.UserSessions {
		if now.After(session.ExpiresAt) {
			delete(j.data.UserSessions, id)
			deleted++
		}
	}

	// Clean up old sessions
	for id, session := range j.data.Sessions {
		if now.After(session.ExpiresAt) {
			delete(j.data.Sessions, id)
			deleted++
		}
	}

	if deleted > 0 {
		return j.save()
	}
	return nil
}

// Consent operations
func (j *JSONStorage) CreateConsent(consent *models.Consent) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if consent.CreatedAt.IsZero() {
		consent.CreatedAt = time.Now()
	}
	consent.UpdatedAt = time.Now()

	key := fmt.Sprintf("%s:%s", consent.UserID, consent.ClientID)
	j.data.Consents[key] = consent
	return j.save()
}

func (j *JSONStorage) GetConsent(userID, clientID string) (*models.Consent, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, clientID)
	consent, exists := j.data.Consents[key]
	if !exists {
		return nil, nil
	}

	return consent, nil
}

func (j *JSONStorage) UpdateConsent(consent *models.Consent) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	consent.UpdatedAt = time.Now()
	key := fmt.Sprintf("%s:%s", consent.UserID, consent.ClientID)
	j.data.Consents[key] = consent
	return j.save()
}

func (j *JSONStorage) DeleteConsent(userID, clientID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, clientID)
	delete(j.data.Consents, key)
	return j.save()
}

func (j *JSONStorage) DeleteConsentsForUser(userID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	deleted := 0
	for key, consent := range j.data.Consents {
		if consent.UserID == userID {
			delete(j.data.Consents, key)
			deleted++
		}
	}

	if deleted > 0 {
		return j.save()
	}
	return nil
}

// ============================================================================
// Initial Access Token Operations
// ============================================================================

func (j *JSONStorage) CreateInitialAccessToken(token *models.InitialAccessToken) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data.InitialAccessTokens[token.Token] = token
	return j.save()
}

func (j *JSONStorage) GetInitialAccessToken(token string) (*models.InitialAccessToken, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	t, exists := j.data.InitialAccessTokens[token]
	if !exists {
		return nil, nil
	}
	return t, nil
}

func (j *JSONStorage) UpdateInitialAccessToken(token *models.InitialAccessToken) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data.InitialAccessTokens[token.Token] = token
	return j.save()
}

func (j *JSONStorage) DeleteInitialAccessToken(token string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.InitialAccessTokens, token)
	return j.save()
}

func (j *JSONStorage) GetAllInitialAccessTokens() ([]*models.InitialAccessToken, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	tokens := make([]*models.InitialAccessToken, 0, len(j.data.InitialAccessTokens))
	for _, token := range j.data.InitialAccessTokens {
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// SigningKey operations

func (j *JSONStorage) CreateSigningKey(key *models.SigningKey) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data.SigningKeys[key.ID] = key
	return j.save()
}

func (j *JSONStorage) GetSigningKey(id string) (*models.SigningKey, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	key, exists := j.data.SigningKeys[id]
	if !exists {
		return nil, nil
	}
	return key, nil
}

func (j *JSONStorage) GetSigningKeyByKID(kid string) (*models.SigningKey, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, key := range j.data.SigningKeys {
		if key.KID == kid {
			return key, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) GetAllSigningKeys() ([]*models.SigningKey, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	keys := make([]*models.SigningKey, 0, len(j.data.SigningKeys))
	for _, key := range j.data.SigningKeys {
		keys = append(keys, key)
	}
	return keys, nil
}

func (j *JSONStorage) GetActiveSigningKey() (*models.SigningKey, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, key := range j.data.SigningKeys {
		if key.IsActive && !key.IsExpired() {
			return key, nil
		}
	}
	return nil, fmt.Errorf("no active signing key found")
}

func (j *JSONStorage) UpdateSigningKey(key *models.SigningKey) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, exists := j.data.SigningKeys[key.ID]; !exists {
		return fmt.Errorf("signing key not found")
	}

	j.data.SigningKeys[key.ID] = key
	return j.save()
}

func (j *JSONStorage) DeleteSigningKey(id string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.data.SigningKeys, id)
	return j.save()
}

// GetActiveTokensCount returns the count of non-expired tokens
func (j *JSONStorage) GetActiveTokensCount() int {
	j.mu.RLock()
	defer j.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, token := range j.data.Tokens {
		if token.ExpiresAt.After(now) {
			count++
		}
	}
	return count
}

// GetRecentUserSessionsCount returns the count of user sessions created in the last 24 hours
func (j *JSONStorage) GetRecentUserSessionsCount() int {
	j.mu.RLock()
	defer j.mu.RUnlock()

	count := 0
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, session := range j.data.UserSessions {
		if session.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count
}
