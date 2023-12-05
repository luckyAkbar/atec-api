package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	httpMock "github.com/sweet-go/stdlib/http/mock"
)

func TestRest_authMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAuthUc := mock.NewMockAuthUsecase(ctrl)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)

	s := &service{
		authUsecase:          mockAuthUc,
		apiResponseGenerator: mockAPIRespGen,
	}

	tests := []common.TestStructure{
		{
			Name:   "auth header is empty",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(false)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "auth header is empty2",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer ")
				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(true)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "failed to validate access",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "errinternal",
					Type:    usecase.ErrAccessTokenExpired,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(false)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase return error internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "errinternal",
					Type:    usecase.ErrInternal,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(true)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "credentials not found",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "not found",
					Type:    usecase.ErrResourceNotFound,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrNotFound.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(false)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "safety test: usecase return nil auth user",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type: nil,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(true)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "admin only & requester is not admin",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type: nil,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(&model.AuthUser{
					AccessToken: "secretboss",
					UserID:      uuid.New(),
					Role:        model.RoleUser,
				}, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrUnauthorized.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				fn := func(c echo.Context) error {
					return nil
				}

				err := s.authMiddleware(true)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "admin only & requester is the holy admin",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type: nil,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(&model.AuthUser{
					AccessToken: "secretboss",
					UserID:      uuid.New(),
					Role:        model.RoleAdmin,
				}, cerr)

				fn := func(c echo.Context) error {
					authUser := model.GetUserFromCtx(c.Request().Context())
					assert.NotNil(t, authUser)
					assert.Equal(t, authUser.Role, model.RoleAdmin)

					return c.JSON(http.StatusOK, `{"message": "ok"}`)
				}

				err := s.authMiddleware(true)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "a valid non admin user",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Authorization", "Bearer secretboss")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type: nil,
				}

				mockAuthUc.EXPECT().ValidateAccess(ectx.Request().Context(), "secretboss").Times(1).Return(&model.AuthUser{
					AccessToken: "secretboss",
					UserID:      uuid.New(),
					Role:        model.RoleUser,
				}, cerr)

				fn := func(c echo.Context) error {
					authUser := model.GetUserFromCtx(c.Request().Context())
					assert.NotNil(t, authUser)
					assert.Equal(t, authUser.Role, model.RoleUser)

					return c.JSON(http.StatusOK, `{"message": "ok"}`)
				}

				err := s.authMiddleware(false)(fn)(ectx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}
