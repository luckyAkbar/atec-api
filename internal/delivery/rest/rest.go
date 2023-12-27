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
	sdpackageUsecase     model.SDPackageUsecase
	sdtestUsecase        model.SDTestUsecase
}

// NewService will create http service and register all of it's routes
func NewService(rootGroup *echo.Group, apiResponseGenerator stdhttp.APIResponseGenerator, userUsecase model.UserUsecase, authUsecase model.AuthUsecase, sdtemplateUsecase model.SDTemplateUsecase, sdpackageUsecase model.SDPackageUsecase, sdtestUsecase model.SDTestUsecase) {
	s := &service{
		rootGroup:            rootGroup,
		apiResponseGenerator: apiResponseGenerator,
		userUsecase:          userUsecase,
		authUsecase:          authUsecase,
		sdtemplateUsecase:    sdtemplateUsecase,
		sdpackageUsecase:     sdpackageUsecase,
		sdtestUsecase:        sdtestUsecase,
	}

	s.initRoutes()
}

func (s *service) initRoutes() {
	s.rootGroup.GET("/users/", s.handleSearchUsers(), s.authMiddleware(true))
	s.rootGroup.POST("/users/accounts/", s.handleSignUp())
	s.rootGroup.POST("/users/accounts/validation/", s.handleAccountVerification())
	s.rootGroup.PATCH("/users/accounts/:id/reset-password/", s.handleInitiateResetUserPassword(), s.authMiddleware(true))
	s.rootGroup.PATCH("/users/accounts/:id/activation-status/", s.handleChangeUserActivationStatus(), s.authMiddleware(true))

	s.rootGroup.POST("/auth/sessions/", s.handleLogIn())
	s.rootGroup.DELETE("/auth/sessions/", s.handleLogOut(), s.authMiddleware(false))
	s.rootGroup.GET("/auth/reset-password/", s.handleValidateResetPasswordSession())
	s.rootGroup.PATCH("/auth/reset-password/", s.handleResetPassword())

	s.rootGroup.POST("/sdt/templates/", s.handleCreateSDTemplate(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/templates/:id/", s.handleFindSDTemplateByID(), s.authMiddleware(true))
	s.rootGroup.PUT("/sdt/templates/:id/", s.handleUpdateSDTemplate(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/templates/", s.handleSearchSDTemplate(), s.authMiddleware(true))
	s.rootGroup.DELETE("/sdt/templates/:id/", s.handleDeleteSDTemplate(), s.authMiddleware(true))
	s.rootGroup.PATCH("/sdt/templates/:id/", s.handleUndoDeleteSDTemplate(), s.authMiddleware(true))
	s.rootGroup.PATCH("/sdt/templates/:id/activation-status/", s.handleChangeSDTemplateActivationStatus(), s.authMiddleware(true))

	s.rootGroup.POST("/sdt/packages/", s.handleCreateSDPackage(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/packages/lists/", s.handleFindReadyToUsePackages())
	s.rootGroup.GET("/sdt/packages/:id/", s.handleFindSDPackageByID(), s.authMiddleware(true))
	s.rootGroup.GET("/sdt/packages/", s.handleSearchSDPackage(), s.authMiddleware(true))
	s.rootGroup.PUT("/sdt/packages/:id/", s.handleUpdateSDPackage(), s.authMiddleware(true))
	s.rootGroup.DELETE("/sdt/packages/:id/", s.handleDeleteSDPackage(), s.authMiddleware(true))
	s.rootGroup.PATCH("/sdt/packages/:id/", s.handleUndoDeleteSDPackage(), s.authMiddleware(true))
	s.rootGroup.PATCH("/sdt/packages/:id/activation-status/", s.handleChangeSDPackageActivationStatus(), s.authMiddleware(true))

	s.rootGroup.POST("/sdt/tests/", s.handleInitiateSDTest(), s.allowUnauthorizedAccess())
	s.rootGroup.POST("/sdt/tests/submissions/", s.handleSubmitSDTestAnswer(), s.allowUnauthorizedAccess())
}
