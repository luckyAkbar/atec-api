package usecase

import "errors"

// error list
var (
	// ErrInternal is returned when internal error occurs, such as database error, etc
	ErrInternal = errors.New("internal error")

	// ErrEmailInputInvalid is returned when email input is invalid, such as empty subject, 0 receipients, etc
	ErrEmailInputInvalid = errors.New("email input invalid")
)
