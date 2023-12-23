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
