package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/app/subscription"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockSubscriptionRepository struct {
	subscriptions map[uuid.UUID]*models.Subscription
	stateChanges  []*models.SubscriptionStateChange
}

func newMockSubscriptionRepository() *mockSubscriptionRepository {
	return &mockSubscriptionRepository{
		subscriptions: make(map[uuid.UUID]*models.Subscription),
		stateChanges:  make([]*models.SubscriptionStateChange, 0),
	}
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	m.subscriptions[subscription.ID] = subscription
	return nil
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	if sub, ok := m.subscriptions[id]; ok {
		return sub, nil
	}
	return nil, errors.ErrSubscriptionNotFound
}

func (m *mockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error) {
	var result []*models.Subscription
	for _, sub := range m.subscriptions {
		if sub.UserID == userID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	if _, ok := m.subscriptions[subscription.ID]; !ok {
		return errors.ErrSubscriptionNotFound
	}
	m.subscriptions[subscription.ID] = subscription
	return nil
}

func (m *mockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.subscriptions[id]; !ok {
		return errors.ErrSubscriptionNotFound
	}
	delete(m.subscriptions, id)
	return nil
}

func (m *mockSubscriptionRepository) CreateStateChange(ctx context.Context, stateChange *models.SubscriptionStateChange) error {
	m.stateChanges = append(m.stateChanges, stateChange)
	return nil
}

func (m *mockSubscriptionRepository) GetStateChangesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.SubscriptionStateChange, error) {
	var result []*models.SubscriptionStateChange
	for _, change := range m.stateChanges {
		if change.SubscriptionID == subscriptionID {
			result = append(result, change)
		}
	}
	return result, nil
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
	m.vouchers[voucher.ID] = voucher
	m.codes[voucher.Code] = voucher
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

// Helper function to create a test subscription
func createTestSubscription(userID, productID uuid.UUID, status models.SubscriptionStatus) *models.Subscription {
	return &models.Subscription{
		ID:            uuid.New(),
		UserID:        userID,
		ProductID:     productID,
		Status:        status,
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(0, 1, 0), // 1 month duration
		OriginalPrice: decimal.NewFromFloat(19.99),
		TaxAmount:     decimal.NewFromFloat(4.00),
		TotalAmount:   decimal.NewFromFloat(23.99),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Helper function to create a test voucher
func createTestVoucher() *models.Voucher {
	return &models.Voucher{
		ID:            uuid.New(),
		Code:          "TEST20",
		DiscountType:  models.DiscountTypePercentage,
		DiscountValue: decimal.NewFromInt(20), // 20% discount
		IsActive:      true,
		ExpiresAt:     time.Now().AddDate(0, 1, 0), // Expires in 1 month
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Tests for SubscriptionService
func TestCreateSubscription(t *testing.T) {
	// Setup
	ctx := context.Background()
	subRepo := newMockSubscriptionRepository()
	productRepo := newMockProductRepository()
	voucherRepo := newMockVoucherRepository()
	service := subscription.NewService(subRepo, productRepo, voucherRepo)

	// Create a test product
	product := createTestProduct()
	err := productRepo.Create(ctx, product)
	if err != nil {
		t.Fatal("Failed to create test product:", err)
	}

	// Test case 1: Create subscription without voucher or trial
	userID := uuid.New()
	input := subscription.CreateSubscriptionInput{
		UserID:    userID,
		ProductID: product.ID,
		WithTrial: false,
	}

	sub, err := service.CreateSubscription(ctx, input)
	if err != nil {
		t.Fatal("Failed to create subscription:", err)
	}

	if sub.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, sub.UserID)
	}

	if sub.ProductID != product.ID {
		t.Errorf("Expected ProductID %v, got %v", product.ID, sub.ProductID)
	}

	if sub.Status != models.SubscriptionStatusActive {
		t.Errorf("Expected status %v, got %v", models.SubscriptionStatusActive, sub.Status)
	}

	if sub.TrialEndDate != nil {
		t.Error("Expected TrialEndDate to be nil")
	}

	// Test case 2: Create subscription with trial period
	input.WithTrial = true

	subWithTrial, err := service.CreateSubscription(ctx, input)
	if err != nil {
		t.Fatal("Failed to create subscription with trial:", err)
	}

	if subWithTrial.TrialEndDate == nil {
		t.Error("Expected TrialEndDate not to be nil")
	}
}

func TestPauseSubscription(t *testing.T) {
	// Setup
	ctx := context.Background()
	subRepo := newMockSubscriptionRepository()
	productRepo := newMockProductRepository()
	voucherRepo := newMockVoucherRepository()
	service := subscription.NewService(subRepo, productRepo, voucherRepo)

	userID := uuid.New()
	productID := uuid.New()

	// Create an active subscription
	activeSub := createTestSubscription(userID, productID, models.SubscriptionStatusActive)
	err := subRepo.Create(ctx, activeSub)
	if err != nil {
		t.Fatal("Failed to create active subscription:", err)
	}

	// Test case 1: Pause an active subscription
	err = service.PauseSubscription(ctx, activeSub.ID)
	if err != nil {
		t.Fatal("Failed to pause subscription:", err)
	}

	// Retrieve the updated subscription
	updatedSub, err := subRepo.GetByID(ctx, activeSub.ID)
	if err != nil {
		t.Fatal("Failed to get updated subscription:", err)
	}

	if updatedSub.Status != models.SubscriptionStatusPaused {
		t.Errorf("Expected status %v, got %v", models.SubscriptionStatusPaused, updatedSub.Status)
	}

	// Test case 2: Pause a subscription that's already paused
	err = service.PauseSubscription(ctx, activeSub.ID)
	if err == nil {
		t.Error("Expected error when pausing already paused subscription")
	}

	// Test case 3: Pause a subscription that's in trial period
	trialSub := createTestSubscription(userID, productID, models.SubscriptionStatusActive)
	trialEnd := time.Now().AddDate(0, 1, 0) // 1 month trial
	trialSub.TrialEndDate = &trialEnd

	err = subRepo.Create(ctx, trialSub)
	if err != nil {
		t.Fatal("Failed to create trial subscription:", err)
	}

	err = service.PauseSubscription(ctx, trialSub.ID)
	if err == nil {
		t.Error("Expected error when pausing subscription in trial period")
	}
}

func TestUnpauseSubscription(t *testing.T) {
	// Setup
	ctx := context.Background()
	subRepo := newMockSubscriptionRepository()
	productRepo := newMockProductRepository()
	voucherRepo := newMockVoucherRepository()
	service := subscription.NewService(subRepo, productRepo, voucherRepo)

	userID := uuid.New()
	productID := uuid.New()

	// Create a paused subscription
	pausedSub := createTestSubscription(userID, productID, models.SubscriptionStatusPaused)
	err := subRepo.Create(ctx, pausedSub)
	if err != nil {
		t.Fatal("Failed to create paused subscription:", err)
	}

	// Test case 1: Unpause a paused subscription
	err = service.UnpauseSubscription(ctx, pausedSub.ID)
	if err != nil {
		t.Fatal("Failed to unpause subscription:", err)
	}

	// Retrieve the updated subscription
	updatedSub, err := subRepo.GetByID(ctx, pausedSub.ID)
	if err != nil {
		t.Fatal("Failed to get updated subscription:", err)
	}

	if updatedSub.Status != models.SubscriptionStatusActive {
		t.Errorf("Expected status %v, got %v", models.SubscriptionStatusActive, updatedSub.Status)
	}

	// Test case 2: Unpause a subscription that's already active
	err = service.UnpauseSubscription(ctx, pausedSub.ID)
	if err == nil {
		t.Error("Expected error when unpausing already active subscription")
	}
}

func TestCancelSubscription(t *testing.T) {
	// Setup
	ctx := context.Background()
	subRepo := newMockSubscriptionRepository()
	productRepo := newMockProductRepository()
	voucherRepo := newMockVoucherRepository()
	service := subscription.NewService(subRepo, productRepo, voucherRepo)

	userID := uuid.New()
	productID := uuid.New()

	// Create subscriptions with different statuses
	activeSub := createTestSubscription(userID, productID, models.SubscriptionStatusActive)
	pausedSub := createTestSubscription(userID, productID, models.SubscriptionStatusPaused)

	err := subRepo.Create(ctx, activeSub)
	if err != nil {
		t.Fatal("Failed to create active subscription:", err)
	}

	err = subRepo.Create(ctx, pausedSub)
	if err != nil {
		t.Fatal("Failed to create paused subscription:", err)
	}

	// Test case 1: Cancel an active subscription
	err = service.CancelSubscription(ctx, activeSub.ID)
	if err != nil {
		t.Fatal("Failed to cancel active subscription:", err)
	}

	// Retrieve the updated subscription
	updatedActiveSub, err := subRepo.GetByID(ctx, activeSub.ID)
	if err != nil {
		t.Fatal("Failed to get updated active subscription:", err)
	}

	if updatedActiveSub.Status != models.SubscriptionStatusCancelled {
		t.Errorf("Expected status %v, got %v", models.SubscriptionStatusCancelled, updatedActiveSub.Status)
	}

	// Test case 2: Cancel a paused subscription
	err = service.CancelSubscription(ctx, pausedSub.ID)
	if err != nil {
		t.Fatal("Failed to cancel paused subscription:", err)
	}

	// Retrieve the updated subscription
	updatedPausedSub, err := subRepo.GetByID(ctx, pausedSub.ID)
	if err != nil {
		t.Fatal("Failed to get updated paused subscription:", err)
	}

	if updatedPausedSub.Status != models.SubscriptionStatusCancelled {
		t.Errorf("Expected status %v, got %v", models.SubscriptionStatusCancelled, updatedPausedSub.Status)
	}
}
