package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prasenjit/openid-golang/internal/models"
)

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database tables
func (s *SQLiteStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT,
		given_name TEXT,
		family_name TEXT,
		picture TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS clients (
		id TEXT PRIMARY KEY,
		secret TEXT NOT NULL,
		name TEXT NOT NULL,
		redirect_uris TEXT NOT NULL,
		grant_types TEXT NOT NULL,
		response_types TEXT NOT NULL,
		scope TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS authorization_codes (
		code TEXT PRIMARY KEY,
		client_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		redirect_uri TEXT NOT NULL,
		scope TEXT NOT NULL,
		nonce TEXT,
		code_challenge TEXT,
		code_challenge_method TEXT,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tokens (
		id TEXT PRIMARY KEY,
		access_token TEXT UNIQUE NOT NULL,
		refresh_token TEXT,
		token_type TEXT NOT NULL,
		client_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		scope TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_tokens_access ON tokens(access_token);
	CREATE INDEX IF NOT EXISTS idx_tokens_refresh ON tokens(refresh_token);
	`

	_, err := s.db.Exec(schema)
	return err
}

// User operations

func (s *SQLiteStorage) CreateUser(user *models.User) error {
	_, err := s.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, name, given_name, family_name, picture, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Name, user.GivenName, user.FamilyName, user.Picture, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, name, given_name, family_name, picture, created_at, updated_at
		FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Name, &user.GivenName, &user.FamilyName, &user.Picture, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (s *SQLiteStorage) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, name, given_name, family_name, picture, created_at, updated_at
		FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Name, &user.GivenName, &user.FamilyName, &user.Picture, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (s *SQLiteStorage) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, name, given_name, family_name, picture, created_at, updated_at
		FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Name, &user.GivenName, &user.FamilyName, &user.Picture, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

// Client operations

func (s *SQLiteStorage) CreateClient(client *models.Client) error {
	redirectURIs, _ := json.Marshal(client.RedirectURIs)
	grantTypes, _ := json.Marshal(client.GrantTypes)
	responseTypes, _ := json.Marshal(client.ResponseTypes)

	_, err := s.db.Exec(`
		INSERT INTO clients (id, secret, name, redirect_uris, grant_types, response_types, scope, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		client.ID, client.Secret, client.Name, redirectURIs, grantTypes, responseTypes, client.Scope, client.CreatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetClientByID(id string) (*models.Client, error) {
	client := &models.Client{}
	var redirectURIs, grantTypes, responseTypes string

	err := s.db.QueryRow(`
		SELECT id, secret, name, redirect_uris, grant_types, response_types, scope, created_at
		FROM clients WHERE id = ?`, id,
	).Scan(&client.ID, &client.Secret, &client.Name, &redirectURIs, &grantTypes, &responseTypes, &client.Scope, &client.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("client not found")
	}
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(redirectURIs), &client.RedirectURIs)
	_ = json.Unmarshal([]byte(grantTypes), &client.GrantTypes)
	_ = json.Unmarshal([]byte(responseTypes), &client.ResponseTypes)

	return client, nil
}

func (s *SQLiteStorage) ValidateClient(clientID, clientSecret string) (*models.Client, error) {
	client, err := s.GetClientByID(clientID)
	if err != nil {
		return nil, err
	}
	if client.Secret != clientSecret {
		return nil, fmt.Errorf("invalid client credentials")
	}
	return client, nil
}

// Authorization code operations

func (s *SQLiteStorage) CreateAuthorizationCode(code *models.AuthorizationCode) error {
	_, err := s.db.Exec(`
		INSERT INTO authorization_codes (code, client_id, user_id, redirect_uri, scope, nonce, code_challenge, code_challenge_method, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scope, code.Nonce, code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt, code.CreatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetAuthorizationCode(code string) (*models.AuthorizationCode, error) {
	authCode := &models.AuthorizationCode{}
	err := s.db.QueryRow(`
		SELECT code, client_id, user_id, redirect_uri, scope, nonce, code_challenge, code_challenge_method, expires_at, created_at
		FROM authorization_codes WHERE code = ?`, code,
	).Scan(&authCode.Code, &authCode.ClientID, &authCode.UserID, &authCode.RedirectURI, &authCode.Scope, &authCode.Nonce, &authCode.CodeChallenge, &authCode.CodeChallengeMethod, &authCode.ExpiresAt, &authCode.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("authorization code not found")
	}
	return authCode, err
}

func (s *SQLiteStorage) DeleteAuthorizationCode(code string) error {
	_, err := s.db.Exec("DELETE FROM authorization_codes WHERE code = ?", code)
	return err
}

// Token operations

func (s *SQLiteStorage) CreateToken(token *models.Token) error {
	_, err := s.db.Exec(`
		INSERT INTO tokens (id, access_token, refresh_token, token_type, client_id, user_id, scope, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		token.ID, token.AccessToken, token.RefreshToken, token.TokenType, token.ClientID, token.UserID, token.Scope, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetTokenByAccessToken(accessToken string) (*models.Token, error) {
	token := &models.Token{}
	err := s.db.QueryRow(`
		SELECT id, access_token, refresh_token, token_type, client_id, user_id, scope, expires_at, created_at
		FROM tokens WHERE access_token = ?`, accessToken,
	).Scan(&token.ID, &token.AccessToken, &token.RefreshToken, &token.TokenType, &token.ClientID, &token.UserID, &token.Scope, &token.ExpiresAt, &token.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	return token, err
}

func (s *SQLiteStorage) GetTokenByRefreshToken(refreshToken string) (*models.Token, error) {
	token := &models.Token{}
	err := s.db.QueryRow(`
		SELECT id, access_token, refresh_token, token_type, client_id, user_id, scope, expires_at, created_at
		FROM tokens WHERE refresh_token = ?`, refreshToken,
	).Scan(&token.ID, &token.AccessToken, &token.RefreshToken, &token.TokenType, &token.ClientID, &token.UserID, &token.Scope, &token.ExpiresAt, &token.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	return token, err
}

func (s *SQLiteStorage) DeleteToken(accessToken string) error {
	_, err := s.db.Exec("DELETE FROM tokens WHERE access_token = ?", accessToken)
	return err
}

// Session operations

func (s *SQLiteStorage) CreateSession(session *models.Session) error {
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)`,
		session.ID, session.UserID, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetSession(id string) (*models.Session, error) {
	session := &models.Session{}
	err := s.db.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions WHERE id = ?`, id,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	return session, err
}

func (s *SQLiteStorage) DeleteSession(id string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
