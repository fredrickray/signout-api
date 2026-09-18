package celebration

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/signout/signout-api/internal/domain"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type ShirtStore interface {
	ListActive(ctx context.Context) ([]domain.ShirtTemplate, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ShirtTemplate, error)
	SeedDefaults(ctx context.Context) error
}

type CelebrationStore interface {
	Create(ctx context.Context, c *domain.Celebration) error
	FindBySlug(ctx context.Context, slug string) (*domain.Celebration, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Celebration, error)
}

type Service struct {
	shirts       ShirtStore
	celebrations CelebrationStore
	now          func() time.Time
}

func NewService(shirts ShirtStore, celebrations CelebrationStore) *Service {
	return &Service{
		shirts:       shirts,
		celebrations: celebrations,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

type CreateInput struct {
	UserID          uuid.UUID
	Slug            string
	DisplayName     string
	School          string
	Faculty         string
	ClassOf         string
	CelebrationType string
	ShirtTemplateID uuid.UUID
}

func (s *Service) ListShirts(ctx context.Context) ([]domain.ShirtTemplate, error) {
	return s.shirts.ListActive(ctx)
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.CelebrationPublic, error) {
	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if !slugRe.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must be lowercase letters, numbers, and hyphens", domain.ErrValidation)
	}
	if in.CelebrationType != "graduation" && in.CelebrationType != "nysc" {
		return nil, fmt.Errorf("%w: celebration_type must be graduation or nysc", domain.ErrValidation)
	}

	shirt, err := s.shirts.FindByID(ctx, in.ShirtTemplateID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	c := &domain.Celebration{
		ID:              uuid.New(),
		UserID:          in.UserID,
		Slug:            slug,
		DisplayName:     strings.TrimSpace(in.DisplayName),
		School:          strings.TrimSpace(in.School),
		Faculty:         strings.TrimSpace(in.Faculty),
		ClassOf:         strings.TrimSpace(in.ClassOf),
		CelebrationType: in.CelebrationType,
		ShirtTemplateID: shirt.ID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.celebrations.Create(ctx, c); err != nil {
		return nil, err
	}
	return &domain.CelebrationPublic{Celebration: *c, Shirt: shirt}, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.CelebrationPublic, error) {
	c, err := s.celebrations.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	shirt, err := s.shirts.FindByID(ctx, c.ShirtTemplateID)
	if err != nil {
		return nil, err
	}
	return &domain.CelebrationPublic{Celebration: *c, Shirt: shirt}, nil
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID) ([]domain.CelebrationPublic, error) {
	list, err := s.celebrations.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CelebrationPublic, 0, len(list))
	for _, c := range list {
		shirt, err := s.shirts.FindByID(ctx, c.ShirtTemplateID)
		if err != nil {
			continue
		}
		out = append(out, domain.CelebrationPublic{Celebration: c, Shirt: shirt})
	}
	return out, nil
}
