package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/prasenjit-net/openid-golang/internal/models"
)

// JSONStorage implements Storage interface using JSON file
type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
	data     *JSONData
}

// JSONData holds all the data
type JSONData struct {
	Users              map[string]*models.User              `json:"users"`
	Clients            map[string]*models.Client            `json:"clients"`
	AuthorizationCodes map[string]*models.AuthorizationCode `json:"authorization_codes"`
	Tokens             map[string]*models.Token             `json:"tokens"`
	Sessions           map[string]*models.Session           `json:"sessions"`
}

// NewJSONStorage creates a new JSON file storage
func NewJSONStorage(filePath string) (*JSONStorage, error) {
	storage := &JSONStorage{
		filePath: filePath,
		data: &JSONData{
			Users:              make(map[string]*models.User),
			Clients:            make(map[string]*models.Client),
			AuthorizationCodes: make(map[string]*models.AuthorizationCode),
			Tokens:             make(map[string]*models.Token),
			Sessions:           make(map[string]*models.Session),
		},
	}

	// Load existing data if file exists
	if _, err := os.Stat(filePath); err == nil {
		if err := storage.load(); err != nil {
			return nil, fmt.Errorf("failed to load existing data: %w", err)
		}
	} else {
		// Create new file
		if err := storage.save(); err != nil {
			return nil, fmt.Errorf("failed to create data file: %w", err)
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
	j.data.Users[user.ID] = user
	return j.save()
}

func (j *JSONStorage) GetUserByUsername(username string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, user := range j.data.Users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (j *JSONStorage) GetUserByID(id string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	user, exists := j.data.Users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (j *JSONStorage) GetUserByEmail(email string) (*models.User, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, user := range j.data.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
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
