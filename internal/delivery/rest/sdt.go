package rest

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/sirupsen/logrus"
	stdhttp "github.com/sweet-go/stdlib/http"
)

func (s *service) handleInitiateSDTest() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.InitiateSDTestInput `json:"request"`
			Signature string                     `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil || input.Request == nil {
			logrus.Error(err)
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		requester := model.GetUserFromCtx(c.Request().Context())
		if requester != nil {
			input.Request.UserID = uuid.NullUUID{
				UUID:  requester.UserID,
				Valid: true,
			}
		}

		resp, custerr := s.sdtestUsecase.Initiate(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle initiate sd test")
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

func (s *service) handleSubmitSDTestAnswer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = struct {
			Request   *model.SubmitSDTestInput `json:"request"`
			Signature string                   `json:"signature"`
		}{}
		if err := c.Bind(&input); err != nil || input.Request == nil {
			logrus.Error(err)
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, custerr := s.sdtestUsecase.Submit(c.Request().Context(), input.Request)
		switch custerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			logrus.WithContext(c.Request().Context()).WithError(custerr.Cause).Error("failed to handle submit sd test")
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

func (s *service) handleViewSDTestHistories() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.ViewHistoriesInput{}

		// no need to check err here. Zero value in input is fine
		_ = c.Bind(input)

		resp, custerr := s.sdtestUsecase.Histories(c.Request().Context(), input)
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

func (s *service) handleGetSDTestStatistic() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := c.Param("user_id")
		id, err := uuid.Parse(input)
		if err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		resp, cerr := s.sdtestUsecase.Statistic(c.Request().Context(), id)
		switch cerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, cerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
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

func (s *service) handleDownloadTestResult() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := c.Param("id")
		id, err := uuid.Parse(input)
		if err != nil {
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
		}

		res, cerr := s.sdtestUsecase.DownloadResult(c.Request().Context(), id)
		switch cerr.Type {
		default:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, cerr.GenerateStdlibHTTPResponse(nil), nil)
		case usecase.ErrInternal:
			return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
		case nil:
			c.Response().Header().Set("Content-Type", res.ContentType)
			c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", res.Buffer.Len()))
			return c.Blob(http.StatusOK, res.ContentType, res.Buffer.Bytes())
		}
	}
}
