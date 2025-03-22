package dto

import (
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/shopspring/decimal"
)

type CreateProductRequest struct {
	Name           string          `json:"name" binding:"required"`
	Description    string          `json:"description"`
	Price          decimal.Decimal `json:"price" binding:"required"`
	DurationMonths int             `json:"duration_months" binding:"required,min=1"`
	TaxRate        decimal.Decimal `json:"tax_rate" binding:"required"`
	IsActive       bool            `json:"is_active"`
}

type UpdateProductRequest struct {
	Name           string          `json:"name" binding:"required"`
	Description    string          `json:"description"`
	Price          decimal.Decimal `json:"price" binding:"required"`
	DurationMonths int             `json:"duration_months" binding:"required,min=1"`
	TaxRate        decimal.Decimal `json:"tax_rate" binding:"required"`
	IsActive       bool            `json:"is_active"`
}

type ProductResponse struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Price          decimal.Decimal `json:"price"`
	DurationMonths int             `json:"duration_months"`
	TaxRate        decimal.Decimal `json:"tax_rate"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func MapProductToResponse(product *models.Product) ProductResponse {
	return ProductResponse{
		ID:             product.ID.String(),
		Name:           product.Name,
		Description:    product.Description,
		Price:          product.Price,
		DurationMonths: product.DurationMonths,
		TaxRate:        product.TaxRate,
		IsActive:       product.IsActive,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	}
}

func MapProductsToResponse(products []*models.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = MapProductToResponse(product)
	}
	return responses
}
