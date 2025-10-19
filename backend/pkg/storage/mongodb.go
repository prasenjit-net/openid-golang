package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// MongoDBStorage implements Storage interface using MongoDB
type MongoDBStorage struct {
	client   *mongo.Client
	db       *mongo.Database
	users    *mongo.Collection
	clients  *mongo.Collection
	codes    *mongo.Collection
	tokens   *mongo.Collection
	sessions *mongo.Collection
}

// NewMongoDBStorage creates a new MongoDB storage
func NewMongoDBStorage(connectionString, database string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(database)
	storage := &MongoDBStorage{
		client:   client,
		db:       db,
		users:    db.Collection("users"),
		clients:  db.Collection("clients"),
		codes:    db.Collection("authorization_codes"),
		tokens:   db.Collection("tokens"),
		sessions: db.Collection("sessions"),
	}

	// Create indexes
	if err := storage.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return storage, nil
}

func (m *MongoDBStorage) createIndexes() error {
	ctx := context.Background()

	// Users indexes
	_, _ = m.users.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
	})

	// Tokens indexes
	_, _ = m.tokens.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "access_token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "refresh_token", Value: 1}}},
	})

	// Codes index
	_, _ = m.codes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})

	// Sessions index
	_, _ = m.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})

	return nil
}

func (m *MongoDBStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

// User operations
func (m *MongoDBStorage) CreateUser(user *models.User) error {
	ctx := context.Background()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := m.users.InsertOne(ctx, user)
	return err
}

func (m *MongoDBStorage) GetUserByUsername(username string) (*models.User, error) {
	ctx := context.Background()
	var user models.User
	err := m.users.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (m *MongoDBStorage) GetUserByID(id string) (*models.User, error) {
	ctx := context.Background()
	var user models.User
	err := m.users.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (m *MongoDBStorage) GetUserByEmail(email string) (*models.User, error) {
	ctx := context.Background()
	var user models.User
	err := m.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (m *MongoDBStorage) GetAllUsers() ([]*models.User, error) {
	ctx := context.Background()
	cursor, err := m.users.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (m *MongoDBStorage) UpdateUser(user *models.User) error {
	ctx := context.Background()
	user.UpdatedAt = time.Now()
	_, err := m.users.UpdateOne(
		ctx,
		bson.M{"id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

func (m *MongoDBStorage) DeleteUser(id string) error {
	ctx := context.Background()
	_, err := m.users.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// Client operations
func (m *MongoDBStorage) CreateClient(client *models.Client) error {
	ctx := context.Background()
	client.CreatedAt = time.Now()
	_, err := m.clients.InsertOne(ctx, client)
	return err
}

func (m *MongoDBStorage) GetClientByID(id string) (*models.Client, error) {
	ctx := context.Background()
	var client models.Client
	err := m.clients.FindOne(ctx, bson.M{"id": id}).Decode(&client)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &client, err
}

func (m *MongoDBStorage) GetAllClients() ([]*models.Client, error) {
	ctx := context.Background()
	cursor, err := m.clients.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clients []*models.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (m *MongoDBStorage) UpdateClient(client *models.Client) error {
	ctx := context.Background()
	_, err := m.clients.UpdateOne(
		ctx,
		bson.M{"id": client.ID},
		bson.M{"$set": client},
	)
	return err
}

func (m *MongoDBStorage) DeleteClient(id string) error {
	ctx := context.Background()
	_, err := m.clients.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (m *MongoDBStorage) ValidateClient(clientID, clientSecret string) (*models.Client, error) {
	ctx := context.Background()
	var client models.Client
	err := m.clients.FindOne(ctx, bson.M{
		"id":     clientID,
		"secret": clientSecret,
	}).Decode(&client)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &client, err
}

// Authorization code operations
func (m *MongoDBStorage) CreateAuthorizationCode(code *models.AuthorizationCode) error {
	ctx := context.Background()
	code.CreatedAt = time.Now()
	_, err := m.codes.InsertOne(ctx, code)
	return err
}

func (m *MongoDBStorage) GetAuthorizationCode(code string) (*models.AuthorizationCode, error) {
	ctx := context.Background()
	var authCode models.AuthorizationCode
	err := m.codes.FindOne(ctx, bson.M{"code": code}).Decode(&authCode)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &authCode, err
}

func (m *MongoDBStorage) DeleteAuthorizationCode(code string) error {
	ctx := context.Background()
	_, err := m.codes.DeleteOne(ctx, bson.M{"code": code})
	return err
}

// Token operations
func (m *MongoDBStorage) CreateToken(token *models.Token) error {
	ctx := context.Background()
	token.CreatedAt = time.Now()
	_, err := m.tokens.InsertOne(ctx, token)
	return err
}

func (m *MongoDBStorage) GetTokenByAccessToken(accessToken string) (*models.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var token models.Token
	err := m.tokens.FindOne(ctx, bson.M{"access_token": accessToken}).Decode(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (m *MongoDBStorage) GetTokenByRefreshToken(refreshToken string) (*models.Token, error) {
	ctx := context.Background()
	var token models.Token
	err := m.tokens.FindOne(ctx, bson.M{"refresh_token": refreshToken}).Decode(&token)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &token, err
}

func (m *MongoDBStorage) DeleteToken(tokenID string) error {
	ctx := context.Background()
	_, err := m.tokens.DeleteOne(ctx, bson.M{"id": tokenID})
	return err
}

// Session operations
func (m *MongoDBStorage) CreateSession(session *models.Session) error {
	ctx := context.Background()
	session.CreatedAt = time.Now()
	_, err := m.sessions.InsertOne(ctx, session)
	return err
}

func (m *MongoDBStorage) GetSession(sessionID string) (*models.Session, error) {
	ctx := context.Background()
	var session models.Session
	err := m.sessions.FindOne(ctx, bson.M{"id": sessionID}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &session, err
}

func (m *MongoDBStorage) DeleteSession(sessionID string) error {
	ctx := context.Background()
	_, err := m.sessions.DeleteOne(ctx, bson.M{"id": sessionID})
	return err
}
