package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("forbidden access")

	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrProductNotFound = errors.New("product not found")
	ErrInactiveProduct = errors.New("product is not active")

	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrSubscriptionNotActive     = errors.New("subscription is not active")
	ErrSubscriptionInTrial       = errors.New("subscription is in trial period")
	ErrSubscriptionAlreadyPaused = errors.New("subscription is already paused")

	ErrVoucherNotFound = errors.New("voucher not found")
	ErrVoucherExpired  = errors.New("voucher is expired")
	ErrVoucherInactive = errors.New("voucher is not active")
	ErrVoucherInvalid  = errors.New("voucher is invalid")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	if len(e) == 1 {
		return e[0].Error()
	}

	return fmt.Sprintf("%d validation errors occurred", len(e))
}
