package dto

import (
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/shopspring/decimal"
)

type CreateVoucherRequest struct {
	Code          string          `json:"code" binding:"required"`
	DiscountType  string          `json:"discount_type" binding:"required,oneof=fixed percentage"`
	DiscountValue decimal.Decimal `json:"discount_value" binding:"required"`
	ProductID     *string         `json:"product_id,omitempty" binding:"omitempty,uuid"`
	ExpiresAt     time.Time       `json:"expires_at" binding:"required"`
	IsActive      bool            `json:"is_active"`
}

type UpdateVoucherRequest struct {
	Code          string          `json:"code" binding:"required"`
	DiscountType  string          `json:"discount_type" binding:"required,oneof=fixed percentage"`
	DiscountValue decimal.Decimal `json:"discount_value" binding:"required"`
	ProductID     *string         `json:"product_id,omitempty" binding:"omitempty,uuid"`
	ExpiresAt     time.Time       `json:"expires_at" binding:"required"`
	IsActive      bool            `json:"is_active"`
}

type ValidateVoucherRequest struct {
	Code      string `json:"code" binding:"required"`
	ProductID string `json:"product_id" binding:"required,uuid"`
}

type VoucherResponse struct {
	ID            string          `json:"id"`
	Code          string          `json:"code"`
	DiscountType  string          `json:"discount_type"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	ProductID     *string         `json:"product_id,omitempty"`
	IsActive      bool            `json:"is_active"`
	ExpiresAt     time.Time       `json:"expires_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type ValidateVoucherResponse struct {
	Valid   bool             `json:"valid"`
	Voucher *VoucherResponse `json:"voucher,omitempty"`
	Error   string           `json:"error,omitempty"`
}

func MapVoucherToResponse(voucher *models.Voucher) VoucherResponse {
	response := VoucherResponse{
		ID:            voucher.ID.String(),
		Code:          voucher.Code,
		DiscountType:  string(voucher.DiscountType),
		DiscountValue: voucher.DiscountValue,
		IsActive:      voucher.IsActive,
		ExpiresAt:     voucher.ExpiresAt,
		CreatedAt:     voucher.CreatedAt,
		UpdatedAt:     voucher.UpdatedAt,
	}

	if voucher.ProductID != nil {
		productID := voucher.ProductID.String()
		response.ProductID = &productID
	}

	return response
}

func MapVouchersToResponse(vouchers []*models.Voucher) []VoucherResponse {
	responses := make([]VoucherResponse, len(vouchers))
	for i, voucher := range vouchers {
		responses[i] = MapVoucherToResponse(voucher)
	}
	return responses
}
