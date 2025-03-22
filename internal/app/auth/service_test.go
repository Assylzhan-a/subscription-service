package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/app/auth"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	users   map[uuid.UUID]*models.User
	byEmail map[string]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:   make(map[uuid.UUID]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	if _, exists := m.byEmail[user.Email]; exists {
		return errors.ErrUserAlreadyExists
	}

	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := m.byEmail[email]; ok {
		return user, nil
	}
	return nil, errors.ErrUserNotFound
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.ErrUserNotFound
	}

	// Check if email is being changed and if it conflicts with existing email
	oldUser := m.users[user.ID]
	if oldUser.Email != user.Email {
		if _, exists := m.byEmail[user.Email]; exists {
			return errors.ErrUserAlreadyExists
		}
		delete(m.byEmail, oldUser.Email)
		m.byEmail[user.Email] = user
	}

	m.users[user.ID] = user
	return nil
}

type mockJWTManager struct{}

func newMockJWTManager() *jwt.Manager {
	return jwt.NewManager("test-secret-key", "test-issuer")
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// Tests for AuthService
func TestRegisterUser(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	service := auth.NewService(userRepo, jwtManager, time.Hour)

	// Test case 1: Register valid user
	input := auth.RegisterUserInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := service.RegisterUser(ctx, input)
	if err != nil {
		t.Fatal("Failed to register user:", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name Test User, got %v", user.Name)
	}

	// Password should not be returned
	if user.Password != "" {
		t.Errorf("Expected password to be redacted, got %v", user.Password)
	}

	// Test case 2: Register user with existing email
	_, err = service.RegisterUser(ctx, input)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if err != errors.ErrUserAlreadyExists {
		t.Errorf("Expected error %v, got %v", errors.ErrUserAlreadyExists, err)
	}

	// Test case 3: Register user with invalid data (short password)
	input.Email = "another@example.com"
	input.Password = "short"
	_, err = service.RegisterUser(ctx, input)
	if err == nil {
		t.Error("Expected error for short password")
	}
}

func TestLoginUser(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	service := auth.NewService(userRepo, jwtManager, time.Hour)

	// Create a test user with known password
	hashedPassword := hashPassword("password123")
	testUser := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  hashedPassword,
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatal("Failed to create test user:", err)
	}

	// Test case 1: Login with valid credentials
	input := auth.LoginUserInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := service.LoginUser(ctx, input)
	if err != nil {
		t.Fatal("Failed to login user:", err)
	}

	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", response.User.Email)
	}

	if response.User.Name != "Test User" {
		t.Errorf("Expected name Test User, got %v", response.User.Name)
	}

	if response.User.ID != testUser.ID {
		t.Errorf("Expected user ID %v, got %v", testUser.ID, response.User.ID)
	}

	if response.Token == "" {
		t.Error("Expected token to be non-empty")
	}

	if response.ExpiresAt <= time.Now().Unix() {
		t.Error("Expected token expiry to be in the future")
	}

	// Test case 2: Login with invalid password
	input.Password = "wrongpassword"
	_, err = service.LoginUser(ctx, input)
	if err == nil {
		t.Error("Expected error for wrong password")
	}
	if err != errors.ErrInvalidCredentials {
		t.Errorf("Expected error %v, got %v", errors.ErrInvalidCredentials, err)
	}

	// Test case 3: Login with non-existent email
	input.Email = "nonexistent@example.com"
	input.Password = "password123"
	_, err = service.LoginUser(ctx, input)
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
	if err != errors.ErrInvalidCredentials {
		t.Errorf("Expected error %v, got %v", errors.ErrInvalidCredentials, err)
	}

	// Test case 4: Login with invalid input (empty email)
	input.Email = ""
	input.Password = "password123"
	_, err = service.LoginUser(ctx, input)
	if err == nil {
		t.Error("Expected error for empty email")
	}
}

func TestGetUserByID(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	service := auth.NewService(userRepo, jwtManager, time.Hour)

	// Create a test user
	testUser := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  hashPassword("password123"),
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatal("Failed to create test user:", err)
	}

	// Test case 1: Get existing user
	user, err := service.GetUserByID(ctx, testUser.ID)
	if err != nil {
		t.Fatal("Failed to get user:", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name Test User, got %v", user.Name)
	}

	// Password should be redacted
	if user.Password != "" {
		t.Errorf("Expected password to be redacted, got %v", user.Password)
	}

	// Test case 2: Get non-existent user
	_, err = service.GetUserByID(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
	if err != errors.ErrUserNotFound {
		t.Errorf("Expected error %v, got %v", errors.ErrUserNotFound, err)
	}
}
