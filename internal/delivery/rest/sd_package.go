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

func (s *service) handleCreateSDPackage() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.SDPackage `json:"request"`
			Signature string           `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil || input.Request == nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdpackageUsecase.Create(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle create sd package request")
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

func (s *service) handleFindSDPackageByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := c.Param("id")
		id, err := uuid.Parse(input)
		if err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdpackageUsecase.FindByID(c.Request().Context(), id)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle find sd package by id request")
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

func (s *service) handleSearchSDPackage() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.SearchSDPackageInput{}
		if err := c.Bind(input); err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdpackageUsecase.Search(c.Request().Context(), input)
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
