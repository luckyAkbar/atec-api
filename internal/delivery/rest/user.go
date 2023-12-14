package rest

import (
	"net/http"

	"github.com/google/uuid"
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
		if c.Bind(&input) != nil || input.Request == nil {
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
		if c.Bind(&input) != nil || input.Request == nil {
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

func (s *service) handleInitiateResetUserPassword() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		userID, err := uuid.Parse(id)
		if err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		res, cerr := s.userUsecase.InitiateResetPassword(c.Request().Context(), userID)
		switch cerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, cerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(cerr.Cause).Error("failed to initiate reset password")
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

func (s *service) handleSearchUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.SearchUserInput{}
		if err := c.Bind(input); err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.userUsecase.Search(c.Request().Context(), input)
		if custerr.Type != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		}

		return s.apiResponseGenerator.GenerateEchoAPIResponse(c, &stdhttp.StandardResponse{
			Success: true,
			Message: "success",
			Status:  http.StatusOK,
			Data:    resp,
		}, nil)
	}
}
