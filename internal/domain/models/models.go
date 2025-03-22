package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// User represents a service user
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product represents a subscription product
type Product struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Price          decimal.Decimal `json:"price"` // Using decimal for currency
	DurationMonths int             `json:"duration_months"`
	TaxRate        decimal.Decimal `json:"tax_rate"` // Using decimal for tax rate
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// SubscriptionStatus represents the current state of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusPaused    SubscriptionStatus = "paused"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

// Subscription represents a user's subscription to a product
type Subscription struct {
	ID            uuid.UUID          `json:"id"`
	UserID        uuid.UUID          `json:"user_id"`
	ProductID     uuid.UUID          `json:"product_id"`
	Status        SubscriptionStatus `json:"status"`
	StartDate     time.Time          `json:"start_date"`
	EndDate       time.Time          `json:"end_date"`
	TrialEndDate  *time.Time         `json:"trial_end_date,omitempty"`
	OriginalPrice decimal.Decimal    `json:"original_price"`
	TaxAmount     decimal.Decimal    `json:"tax_amount"`
	TotalAmount   decimal.Decimal    `json:"total_amount"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}
