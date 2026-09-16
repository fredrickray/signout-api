package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/signout/signout-api/internal/domain"
)

type userDoc struct {
	ID           string    `bson:"_id"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	FullName     string    `bson:"full_name"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{col: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	doc := userDoc{
		ID:           user.ID.String(),
		Email:        strings.ToLower(strings.TrimSpace(user.Email)),
		PasswordHash: user.PasswordHash,
		FullName:     strings.TrimSpace(user.FullName),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	_, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var doc userDoc
	err := r.col.FindOne(ctx, bson.M{"email": strings.ToLower(strings.TrimSpace(email))}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return doc.toDomain()
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var doc userDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return doc.toDomain()
}

func (d userDoc) toDomain() (*domain.User, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	return &domain.User{
		ID:           id,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		FullName:     d.FullName,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}, nil
}

type refreshDoc struct {
	ID        string     `bson:"_id"`
	UserID    string     `bson:"user_id"`
	TokenHash string     `bson:"token_hash"`
	ExpiresAt time.Time  `bson:"expires_at"`
	RevokedAt *time.Time `bson:"revoked_at,omitempty"`
	CreatedAt time.Time  `bson:"created_at"`
}

type RefreshTokenRepository struct {
	col *mongo.Collection
}

func NewRefreshTokenRepository(db *mongo.Database) *RefreshTokenRepository {
	return &RefreshTokenRepository{col: db.Collection("refresh_tokens")}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	doc := refreshDoc{
		ID:        token.ID.String(),
		UserID:    token.UserID.String(),
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
	_, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var doc refreshDoc
	err := r.col.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return doc.toDomain()
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id.String(), "revoked_at": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID.String(), "revoked_at": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	if err != nil {
		return fmt.Errorf("revoke user refresh tokens: %w", err)
	}
	return nil
}

func (d refreshDoc) toDomain() (*domain.RefreshToken, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("parse refresh id: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse refresh user id: %w", err)
	}
	return &domain.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: d.TokenHash,
		ExpiresAt: d.ExpiresAt,
		RevokedAt: d.RevokedAt,
		CreatedAt: d.CreatedAt,
	}, nil
}
