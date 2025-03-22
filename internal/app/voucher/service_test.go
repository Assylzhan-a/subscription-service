package voucher_test

import (
	"context"
	"testing"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/app/voucher"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockVoucherRepository struct {
	vouchers map[uuid.UUID]*models.Voucher
	codes    map[string]*models.Voucher
}

func newMockVoucherRepository() *mockVoucherRepository {
	return &mockVoucherRepository{
		vouchers: make(map[uuid.UUID]*models.Voucher),
		codes:    make(map[string]*models.Voucher),
	}
}

func (m *mockVoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	m.vouchers[voucher.ID] = voucher
	m.codes[voucher.Code] = voucher
	return nil
}

func (m *mockVoucherRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	if voucher, ok := m.vouchers[id]; ok {
		return voucher, nil
	}
	return nil, errors.ErrVoucherNotFound
}

func (m *mockVoucherRepository) GetByCode(ctx context.Context, code string) (*models.Voucher, error) {
	if voucher, ok := m.codes[code]; ok {
		return voucher, nil
	}
	return nil, errors.ErrVoucherNotFound
}

func (m *mockVoucherRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Voucher, error) {
	var result []*models.Voucher
	for _, v := range m.vouchers {
		if v.ProductID != nil && *v.ProductID == productID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVoucherRepository) GetAllActive(ctx context.Context) ([]*models.Voucher, error) {
	var result []*models.Voucher
	for _, v := range m.vouchers {
		if v.IsActive && v.ExpiresAt.After(time.Now()) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	if _, ok := m.vouchers[voucher.ID]; !ok {
		return errors.ErrVoucherNotFound
	}

	// Update code map as well if code changed
	oldVoucher := m.vouchers[voucher.ID]
	if oldVoucher.Code != voucher.Code {
		delete(m.codes, oldVoucher.Code)
		m.codes[voucher.Code] = voucher
	} else {
		m.codes[voucher.Code] = voucher
	}

	m.vouchers[voucher.ID] = voucher
	return nil
}

func (m *mockVoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if voucher, ok := m.vouchers[id]; ok {
		delete(m.vouchers, id)
		delete(m.codes, voucher.Code)
		return nil
	}
	return errors.ErrVoucherNotFound
}

type mockProductRepository struct {
	products map[uuid.UUID]*models.Product
}

func newMockProductRepository() *mockProductRepository {
	return &mockProductRepository{
		products: make(map[uuid.UUID]*models.Product),
	}
}

func (m *mockProductRepository) Create(ctx context.Context, product *models.Product) error {
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) GetAll(ctx context.Context) ([]*models.Product, error) {
	products := make([]*models.Product, 0, len(m.products))
	for _, p := range m.products {
		products = append(products, p)
	}
	return products, nil
}

func (m *mockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if product, ok := m.products[id]; ok {
		return product, nil
	}
	return nil, errors.ErrProductNotFound
}

func (m *mockProductRepository) Update(ctx context.Context, product *models.Product) error {
	if _, ok := m.products[product.ID]; !ok {
		return errors.ErrProductNotFound
	}
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.products[id]; !ok {
		return errors.ErrProductNotFound
	}
	delete(m.products, id)
	return nil
}

// Helper function to create a test product
func createTestProduct() *models.Product {
	return &models.Product{
		ID:             uuid.New(),
		Name:           "Test Product",
		Description:    "Test Description",
		Price:          decimal.NewFromFloat(19.99),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20), // 20% tax
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Tests for VoucherService
func TestCreateVoucher(t *testing.T) {
	// Setup
	ctx := context.Background()
	voucherRepo := newMockVoucherRepository()
	productRepo := newMockProductRepository()
	service := voucher.NewService(voucherRepo, productRepo)

	// Create a test product
	product := createTestProduct()
	err := productRepo.Create(ctx, product)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Test case 1: Create percentage voucher without product restriction
	input := voucher.CreateVoucherInput{
		Code:          "PERCENT25",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(25), // 25% discount
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0), // 1 month expiry
	}

	v, err := service.CreateVoucher(ctx, input)
	if err != nil {
		t.Fatal("Failed to create percentage voucher:", err)
	}

	if v.Code != "PERCENT25" {
		t.Errorf("Expected code PERCENT25, got %v", v.Code)
	}

	if v.DiscountType != models.DiscountTypePercentage {
		t.Errorf("Expected discount type %v, got %v", models.DiscountTypePercentage, v.DiscountType)
	}

	if !v.DiscountValue.Equal(decimal.NewFromInt(25)) {
		t.Errorf("Expected discount value 25, got %v", v.DiscountValue)
	}

	if v.ProductID != nil {
		t.Error("Expected ProductID to be nil")
	}

	// Test case 2: Create fixed voucher with product restriction
	productID := product.ID
	input = voucher.CreateVoucherInput{
		Code:          "FIXED10",
		DiscountType:  models.DiscountTypeFixed,
		DiscountValue: decimal.NewFromInt(10), // $10 discount
		ProductID:     &productID,
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0), // 1 month expiry
	}

	v, err = service.CreateVoucher(ctx, input)
	if err != nil {
		t.Fatal("Failed to create fixed voucher:", err)
	}

	if v.Code != "FIXED10" {
		t.Errorf("Expected code FIXED10, got %v", v.Code)
	}

	if v.DiscountType != models.DiscountTypeFixed {
		t.Errorf("Expected discount type %v, got %v", models.DiscountTypeFixed, v.DiscountType)
	}

	if !v.DiscountValue.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected discount value 10, got %v", v.DiscountValue)
	}

	if v.ProductID == nil {
		t.Error("Expected ProductID not to be nil")
	} else if *v.ProductID != productID {
		t.Errorf("Expected ProductID %v, got %v", productID, *v.ProductID)
	}

	// Test case 3: Create voucher with invalid discount value (negative)
	input = voucher.CreateVoucherInput{
		Code:          "INVALID",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(-10), // Invalid negative value
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0),
	}

	_, err = service.CreateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for negative discount value")
	}

	// Test case 4: Create voucher with invalid percentage (over 100%)
	input = voucher.CreateVoucherInput{
		Code:          "INVALID",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(120), // Invalid percentage
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0),
	}

	_, err = service.CreateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for percentage over 100%")
	}

	// Test case 5: Create voucher with past expiry date
	input = voucher.CreateVoucherInput{
		Code:          "INVALID",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(50),
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 0, -1), // Expired yesterday
	}

	_, err = service.CreateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for past expiry date")
	}
}

func TestValidateVoucher(t *testing.T) {
	// Setup
	ctx := context.Background()
	voucherRepo := newMockVoucherRepository()
	productRepo := newMockProductRepository()
	service := voucher.NewService(voucherRepo, productRepo)

	// Create a test product
	product := createTestProduct()
	err := productRepo.Create(ctx, product)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Create another test product
	product2 := createTestProduct()
	product2.ID = uuid.New() // Ensure different ID
	err = productRepo.Create(ctx, product2)
	if err != nil {
		t.Fatal("Failed to create second test product:", err)
	}

	// Create vouchers for testing
	// 1. Global voucher (works for any product)
	globalVoucher := &models.Voucher{
		ID:            uuid.New(),
		Code:          "GLOBAL20",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(20),
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = voucherRepo.Create(ctx, globalVoucher)
	if err != nil {
		t.Fatal("Failed to create global voucher:", err)
	}

	// 2. Product-specific voucher
	productID := product.ID
	productVoucher := &models.Voucher{
		ID:            uuid.New(),
		Code:          "PRODUCT10",
		DiscountType:  models.DiscountTypeFixed,
		DiscountValue: decimal.NewFromInt(10),
		ProductID:     &productID,
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = voucherRepo.Create(ctx, productVoucher)
	if err != nil {
		t.Fatal("Failed to create product voucher:", err)
	}

	// 3. Inactive voucher
	inactiveVoucher := &models.Voucher{
		ID:            uuid.New(),
		Code:          "INACTIVE",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(30),
		IsActive:      false,
		ExpiresAt:     time.Now().AddDate(0, 1, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = voucherRepo.Create(ctx, inactiveVoucher)
	if err != nil {
		t.Fatal("Failed to create inactive voucher:", err)
	}

	// 4. Expired voucher
	expiredVoucher := &models.Voucher{
		ID:            uuid.New(),
		Code:          "EXPIRED",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(40),
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 0, -1), // Expired yesterday
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = voucherRepo.Create(ctx, expiredVoucher)
	if err != nil {
		t.Fatal("Failed to create expired voucher:", err)
	}

	// Test case 1: Validate global voucher with product1
	input := voucher.ValidateVoucherInput{
		Code:      "GLOBAL20",
		ProductID: product.ID,
	}

	v, err := service.ValidateVoucher(ctx, input)
	if err != nil {
		t.Fatal("Failed to validate global voucher:", err)
	}

	if v.Code != "GLOBAL20" {
		t.Errorf("Expected code GLOBAL20, got %v", v.Code)
	}

	// Test case 2: Validate product voucher with matching product
	input = voucher.ValidateVoucherInput{
		Code:      "PRODUCT10",
		ProductID: product.ID,
	}

	v, err = service.ValidateVoucher(ctx, input)
	if err != nil {
		t.Fatal("Failed to validate product voucher:", err)
	}

	if v.Code != "PRODUCT10" {
		t.Errorf("Expected code PRODUCT10, got %v", v.Code)
	}

	// Test case 3: Validate product voucher with non-matching product
	input = voucher.ValidateVoucherInput{
		Code:      "PRODUCT10",
		ProductID: product2.ID,
	}

	_, err = service.ValidateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for product voucher with non-matching product")
	}

	// Test case 4: Validate inactive voucher
	input = voucher.ValidateVoucherInput{
		Code:      "INACTIVE",
		ProductID: product.ID,
	}

	_, err = service.ValidateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for inactive voucher")
	}
	if err != errors.ErrVoucherInactive {
		t.Errorf("Expected error %v, got %v", errors.ErrVoucherInactive, err)
	}

	// Test case 5: Validate expired voucher
	input = voucher.ValidateVoucherInput{
		Code:      "EXPIRED",
		ProductID: product.ID,
	}

	_, err = service.ValidateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for expired voucher")
	}
	if err != errors.ErrVoucherExpired {
		t.Errorf("Expected error %v, got %v", errors.ErrVoucherExpired, err)
	}

	// Test case 6: Validate non-existent voucher
	input = voucher.ValidateVoucherInput{
		Code:      "NONEXISTENT",
		ProductID: product.ID,
	}

	_, err = service.ValidateVoucher(ctx, input)
	if err == nil {
		t.Error("Expected error for non-existent voucher")
	}
	if err != errors.ErrVoucherNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrVoucherNotFound, err)
	}
}
