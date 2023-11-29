package usecase

import (
	"errors"

	"github.com/luckyAkbar/atec-api/internal/common"
)

var (
	// ErrInternal is returned when internal error occurs, such as database error, etc
	ErrInternal = errors.New("000000")

	// ErrEmailInputInvalid is returned when email input is invalid, such as empty subject, 0 receipients, etc
	ErrEmailInputInvalid = errors.New("email input invalid")

	// ErrInputSignUpInvalid will be returned when input is invalid, such as empty email, empty password, etc
	ErrInputSignUpInvalid = errors.New("001001")

	// ErrEmailAlreadyRegistered will be returned when user tried to sign up using an already registered email
	ErrEmailAlreadyRegistered = errors.New("001002")
)

var nilErr = &common.Error{
	Type: nil,
}
