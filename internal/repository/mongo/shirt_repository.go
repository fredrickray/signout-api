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
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/signout/signout-api/internal/domain"
)

type shirtDoc struct {
	ID           string    `bson:"_id"`
	Slug         string    `bson:"slug"`
	Name         string    `bson:"name"`
	Description  string    `bson:"description"`
	Category     string    `bson:"category"`
	ModelURL     string    `bson:"model_url"`
	PreviewImage string    `bson:"preview_image"`
	Active       bool      `bson:"active"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type ShirtRepository struct {
	col *mongo.Collection
}

func NewShirtRepository(db *mongo.Database) *ShirtRepository {
	return &ShirtRepository{col: db.Collection("shirt_templates")}
}

func (r *ShirtRepository) SeedDefaults(ctx context.Context) error {
	defaults := []shirtDoc{
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111101").String(),
			Slug:         "classic-tee",
			Name:         "Classic Tee",
			Description:  "Short-sleeve henley — clean canvas for signatures.",
			Category:     "graduation",
			ModelURL:     "/models/shirtie.glb",
			PreviewImage: "/images/preview-classic-tee.png",
			Active:       true,
		},
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111102").String(),
			Slug:         "rolled-sleeves",
			Name:         "Rolled Sleeves",
			Description:  "Button-up with rolled sleeves for a casual sign-out look.",
			Category:     "graduation",
			ModelURL:     "/models/rolled-sleeves.glb",
			PreviewImage: "/images/preview-rolled-sleeves.jpg",
			Active:       true,
		},
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111103").String(),
			Slug:         "campus-fit",
			Name:         "Campus Fit",
			Description:  "Draped campus shirt — bold surface for graduation or POP.",
			Category:     "nysc",
			ModelURL:     "/models/shirt-clo.glb",
			PreviewImage: "/images/preview-campus-fit.png",
			Active:       true,
		},
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111104").String(),
			Slug:         "mens-shirt",
			Name:         "Men's Shirt",
			Description:  "Classic long-sleeve dress shirt with collar and cuffs.",
			Category:     "graduation",
			ModelURL:     "/models/mens-shirt.glb",
			PreviewImage: "/images/preview-mens-shirt.png",
			Active:       true,
		},
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111105").String(),
			Slug:         "hood-down",
			Name:         "Hood Down",
			Description:  "Hoodie with the hood down — sleeves and chest ready to sign.",
			Category:     "graduation",
			ModelURL:     "/models/hood-down.glb",
			PreviewImage: "/images/preview-hood-down.png",
			Active:       true,
		},
	}

	now := time.Now().UTC()
	for _, d := range defaults {
		d.CreatedAt = now
		d.UpdatedAt = now
		_, err := r.col.UpdateOne(
			ctx,
			bson.M{"_id": d.ID},
			bson.M{
				"$set": bson.M{
					"slug":          d.Slug,
					"name":          d.Name,
					"description":   d.Description,
					"category":      d.Category,
					"model_url":     d.ModelURL,
					"preview_image": d.PreviewImage,
					"active":        d.Active,
					"updated_at":    d.UpdatedAt,
				},
				"$setOnInsert": bson.M{
					"_id":        d.ID,
					"created_at": d.CreatedAt,
				},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("seed shirt %s: %w", d.Slug, err)
		}
	}
	return nil
}

func (r *ShirtRepository) ListActive(ctx context.Context) ([]domain.ShirtTemplate, error) {
	cur, err := r.col.Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list shirts: %w", err)
	}
	defer cur.Close(ctx)

	var out []domain.ShirtTemplate
	for cur.Next(ctx) {
		var d shirtDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		t, err := d.toDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, cur.Err()
}

func (r *ShirtRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ShirtTemplate, error) {
	var d shirtDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id.String(), "active": true}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find shirt: %w", err)
	}
	return d.toDomain()
}

func (d shirtDoc) toDomain() (*domain.ShirtTemplate, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, err
	}
	return &domain.ShirtTemplate{
		ID:           id,
		Slug:         d.Slug,
		Name:         d.Name,
		Description:  d.Description,
		Category:     d.Category,
		ModelURL:     d.ModelURL,
		PreviewImage: d.PreviewImage,
		Active:       d.Active,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}, nil
}

type celebrationDoc struct {
	ID              string    `bson:"_id"`
	UserID          string    `bson:"user_id"`
	Slug            string    `bson:"slug"`
	DisplayName     string    `bson:"display_name"`
	School          string    `bson:"school"`
	Faculty         string    `bson:"faculty"`
	ClassOf         string    `bson:"class_of"`
	CelebrationType string    `bson:"celebration_type"`
	ShirtTemplateID string    `bson:"shirt_template_id"`
	CreatedAt       time.Time `bson:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
}

type CelebrationRepository struct {
	col *mongo.Collection
}

func NewCelebrationRepository(db *mongo.Database) *CelebrationRepository {
	return &CelebrationRepository{col: db.Collection("celebrations")}
}

func (r *CelebrationRepository) Create(ctx context.Context, c *domain.Celebration) error {
	doc := celebrationDoc{
		ID:              c.ID.String(),
		UserID:          c.UserID.String(),
		Slug:            strings.ToLower(strings.TrimSpace(c.Slug)),
		DisplayName:     strings.TrimSpace(c.DisplayName),
		School:          strings.TrimSpace(c.School),
		Faculty:         strings.TrimSpace(c.Faculty),
		ClassOf:         strings.TrimSpace(c.ClassOf),
		CelebrationType: c.CelebrationType,
		ShirtTemplateID: c.ShirtTemplateID.String(),
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
	_, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("create celebration: %w", err)
	}
	return nil
}

func (r *CelebrationRepository) FindBySlug(ctx context.Context, slug string) (*domain.Celebration, error) {
	var d celebrationDoc
	err := r.col.FindOne(ctx, bson.M{"slug": strings.ToLower(strings.TrimSpace(slug))}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find celebration: %w", err)
	}
	return d.toDomain()
}

func (r *CelebrationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Celebration, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID.String()}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list celebrations: %w", err)
	}
	defer cur.Close(ctx)

	var out []domain.Celebration
	for cur.Next(ctx) {
		var d celebrationDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		c, err := d.toDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, cur.Err()
}

func (d celebrationDoc) toDomain() (*domain.Celebration, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, err
	}
	shirtID, err := uuid.Parse(d.ShirtTemplateID)
	if err != nil {
		return nil, err
	}
	return &domain.Celebration{
		ID:              id,
		UserID:          userID,
		Slug:            d.Slug,
		DisplayName:     d.DisplayName,
		School:          d.School,
		Faculty:         d.Faculty,
		ClassOf:         d.ClassOf,
		CelebrationType: d.CelebrationType,
		ShirtTemplateID: shirtID,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}, nil
}
