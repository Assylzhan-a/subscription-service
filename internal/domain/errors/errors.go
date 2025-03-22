package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")

	ErrProductNotFound = errors.New("product not found")
	ErrInactiveProduct = errors.New("product is not active")
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
