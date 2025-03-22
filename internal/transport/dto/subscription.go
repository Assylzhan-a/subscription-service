package dto

import (
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/shopspring/decimal"
)

type CreateSubscriptionRequest struct {
	ProductID   string `json:"product_id" binding:"required,uuid"`
	VoucherCode string `json:"voucher_code"`
	WithTrial   bool   `json:"with_trial"`
}

type SubscriptionResponse struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	ProductID       string           `json:"product_id"`
	VoucherID       *string          `json:"voucher_id,omitempty"`
	Status          string           `json:"status"`
	StartDate       time.Time        `json:"start_date"`
	EndDate         time.Time        `json:"end_date"`
	TrialEndDate    *time.Time       `json:"trial_end_date,omitempty"`
	OriginalPrice   decimal.Decimal  `json:"original_price"`
	DiscountedPrice *decimal.Decimal `json:"discounted_price,omitempty"`
	TaxAmount       decimal.Decimal  `json:"tax_amount"`
	TotalAmount     decimal.Decimal  `json:"total_amount"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Product         *ProductResponse `json:"product,omitempty"`
	Voucher         *VoucherResponse `json:"voucher,omitempty"`
}

type SubscriptionStateChangeResponse struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	PreviousState  string    `json:"previous_state"`
	NewState       string    `json:"new_state"`
	ChangedAt      time.Time `json:"changed_at"`
	Reason         string    `json:"reason"`
}

func MapSubscriptionToResponse(subscription *models.Subscription) SubscriptionResponse {
	response := SubscriptionResponse{
		ID:            subscription.ID.String(),
		UserID:        subscription.UserID.String(),
		ProductID:     subscription.ProductID.String(),
		Status:        string(subscription.Status),
		StartDate:     subscription.StartDate,
		EndDate:       subscription.EndDate,
		OriginalPrice: subscription.OriginalPrice,
		TaxAmount:     subscription.TaxAmount,
		TotalAmount:   subscription.TotalAmount,
		CreatedAt:     subscription.CreatedAt,
		UpdatedAt:     subscription.UpdatedAt,
	}

	if subscription.VoucherID != nil {
		voucherID := subscription.VoucherID.String()
		response.VoucherID = &voucherID
	}

	if subscription.TrialEndDate != nil {
		response.TrialEndDate = subscription.TrialEndDate
	}

	if subscription.DiscountedPrice != nil {
		response.DiscountedPrice = subscription.DiscountedPrice
	}

	if subscription.Product != nil {
		product := MapProductToResponse(subscription.Product)
		response.Product = &product
	}

	if subscription.Voucher != nil {
		voucher := MapVoucherToResponse(subscription.Voucher)
		response.Voucher = &voucher
	}

	return response
}

func MapSubscriptionsToResponse(subscriptions []*models.Subscription) []SubscriptionResponse {
	responses := make([]SubscriptionResponse, len(subscriptions))
	for i, subscription := range subscriptions {
		responses[i] = MapSubscriptionToResponse(subscription)
	}
	return responses
}

func MapStateChangeToResponse(stateChange *models.SubscriptionStateChange) SubscriptionStateChangeResponse {
	return SubscriptionStateChangeResponse{
		ID:             stateChange.ID.String(),
		SubscriptionID: stateChange.SubscriptionID.String(),
		PreviousState:  string(stateChange.PreviousState),
		NewState:       string(stateChange.NewState),
		ChangedAt:      stateChange.ChangedAt,
		Reason:         stateChange.Reason,
	}
}

func MapStateChangesToResponse(stateChanges []*models.SubscriptionStateChange) []SubscriptionStateChangeResponse {
	responses := make([]SubscriptionStateChangeResponse, len(stateChanges))
	for i, stateChange := range stateChanges {
		responses[i] = MapStateChangeToResponse(stateChange)
	}
	return responses
}
