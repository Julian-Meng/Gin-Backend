package dao

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrBusinessRule       = errors.New("business rule violation")
)

func NotFound(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrNotFound)
}

func InvalidCredentials(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrInvalidCredentials)
}

func Business(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrBusinessRule)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

func IsInvalidCredentials(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}

func IsBusiness(err error) bool {
	return errors.Is(err, ErrBusinessRule)
}
