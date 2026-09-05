package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/signout/signout-api/internal/domain"
)

type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type RefreshStore interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	users    UserStore
	refresh  RefreshStore
	tokens   *TokenManager
	now      func() time.Time
}

func NewService(users UserStore, refresh RefreshStore, tokens *TokenManager) *Service {
	return &Service{
		users:   users,
		refresh: refresh,
		tokens:  tokens,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	User   *domain.User `json:"user"`
	Tokens TokenPair    `json:"tokens"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	fullName := strings.TrimSpace(in.FullName)

	hash, err := HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	now := s.now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, domain.ErrConflict
		}
		return nil, err
	}

	pair, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, Tokens: *pair}, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !CheckPassword(user.PasswordHash, in.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	pair, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, Tokens: *pair}, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (*TokenPair, error) {
	hash := HashRefreshToken(strings.TrimSpace(rawRefresh))
	stored, err := s.refresh.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrRefreshTokenInvalid
		}
		return nil, err
	}
	if !stored.IsActive(s.now()) {
		return nil, domain.ErrRefreshTokenInvalid
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	// Rotate refresh token
	if err := s.refresh.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}
	return s.issueSession(ctx, user)
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return nil
	}
	hash := HashRefreshToken(rawRefresh)
	stored, err := s.refresh.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.refresh.Revoke(ctx, stored.ID)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) issueSession(ctx context.Context, user *domain.User) (*TokenPair, error) {
	access, _, err := s.tokens.IssueAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(s.tokens.RefreshTTL()),
		CreatedAt: now,
	}
	if err := s.refresh.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("persist refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.tokens.AccessTTL().Seconds()),
		TokenType:    "Bearer",
	}, nil
}
