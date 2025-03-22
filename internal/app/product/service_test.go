package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/app/product"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

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

func TestCreateProduct(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := newMockProductRepository()
	service := product.NewService(repo)

	// Test case 1: Create valid product
	input := product.CreateProductInput{
		Name:           "Premium Plan",
		Description:    "Our premium subscription plan",
		Price:          decimal.NewFromFloat(29.99),
		DurationMonths: 3,
		TaxRate:        decimal.NewFromFloat(0.20), // 20% tax
		IsActive:       true,
	}

	p, err := service.CreateProduct(ctx, input)
	if err != nil {
		t.Fatal("Failed to create product:", err)
	}

	if p.Name != "Premium Plan" {
		t.Errorf("Expected name Premium Plan, got %v", p.Name)
	}

	if p.Description != "Our premium subscription plan" {
		t.Errorf("Expected description 'Our premium subscription plan', got %v", p.Description)
	}

	if !p.Price.Equal(decimal.NewFromFloat(29.99)) {
		t.Errorf("Expected price 29.99, got %v", p.Price)
	}

	if p.DurationMonths != 3 {
		t.Errorf("Expected duration 3 months, got %v", p.DurationMonths)
	}

	if !p.TaxRate.Equal(decimal.NewFromFloat(0.20)) {
		t.Errorf("Expected tax rate 0.20, got %v", p.TaxRate)
	}

	if !p.IsActive {
		t.Errorf("Expected isActive true, got %v", p.IsActive)
	}

	// Test case 2: Create product with negative price
	input = product.CreateProductInput{
		Name:           "Invalid Product",
		Description:    "This should fail validation",
		Price:          decimal.NewFromFloat(-10.00),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
	}

	_, err = service.CreateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for negative price")
	}

	// Test case 3: Create product with zero duration
	input = product.CreateProductInput{
		Name:           "Invalid Product",
		Description:    "This should fail validation",
		Price:          decimal.NewFromFloat(10.00),
		DurationMonths: 0,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
	}

	_, err = service.CreateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for zero duration")
	}

	// Test case 4: Create product with negative tax rate
	input = product.CreateProductInput{
		Name:           "Invalid Product",
		Description:    "This should fail validation",
		Price:          decimal.NewFromFloat(10.00),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(-0.20),
		IsActive:       true,
	}

	_, err = service.CreateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for negative tax rate")
	}

	// Test case 5: Create product with empty name
	input = product.CreateProductInput{
		Name:           "",
		Description:    "This should fail validation",
		Price:          decimal.NewFromFloat(10.00),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
	}

	_, err = service.CreateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestGetProductByID(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := newMockProductRepository()
	service := product.NewService(repo)

	// Create a test product
	testProduct := &models.Product{
		ID:             uuid.New(),
		Name:           "Test Product",
		Description:    "Test Description",
		Price:          decimal.NewFromFloat(19.99),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.Create(ctx, testProduct)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Test case 1: Get existing product
	p, err := service.GetProductByID(ctx, testProduct.ID)
	if err != nil {
		t.Fatal("Failed to get product:", err)
	}

	if p.ID != testProduct.ID {
		t.Errorf("Expected ID %v, got %v", testProduct.ID, p.ID)
	}

	if p.Name != testProduct.Name {
		t.Errorf("Expected name %v, got %v", testProduct.Name, p.Name)
	}

	// Test case 2: Get non-existent product
	_, err = service.GetProductByID(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error for non-existent product")
	}
	if err != errors.ErrProductNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrProductNotFound, err)
	}
}

func TestUpdateProduct(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := newMockProductRepository()
	service := product.NewService(repo)

	// Create a test product
	testProduct := &models.Product{
		ID:             uuid.New(),
		Name:           "Original Name",
		Description:    "Original Description",
		Price:          decimal.NewFromFloat(19.99),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.Create(ctx, testProduct)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Test case 1: Update existing product
	input := product.UpdateProductInput{
		ID:             testProduct.ID,
		Name:           "Updated Name",
		Description:    "Updated Description",
		Price:          decimal.NewFromFloat(29.99),
		DurationMonths: 3,
		TaxRate:        decimal.NewFromFloat(0.25),
		IsActive:       false,
	}

	p, err := service.UpdateProduct(ctx, input)
	if err != nil {
		t.Fatal("Failed to update product:", err)
	}

	if p.Name != "Updated Name" {
		t.Errorf("Expected name Updated Name, got %v", p.Name)
	}

	if p.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %v", p.Description)
	}

	if !p.Price.Equal(decimal.NewFromFloat(29.99)) {
		t.Errorf("Expected price 29.99, got %v", p.Price)
	}

	if p.DurationMonths != 3 {
		t.Errorf("Expected duration 3 months, got %v", p.DurationMonths)
	}

	if !p.TaxRate.Equal(decimal.NewFromFloat(0.25)) {
		t.Errorf("Expected tax rate 0.25, got %v", p.TaxRate)
	}

	if p.IsActive {
		t.Errorf("Expected isActive false, got %v", p.IsActive)
	}

	// Test case 2: Update non-existent product
	input.ID = uuid.New()
	_, err = service.UpdateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for non-existent product")
	}
	if err != errors.ErrProductNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrProductNotFound, err)
	}

	// Test case 3: Update with invalid data
	input.ID = testProduct.ID
	input.Price = decimal.NewFromFloat(-10.00)
	_, err = service.UpdateProduct(ctx, input)
	if err == nil {
		t.Error("Expected error for negative price")
	}
}

func TestDeleteProduct(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := newMockProductRepository()
	service := product.NewService(repo)

	// Create a test product
	testProduct := &models.Product{
		ID:             uuid.New(),
		Name:           "Test Product",
		Description:    "Test Description",
		Price:          decimal.NewFromFloat(19.99),
		DurationMonths: 1,
		TaxRate:        decimal.NewFromFloat(0.20),
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.Create(ctx, testProduct)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Test case 1: Delete existing product
	err = service.DeleteProduct(ctx, testProduct.ID)
	if err != nil {
		t.Fatal("Failed to delete product:", err)
	}

	// Verify product is deleted
	_, err = service.GetProductByID(ctx, testProduct.ID)
	if err == nil {
		t.Error("Expected error after product deletion")
	}
	if err != errors.ErrProductNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrProductNotFound, err)
	}

	// Test case 2: Delete non-existent product
	err = service.DeleteProduct(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error for non-existent product")
	}
	if err != errors.ErrProductNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrProductNotFound, err)
	}
}

func TestGetAllProducts(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := newMockProductRepository()
	service := product.NewService(repo)

	// Test with empty repository
	products, err := service.GetAllProducts(ctx)
	if err != nil {
		t.Fatal("Failed to get products:", err)
	}
	if len(products) != 0 {
		t.Errorf("Expected 0 products, got %d", len(products))
	}

	// Create test products
	for i := 0; i < 3; i++ {
		product := &models.Product{
			ID:             uuid.New(),
			Name:           "Product " + string(rune(i+65)), // A, B, C
			Description:    "Description " + string(rune(i+65)),
			Price:          decimal.NewFromFloat(float64(10 * (i + 1))),
			DurationMonths: i + 1,
			TaxRate:        decimal.NewFromFloat(0.20),
			IsActive:       true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := repo.Create(ctx, product)
		if err != nil {
			t.Fatal("Failed to create test product:", err)
		}
	}

	// Test with populated repository
	products, err = service.GetAllProducts(ctx)
	if err != nil {
		t.Fatal("Failed to get products:", err)
	}
	if len(products) != 3 {
		t.Errorf("Expected 3 products, got %d", len(products))
	}
}
