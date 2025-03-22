package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/internal/repository"
	"github.com/assylzhan-a/subscription-service/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
	jwtTTL     time.Duration
}

func NewService(
	userRepo repository.UserRepository,
	jwtManager *jwt.Manager,
	jwtTTL time.Duration,
) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		jwtTTL:     jwtTTL,
	}
}

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

func (i *RegisterUserInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if strings.TrimSpace(i.Email) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "email",
			Message: "must not be empty",
		})
	}

	if len(i.Password) < 8 {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "password",
			Message: "must be at least 8 characters long",
		})
	}

	return validationErrors
}

func (s *Service) RegisterUser(ctx context.Context, input RegisterUserInput) (*models.User, error) {
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	_, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.ErrUserAlreadyExists
	} else if err != errors.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    input.Email,
		Password: hashedPassword,
		Name:     input.Name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.Password = ""
	return user, nil
}

type LoginUserInput struct {
	Email    string
	Password string
}

func (i *LoginUserInput) Validate() errors.ValidationErrors {
	var validationErrors errors.ValidationErrors

	if strings.TrimSpace(i.Email) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "email",
			Message: "must not be empty",
		})
	}

	if strings.TrimSpace(i.Password) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "password",
			Message: "must not be empty",
		})
	}

	return validationErrors
}

type LoginResponse struct {
	User      *models.User `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expires_at"`
}

func (s *Service) LoginUser(ctx context.Context, input LoginUserInput) (*LoginResponse, error) {
	if validationErrors := input.Validate(); len(validationErrors) > 0 {
		return nil, validationErrors
	}

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	if err := validatePassword(input.Password, user.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.jwtTTL)
	token, err := s.jwtManager.GenerateToken(user.ID, s.jwtTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	user.Password = ""

	return &LoginResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func validatePassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
