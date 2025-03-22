package product

import (
	"context"
	"fmt"
	"strings"

	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo repository.ProductRepository
}

func NewService(repo repository.ProductRepository) *Service {
	return &Service{repo: repo}
}

type CreateProductInput struct {
	Name           string
	Description    string
	Price          decimal.Decimal
	DurationMonths int
	TaxRate        decimal.Decimal
	IsActive       bool
}

func (i *CreateProductInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if strings.TrimSpace(i.Name) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "name",
			Message: "must not be empty",
		})
	}

	if i.Price.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "price",
			Message: "must not be negative",
		})
	}

	if i.DurationMonths <= 0 {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "duration_months",
			Message: "must be greater than 0",
		})
	}

	if i.TaxRate.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "tax_rate",
			Message: "must not be negative",
		})
	}

	return validationErrors
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (*models.Product, error) {
	// Validate input
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	product := &models.Product{
		ID:             uuid.New(),
		Name:           input.Name,
		Description:    input.Description,
		Price:          input.Price,
		DurationMonths: input.DurationMonths,
		TaxRate:        input.TaxRate,
		IsActive:       input.IsActive,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *Service) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	return s.repo.GetAll(ctx)
}

type UpdateProductInput struct {
	ID             uuid.UUID
	Name           string
	Description    string
	Price          decimal.Decimal
	DurationMonths int
	TaxRate        decimal.Decimal
	IsActive       bool
}

// Validate validates the input for updating a product
func (i *UpdateProductInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if i.ID == uuid.Nil {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "must not be empty",
		})
	}

	if strings.TrimSpace(i.Name) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "name",
			Message: "must not be empty",
		})
	}

	if i.Price.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "price",
			Message: "must not be negative",
		})
	}

	if i.DurationMonths <= 0 {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "duration_months",
			Message: "must be greater than 0",
		})
	}

	if i.TaxRate.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "tax_rate",
			Message: "must not be negative",
		})
	}

	return validationErrors
}

func (s *Service) UpdateProduct(ctx context.Context, input UpdateProductInput) (*models.Product, error) {
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// Ensure product exists
	existingProduct, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	existingProduct.Name = input.Name
	existingProduct.Description = input.Description
	existingProduct.Price = input.Price
	existingProduct.DurationMonths = input.DurationMonths
	existingProduct.TaxRate = input.TaxRate
	existingProduct.IsActive = input.IsActive

	if err := s.repo.Update(ctx, existingProduct); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return existingProduct, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
