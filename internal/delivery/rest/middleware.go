package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (s *service) authMiddleware(adminOnly bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := getAccessToken(c.Request())
			fmt.Println(token)
			if token == "" {
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil)
			}

			authUser, custErr := s.authUsecase.ValidateAccess(c.Request().Context(), token)
			switch custErr.Type {
			default:
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, custErr.GenerateStdlibHTTPResponse(nil), nil)
			case usecase.ErrInternal:
				logrus.WithContext(c.Request().Context()).WithFields(logrus.Fields{
					"error": custErr,
					"token": token,
				}).Error("failed to validate access")
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
			case usecase.ErrResourceNotFound:
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil)
			case nil:
				break
			}

			// safety check
			if authUser == nil {
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil)
			}

			if adminOnly && !authUser.IsAdmin() {
				return s.apiResponseGenerator.GenerateEchoAPIResponse(c, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil)
			}

			ctx := c.Request().Context()
			newCtx := model.SetUserToCtx(ctx, *authUser)

			c.SetRequest(c.Request().WithContext(newCtx))
			return next(c)
		}
	}
}

func getAccessToken(req *http.Request) (accessToken string) {
	authHeaders := strings.Split(req.Header.Get("Authorization"), " ")

	if (len(authHeaders) != 2) || (authHeaders[0] != "Bearer") {
		return ""
	}

	return strings.TrimSpace(authHeaders[1])
}
