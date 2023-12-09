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
	sdtemplateUsecase    model.SDTemplateUsecase
}

// NewService will create http service and register all of it's routes
func NewService(rootGroup *echo.Group, apiResponseGenerator stdhttp.APIResponseGenerator, userUsecase model.UserUsecase, authUsecase model.AuthUsecase, sdtemplateUsecase model.SDTemplateUsecase) {
	s := &service{
		rootGroup:            rootGroup,
		apiResponseGenerator: apiResponseGenerator,
		userUsecase:          userUsecase,
		authUsecase:          authUsecase,
		sdtemplateUsecase:    sdtemplateUsecase,
	}

	s.initRoutes()
}

func (s *service) initRoutes() {
	s.rootGroup.POST("/users/accounts/", s.handleSignUp())
	s.rootGroup.POST("/users/accounts/validation/", s.handleAccountVerification())
	s.rootGroup.PATCH("/users/accounts/:id/reset-password/", s.handleInitiateResetUserPassword(), s.authMiddleware(true))

	s.rootGroup.POST("/auth/sessions/", s.handleLogIn())
	s.rootGroup.DELETE("/auth/sessions/", s.handleLogOut(), s.authMiddleware(false))
	s.rootGroup.GET("/auth/reset-password/", s.handleValidateResetPasswordSession())
	s.rootGroup.PATCH("/auth/reset-password/", s.handleResetPassword())

	s.rootGroup.POST("/sdt/templates/", s.handleCreateSDTemplate(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/templates/:id/", s.handleFindSDTemplateByID(), s.authMiddleware(true))
	s.rootGroup.PUT("/sdt/templates/:id/", s.handleUpdateSDTemplate(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/templates/", s.handleSearchSDTemplate(), s.authMiddleware(true))
}
