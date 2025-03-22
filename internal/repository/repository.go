package repository

import (
	"context"

	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
)

// UserRepository defines operations for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// ProductRepository defines operations for product persistence
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetAll(ctx context.Context) ([]*models.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubscriptionRepository defines operations for subscription persistence
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateStateChange(ctx context.Context, stateChange *models.SubscriptionStateChange) error
	GetStateChangesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.SubscriptionStateChange, error)
}
