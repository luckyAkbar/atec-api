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
}

// NewService will create http service and register all of it's routes
func NewService(rootGroup *echo.Group, apiResponseGenerator stdhttp.APIResponseGenerator, userUsecase model.UserUsecase) {
	s := &service{
		rootGroup:            rootGroup,
		apiResponseGenerator: apiResponseGenerator,
		userUsecase:          userUsecase,
	}

	s.initRoutes()
}

func (s *service) initRoutes() {
	s.rootGroup.POST("/users/accounts/", s.handleSignUp())
	s.rootGroup.POST("/users/accounts/validation/", s.handleAccountVerification())
}
