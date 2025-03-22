package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusPaused    SubscriptionStatus = "paused"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

type Subscription struct {
	ID              uuid.UUID          `json:"id"`
	UserID          uuid.UUID          `json:"user_id"`
	ProductID       uuid.UUID          `json:"product_id"`
	VoucherID       *uuid.UUID         `json:"voucher_id,omitempty"`
	Status          SubscriptionStatus `json:"status"`
	StartDate       time.Time          `json:"start_date"`
	EndDate         time.Time          `json:"end_date"`
	TrialEndDate    *time.Time         `json:"trial_end_date,omitempty"`
	OriginalPrice   decimal.Decimal    `json:"original_price"`
	DiscountedPrice *decimal.Decimal   `json:"discounted_price,omitempty"`
	TaxAmount       decimal.Decimal    `json:"tax_amount"`
	TotalAmount     decimal.Decimal    `json:"total_amount"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`

	// Relations (not stored in DB)
	Product *Product `json:"product,omitempty"`
	Voucher *Voucher `json:"voucher,omitempty"`
}

type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypePercentage DiscountType = "percentage"
)

type Voucher struct {
	ID            uuid.UUID       `json:"id"`
	Code          string          `json:"code"`
	DiscountType  DiscountType    `json:"discount_type"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"` // If null, applies to all products
	IsActive      bool            `json:"is_active"`
	ExpiresAt     time.Time       `json:"expires_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type SubscriptionStateChange struct {
	ID             uuid.UUID          `json:"id"`
	SubscriptionID uuid.UUID          `json:"subscription_id"`
	PreviousState  SubscriptionStatus `json:"previous_state"`
	NewState       SubscriptionStatus `json:"new_state"`
	ChangedAt      time.Time          `json:"changed_at"`
	Reason         string             `json:"reason"`
}
