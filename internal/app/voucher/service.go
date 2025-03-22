package voucher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo        repository.VoucherRepository
	productRepo repository.ProductRepository
}

func NewService(
	repo repository.VoucherRepository,
	productRepo repository.ProductRepository,
) *Service {
	return &Service{
		repo:        repo,
		productRepo: productRepo,
	}
}

type CreateVoucherInput struct {
	Code          string
	DiscountType  models.DiscountType
	DiscountValue decimal.Decimal
	ProductID     *uuid.UUID
	IsActive      bool
	ExpiresAt     time.Time
}

func (i *CreateVoucherInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if strings.TrimSpace(i.Code) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "code",
			Message: "must not be empty",
		})
	}

	if i.DiscountType != models.DiscountTypeFixed && i.DiscountType != models.DiscountTypePercentage {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_type",
			Message: "must be either 'fixed' or 'percentage'",
		})
	}

	if i.DiscountValue.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_value",
			Message: "must not be negative",
		})
	}

	if i.DiscountType == models.DiscountTypePercentage && i.DiscountValue.GreaterThan(decimal.NewFromInt(100)) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_value",
			Message: "percentage cannot be greater than 100",
		})
	}

	if i.ExpiresAt.Before(time.Now()) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "expires_at",
			Message: "must be in the future",
		})
	}

	return validationErrors
}

func (s *Service) CreateVoucher(ctx context.Context, input CreateVoucherInput) (*models.Voucher, error) {
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// Check if product exists if productID is provided
	if input.ProductID != nil {
		_, err := s.productRepo.GetByID(ctx, *input.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}
	}

	voucher := &models.Voucher{
		ID:            uuid.New(),
		Code:          strings.ToUpper(input.Code),
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
		ProductID:     input.ProductID,
		IsActive:      input.IsActive,
		ExpiresAt:     input.ExpiresAt,
	}

	if err := s.repo.Create(ctx, voucher); err != nil {
		return nil, fmt.Errorf("failed to create voucher: %w", err)
	}

	return voucher, nil
}

func (s *Service) GetVoucherByID(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetVoucherByCode(ctx context.Context, code string) (*models.Voucher, error) {
	return s.repo.GetByCode(ctx, strings.ToUpper(code))
}

func (s *Service) GetVouchersByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Voucher, error) {
	return s.repo.GetByProductID(ctx, productID)
}

func (s *Service) GetAllActiveVouchers(ctx context.Context) ([]*models.Voucher, error) {
	return s.repo.GetAllActive(ctx)
}

type UpdateVoucherInput struct {
	ID            uuid.UUID
	Code          string
	DiscountType  models.DiscountType
	DiscountValue decimal.Decimal
	ProductID     *uuid.UUID
	IsActive      bool
	ExpiresAt     time.Time
}

func (i *UpdateVoucherInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if i.ID == uuid.Nil {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "must not be empty",
		})
	}

	if strings.TrimSpace(i.Code) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "code",
			Message: "must not be empty",
		})
	}

	if i.DiscountType != models.DiscountTypeFixed && i.DiscountType != models.DiscountTypePercentage {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_type",
			Message: "must be either 'fixed' or 'percentage'",
		})
	}

	if i.DiscountValue.IsNegative() {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_value",
			Message: "must not be negative",
		})
	}

	if i.DiscountType == models.DiscountTypePercentage && i.DiscountValue.GreaterThan(decimal.NewFromInt(100)) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "discount_value",
			Message: "percentage cannot be greater than 100",
		})
	}

	return validationErrors
}

func (s *Service) UpdateVoucher(ctx context.Context, input UpdateVoucherInput) (*models.Voucher, error) {
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// Ensure voucher exists
	existingVoucher, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Check if product exists if productID is provided
	if input.ProductID != nil {
		_, err := s.productRepo.GetByID(ctx, *input.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}
	}

	// Update fields
	existingVoucher.Code = strings.ToUpper(input.Code)
	existingVoucher.DiscountType = input.DiscountType
	existingVoucher.DiscountValue = input.DiscountValue
	existingVoucher.ProductID = input.ProductID
	existingVoucher.IsActive = input.IsActive
	existingVoucher.ExpiresAt = input.ExpiresAt

	if err := s.repo.Update(ctx, existingVoucher); err != nil {
		return nil, fmt.Errorf("failed to update voucher: %w", err)
	}

	return existingVoucher, nil
}

func (s *Service) DeleteVoucher(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

type ValidateVoucherInput struct {
	Code      string
	ProductID uuid.UUID
}

func (s *Service) ValidateVoucher(ctx context.Context, input ValidateVoucherInput) (*models.Voucher, error) {
	voucher, err := s.repo.GetByCode(ctx, strings.ToUpper(input.Code))
	if err != nil {
		return nil, err
	}

	// Check if voucher is active
	if !voucher.IsActive {
		return nil, errors.ErrVoucherInactive
	}

	// Check if voucher is expired
	if time.Now().After(voucher.ExpiresAt) {
		return nil, errors.ErrVoucherExpired
	}

	// Check if voucher is applicable to this product
	if voucher.ProductID != nil && *voucher.ProductID != input.ProductID {
		return nil, errors.ErrVoucherInvalid
	}

	return voucher, nil
}
