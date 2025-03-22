package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/internal/repository"
	"github.com/google/uuid"
)

type Service struct {
	repo        repository.SubscriptionRepository
	productRepo repository.ProductRepository
}

func NewService(
	repo repository.SubscriptionRepository,
	productRepo repository.ProductRepository,
) *Service {
	return &Service{
		repo:        repo,
		productRepo: productRepo,
	}
}

type CreateSubscriptionInput struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	WithTrial bool
}

func (i *CreateSubscriptionInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if i.UserID == uuid.Nil {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "user_id",
			Message: "must not be empty",
		})
	}

	if i.ProductID == uuid.Nil {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "product_id",
			Message: "must not be empty",
		})
	}

	return validationErrors
}

func (s *Service) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error) {
	// Validate input
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// Get product
	product, err := s.productRepo.GetByID(ctx, input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check if product is active
	if !product.IsActive {
		return nil, errors.ErrInactiveProduct
	}

	// Calculate pricing and dates
	startDate := time.Now()
	endDate := startDate.AddDate(0, product.DurationMonths, 0)
	var trialEndDate *time.Time

	// Handle trial period if requested
	if input.WithTrial {
		trialEnd := startDate.AddDate(0, 1, 0) // 1 month trial
		trialEndDate = &trialEnd

		// Start date is after trial period
		startDate = trialEnd
		endDate = trialEnd.AddDate(0, product.DurationMonths, 0)
	}

	// Create subscription object
	subscription := &models.Subscription{
		ID:            uuid.New(),
		UserID:        input.UserID,
		ProductID:     input.ProductID,
		Status:        models.SubscriptionStatusActive,
		StartDate:     startDate,
		EndDate:       endDate,
		TrialEndDate:  trialEndDate,
		OriginalPrice: product.Price,
		TaxAmount:     product.Price.Mul(product.TaxRate),
		TotalAmount:   product.Price.Add(product.Price.Mul(product.TaxRate)),
	}

	// Save subscription
	if err := s.repo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Set product relationship for the response
	subscription.Product = product

	return subscription, nil
}

func (s *Service) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) PauseSubscription(ctx context.Context, id uuid.UUID) error {
	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the subscription is active
	if subscription.Status != models.SubscriptionStatusActive {
		return errors.ErrSubscriptionNotActive
	}

	// Check if subscription is in trial period
	if subscription.TrialEndDate != nil && time.Now().Before(*subscription.TrialEndDate) {
		return errors.ErrSubscriptionInTrial
	}

	// Create state change record
	previousState := subscription.Status
	subscription.Status = models.SubscriptionStatusPaused

	stateChange := &models.SubscriptionStateChange{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		PreviousState:  previousState,
		NewState:       subscription.Status,
		ChangedAt:      time.Now(),
		Reason:         "User requested pause",
	}

	// Update subscription
	if err := s.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log state change
	if err := s.repo.CreateStateChange(ctx, stateChange); err != nil {
		return fmt.Errorf("failed to log state change: %w", err)
	}

	return nil
}

func (s *Service) UnpauseSubscription(ctx context.Context, id uuid.UUID) error {
	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the subscription is paused
	if subscription.Status != models.SubscriptionStatusPaused {
		return errors.ErrSubscriptionAlreadyPaused
	}

	// Create state change record
	previousState := subscription.Status
	subscription.Status = models.SubscriptionStatusActive

	stateChange := &models.SubscriptionStateChange{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		PreviousState:  previousState,
		NewState:       subscription.Status,
		ChangedAt:      time.Now(),
		Reason:         "User requested unpause",
	}

	// Update subscription
	if err := s.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log state change
	if err := s.repo.CreateStateChange(ctx, stateChange); err != nil {
		return fmt.Errorf("failed to log state change: %w", err)
	}

	return nil
}

func (s *Service) CancelSubscription(ctx context.Context, id uuid.UUID) error {
	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Can cancel from any state
	if subscription.Status == models.SubscriptionStatusCancelled {
		return nil // Already cancelled, nothing to do
	}

	// Create state change record
	previousState := subscription.Status
	subscription.Status = models.SubscriptionStatusCancelled

	stateChange := &models.SubscriptionStateChange{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		PreviousState:  previousState,
		NewState:       subscription.Status,
		ChangedAt:      time.Now(),
		Reason:         "User requested cancellation",
	}

	// Update subscription
	if err := s.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log state change
	if err := s.repo.CreateStateChange(ctx, stateChange); err != nil {
		return fmt.Errorf("failed to log state change: %w", err)
	}

	return nil
}
