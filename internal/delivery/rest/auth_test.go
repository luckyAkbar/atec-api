package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	stdhttp "github.com/sweet-go/stdlib/http"
	httpMock "github.com/sweet-go/stdlib/http/mock"
)

func TestRest_handleLogIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockAuthUc := mock.NewMockAuthUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "invalid payload",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"email": "luckyakbar1509@gmail.com",
							"password": "password", <- invalid here
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)
				err := restService.handleLogIn()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase returning internal error",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				input := &model.LogInInput{
					Email:    "testing@gmail.com",
					Password: "password",
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"email": "testing@gmail.com",
							"password": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)
				mockAuthUc.EXPECT().LogIn(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Message: "internal err",
					Cause:   errors.New("err"),
					Code:    http.StatusInternalServerError,
					Type:    usecase.ErrInternal,
				})
				err := restService.handleLogIn()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase returning other specific error",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				input := &model.LogInInput{
					Email:    "testing@gmail.com",
					Password: "password",
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"email": "testing@gmail.com",
							"password": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				cerr := &common.Error{
					Message: "specific err",
					Cause:   errors.New("err"),
					Code:    http.StatusBadRequest,
					Type:    usecase.ErrInvalidLoginInput,
				}

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)
				mockAuthUc.EXPECT().LogIn(ectx.Request().Context(), input).Times(1).Return(nil, cerr)
				err := restService.handleLogIn()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				input := &model.LogInInput{
					Email:    "testing@gmail.com",
					Password: "password",
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"email": "testing@gmail.com",
							"password": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				resp := &model.LogInOutput{
					ID: uuid.New(),
				}

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    resp,
				}, nil).Times(1).Return(nil)
				mockAuthUc.EXPECT().LogIn(ectx.Request().Context(), input).Times(1).Return(resp, &common.Error{
					Type: nil,
				})
				err := restService.handleLogIn()(ectx)
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

func TestRest_handleLogOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockAuthUc := mock.NewMockAuthUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "usecase return err internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"accessToken": "accessToken"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAuthUc.EXPECT().LogOut(ectx.Request().Context()).Times(1).Return(&common.Error{
					Type: usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)
				err := restService.handleLogOut()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase return others specific err",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"accessToken": "accessToken"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type: usecase.ErrInvalidLogoutInput,
				}

				mockAuthUc.EXPECT().LogOut(ectx.Request().Context()).Times(1).Return(cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)
				err := restService.handleLogOut()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
					{
						"request": {
							"accessToken": "accessToken"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAuthUc.EXPECT().LogOut(ectx.Request().Context()).Times(1).Return(&common.Error{
					Type: nil,
				})
				err := restService.handleLogOut()(ectx)
				assert.NoError(t, err)
				assert.Equal(t, rec.Result().StatusCode, http.StatusNoContent)
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

func TestRest_handleValidateResetPasswordSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockAuthUc := mock.NewMockAuthUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "usecase return err internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				key := "key"

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("key", key)

				mockAuthUc.EXPECT().ValidateResetPasswordSession(ectx.Request().Context(), key).Times(1).Return(&common.Error{
					Type: usecase.ErrInternal,
					Code: http.StatusInternalServerError,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleValidateResetPasswordSession()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase return specific err",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				key := "key"
				cerr := &common.Error{
					Message: "reset password session is expired",
					Cause:   errors.New("reset password session is expired"),
					Code:    http.StatusForbidden,
					Type:    usecase.ErrResetPasswordSessionExpired,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("key", key)

				mockAuthUc.EXPECT().ValidateResetPasswordSession(ectx.Request().Context(), key).Times(1).Return(cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleValidateResetPasswordSession()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				key := "key"

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("key", key)

				mockAuthUc.EXPECT().ValidateResetPasswordSession(ectx.Request().Context(), key).Times(1).Return(&common.Error{
					Type: nil,
				})

				err := restService.handleValidateResetPasswordSession()(ectx)
				assert.NoError(t, err)
				assert.Equal(t, rec.Result().StatusCode, http.StatusOK)
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

func TestRest_handleResetPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockAuthUc := mock.NewMockAuthUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "invalid input",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/auth/reset-password/", strings.NewReader(`
					{
						"request": {
							"key": "key",
							"password": "newpw123456",
							"passwordConfirmation": "newpw123456"
						}, <- invalid here
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleResetPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "invalid input",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/auth/reset-password/", strings.NewReader(`
					{
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleResetPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "uc return internal error",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/auth/reset-password/", strings.NewReader(`
					{
						"request": {
							"key": "key",
							"password": "newpw123456",
							"passwordConfirmation": "newpw123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				input := &model.ResetPasswordInput{
					Key:                 "key",
					Password:            "newpw123456",
					PasswordConfimation: "newpw123456",
				}

				mockAuthUc.EXPECT().ResetPassword(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Message: "err",
					Cause:   errors.New("err"),
					Type:    usecase.ErrInternal,
					Code:    http.StatusInternalServerError,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleResetPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "uc return specific error",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/auth/reset-password/", strings.NewReader(`
					{
						"request": {
							"key": "key",
							"password": "newpw123456",
							"passwordConfirmation": "newpw123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				input := &model.ResetPasswordInput{
					Key:                 "key",
					Password:            "newpw123456",
					PasswordConfimation: "newpw123456",
				}

				cerr := &common.Error{
					Message: "err",
					Cause:   errors.New("err"),
					Type:    usecase.ErrInvalidResetPasswordInput,
					Code:    http.StatusInternalServerError,
				}

				mockAuthUc.EXPECT().ResetPassword(ectx.Request().Context(), input).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleResetPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					authUsecase:          mockAuthUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/auth/reset-password/", strings.NewReader(`
					{
						"request": {
							"key": "key",
							"password": "newpw123456",
							"passwordConfirmation": "newpw123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				input := &model.ResetPasswordInput{
					Key:                 "key",
					Password:            "newpw123456",
					PasswordConfimation: "newpw123456",
				}

				cerr := &common.Error{
					Type: nil,
				}

				res := &model.ResetPasswordResponse{
					ID: uuid.New(),
				}

				mockAuthUc.EXPECT().ResetPassword(ectx.Request().Context(), input).Times(1).Return(res, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil)

				err := restService.handleResetPassword()(ectx)
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
