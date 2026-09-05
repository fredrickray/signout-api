package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/signout/signout-api/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, q,
		user.ID,
		strings.ToLower(strings.TrimSpace(user.Email)),
		user.PasswordHash,
		strings.TrimSpace(user.FullName),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`
	return r.scanOne(ctx, q, strings.TrimSpace(email))
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`
	return r.scanOne(ctx, q, id)
}

func (r *UserRepository) scanOne(ctx context.Context, q string, arg any) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, q, arg).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		LIMIT 1
	`
	var t domain.RefreshToken
	err := r.pool.QueryRow(ctx, q, hash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
		&t.RevokedAt,
		&t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("revoke user refresh tokens: %w", err)
	}
	return nil
}
