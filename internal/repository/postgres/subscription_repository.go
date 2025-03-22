package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainErrors "github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	if subscription.ID == uuid.Nil {
		subscription.ID = uuid.New()
	}

	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO subscriptions (
			id, user_id, product_id, status,
			start_date, end_date, trial_end_date, original_price,
			tax_amount, total_amount, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// Handle null values for trial_end_date
	var trialEndDate interface{} = nil
	if subscription.TrialEndDate != nil {
		trialEndDate = *subscription.TrialEndDate
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		subscription.ID,
		subscription.UserID,
		subscription.ProductID,
		subscription.Status,
		subscription.StartDate,
		subscription.EndDate,
		trialEndDate,
		subscription.OriginalPrice,
		subscription.TaxAmount,
		subscription.TotalAmount,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Create initial state change record to track the 'active' state
	stateChangeQuery := `
		INSERT INTO subscription_state_changes (
			id, subscription_id, previous_state, new_state,
			changed_at, reason
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(
		ctx,
		stateChangeQuery,
		uuid.New(),
		subscription.ID,
		"", // No previous state for a new subscription
		subscription.Status,
		now,
		"Subscription created",
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.product_id, s.status,
			s.start_date, s.end_date, s.trial_end_date, s.original_price,
			s.tax_amount, s.total_amount, s.created_at, s.updated_at,
			
			p.id, p.name, p.description, p.price, p.duration_months, 
			p.tax_rate, p.is_active, p.created_at, p.updated_at
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		WHERE s.id = $1
	`

	var subscription models.Subscription
	var product models.Product

	// Nullable fields
	var trialEndDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.ProductID,
		&subscription.Status,
		&subscription.StartDate,
		&subscription.EndDate,
		&trialEndDate,
		&subscription.OriginalPrice,
		&subscription.TaxAmount,
		&subscription.TotalAmount,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,

		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.DurationMonths,
		&product.TaxRate,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrSubscriptionNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if trialEndDate.Valid {
		subscription.TrialEndDate = &trialEndDate.Time
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.product_id, s.status,
			s.start_date, s.end_date, s.trial_end_date, s.original_price,
			s.tax_amount, s.total_amount, s.created_at, s.updated_at,
			
			p.id, p.name, p.description, p.price, p.duration_months, 
			p.tax_rate, p.is_active, p.created_at, p.updated_at
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		var subscription models.Subscription
		var product models.Product

		// Nullable fields
		var trialEndDate sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.ProductID,
			&subscription.Status,
			&subscription.StartDate,
			&subscription.EndDate,
			&trialEndDate,
			&subscription.OriginalPrice,
			&subscription.TaxAmount,
			&subscription.TotalAmount,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,

			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.DurationMonths,
			&product.TaxRate,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if trialEndDate.Valid {
			subscription.TrialEndDate = &trialEndDate.Time
		}

		// Add product relation
		subscription.Product = &product

		subscriptions = append(subscriptions, &subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	subscription.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions
		SET 
			status = $1, 
			start_date = $2, 
			end_date = $3, 
			trial_end_date = $4, 
			updated_at = $5
		WHERE id = $6
	`

	// Handle nullable trial_end_date
	var trialEndDate interface{} = nil
	if subscription.TrialEndDate != nil {
		trialEndDate = *subscription.TrialEndDate
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		subscription.Status,
		subscription.StartDate,
		subscription.EndDate,
		trialEndDate,
		subscription.UpdatedAt,
		subscription.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrSubscriptionNotFound
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First delete state changes (foreign key constraint)
	stateChangeQuery := `DELETE FROM subscription_state_changes WHERE subscription_id = $1`
	_, err = tx.ExecContext(ctx, stateChangeQuery, id)
	if err != nil {
		return err
	}

	// Then delete the subscription
	subscriptionQuery := `DELETE FROM subscriptions WHERE id = $1`
	result, err := tx.ExecContext(ctx, subscriptionQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrSubscriptionNotFound
	}

	return tx.Commit()
}

func (r *SubscriptionRepository) CreateStateChange(ctx context.Context, stateChange *models.SubscriptionStateChange) error {
	if stateChange.ID == uuid.Nil {
		stateChange.ID = uuid.New()
	}

	if stateChange.ChangedAt.IsZero() {
		stateChange.ChangedAt = time.Now()
	}

	query := `
		INSERT INTO subscription_state_changes (
			id, subscription_id, previous_state, new_state,
			changed_at, reason
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		stateChange.ID,
		stateChange.SubscriptionID,
		stateChange.PreviousState,
		stateChange.NewState,
		stateChange.ChangedAt,
		stateChange.Reason,
	)

	return err
}

func (r *SubscriptionRepository) GetStateChangesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.SubscriptionStateChange, error) {
	query := `
		SELECT 
			id, subscription_id, previous_state, new_state,
			changed_at, reason
		FROM subscription_state_changes
		WHERE subscription_id = $1
		ORDER BY changed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stateChanges []*models.SubscriptionStateChange

	for rows.Next() {
		stateChange := &models.SubscriptionStateChange{}
		err := rows.Scan(
			&stateChange.ID,
			&stateChange.SubscriptionID,
			&stateChange.PreviousState,
			&stateChange.NewState,
			&stateChange.ChangedAt,
			&stateChange.Reason,
		)

		if err != nil {
			return nil, err
		}

		stateChanges = append(stateChanges, stateChange)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stateChanges, nil
}
