// Package rest will be an rest based presenter layer
package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/common"
)

// list of general error code as defined in api contract
var (
	ErrBadRequest = &common.Error{
		Message: "invalid payload",
		Cause:   errors.New("invalid payload"),
		Code:    http.StatusBadRequest,
		Type:    echo.ErrBadRequest,
	}

	ErrUnauthorized = &common.Error{
		Message: "unauthorized",
		Cause:   errors.New("unauthorized"),
		Code:    http.StatusUnauthorized,
		Type:    echo.ErrUnauthorized,
	}

	ErrNotFound = &common.Error{
		Message: "resource not found",
		Cause:   errors.New("resource not found"),
		Code:    http.StatusNotFound,
		Type:    echo.ErrNotFound,
	}

	ErrInternal = &common.Error{
		Message: "internal server error",
		Cause:   errors.New("internal server error"),
		Code:    http.StatusInternalServerError,
		Type:    echo.ErrInternalServerError,
	}
)
