package repository

import (
	"context"

	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
)

// ProductRepository defines operations for product persistence
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetAll(ctx context.Context) ([]*models.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines operations for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}
