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

func (s *service) handleCreateSDTemplate() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.SDTemplate `json:"request"`
			Signature string            `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil || input.Request == nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdtemplateUsecase.Create(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle log out request")
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

func (s *service) handleFindSDTemplateByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := c.Param("id")
		id, err := uuid.Parse(input)
		if err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdtemplateUsecase.FindByID(c.Request().Context(), id)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle log out request")
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

func (s *service) handleSearchSDTemplate() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.SearchSDTemplateInput{}
		if err := c.Bind(input); err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdtemplateUsecase.Search(c.Request().Context(), input)
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

func (s *service) handleUpdateSDTemplate() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := struct {
			Request   *model.SDTemplate `json:"request"`
			Signature string            `json:"signature"`
		}{}

		id := c.Param("id")
		templateID, parsingErr := uuid.Parse(id)

		if err := c.Bind(&input); err != nil || input.Request == nil || parsingErr != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdtemplateUsecase.Update(c.Request().Context(), templateID, input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle log out request")
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