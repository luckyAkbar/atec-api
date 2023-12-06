package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/model"
	stdhttp "github.com/sweet-go/stdlib/http"
)

type service struct {
	rootGroup            *echo.Group
	apiResponseGenerator stdhttp.APIResponseGenerator
	userUsecase          model.UserUsecase
	authUsecase          model.AuthUsecase
}

// NewService will create http service and register all of it's routes
func NewService(rootGroup *echo.Group, apiResponseGenerator stdhttp.APIResponseGenerator, userUsecase model.UserUsecase, authUsecase model.AuthUsecase) {
	s := &service{
		rootGroup:            rootGroup,
		apiResponseGenerator: apiResponseGenerator,
		userUsecase:          userUsecase,
		authUsecase:          authUsecase,
	}

	s.initRoutes()
}

func (s *service) initRoutes() {
	s.rootGroup.POST("/users/accounts/", s.handleSignUp())
	s.rootGroup.POST("/users/accounts/validation/", s.handleAccountVerification())
	s.rootGroup.PATCH("/users/accounts/:id/reset-password/", s.handleInitiateResetUserPassword(), s.authMiddleware(true))

	s.rootGroup.POST("/auth/sessions/", s.handleLogIn())
	s.rootGroup.DELETE("/auth/sessions/", s.handleLogOut(), s.authMiddleware(false))
}
