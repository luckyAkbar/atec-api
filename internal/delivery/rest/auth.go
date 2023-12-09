package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/sirupsen/logrus"
	stdhttp "github.com/sweet-go/stdlib/http"
)

func (s *service) handleLogIn() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.LogInInput `json:"request"`
			Signature string            `json:"signature"`
		}{}
		if c.Bind(&input) != nil || input.Request == nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.authUsecase.LogIn(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle log in request")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, &stdhttp.StandardResponse{
				Success: true,
				Message: "success",
				Status:  http.StatusOK,
				Data:    resp,
			}, nil)
		}
	}
}

func (s *service) handleLogOut() echo.HandlerFunc {
	return func(c echo.Context) error {
		custerr := s.authUsecase.LogOut(c.Request().Context())
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle log out request")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return c.NoContent(http.StatusNoContent)
		}
	}
}

func (s *service) handleValidateResetPasswordSession() echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.QueryParam("key")
		custerr := s.authUsecase.ValidateResetPasswordSession(c.Request().Context(), key)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle validate reset password session request")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return c.NoContent(http.StatusOK)
		}

	}
}

func (s *service) handleResetPassword() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.ResetPasswordInput `json:"request"`
			Signature string                    `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil || input.Request == nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.authUsecase.ResetPassword(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle reset password request")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, &stdhttp.StandardResponse{
				Success: true,
				Message: "success",
				Status:  http.StatusOK,
				Data:    resp,
			}, nil)
		}
	}
}
