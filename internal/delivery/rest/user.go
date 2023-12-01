package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/sirupsen/logrus"
	stdhttp "github.com/sweet-go/stdlib/http"
)

func (s *service) handleSignUp() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.SignUpInput `json:"request"`
			Signature string             `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		res, custerr := s.userUsecase.SignUp(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to perform signup new user")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, &stdhttp.StandardResponse{
				Success: true,
				Message: "success",
				Status:  http.StatusOK,
				Data:    res,
			}, nil)
		}
	}
}

func (s *service) handleAccountVerification() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.AccountVerificationInput `json:"request"`
			Signature string                          `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		success, failed, custerr := s.userUsecase.VerifyAccount(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(failed), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle account verification")
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, &stdhttp.StandardResponse{
				Success: true,
				Message: "success",
				Status:  http.StatusOK,
				Data:    success,
			}, nil)
		}
	}
}
