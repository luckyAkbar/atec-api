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

	ErrInternal = &common.Error{
		Message: "internal server error",
		Cause:   errors.New("internal server error"),
		Code:    http.StatusInternalServerError,
		Type:    echo.ErrInternalServerError,
	}
)
