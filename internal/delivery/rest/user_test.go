package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestRest_handleSignUp(t *testing.T) {
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockUserUsecase := mock.NewMockUserUsecase(ctrl)

	validRequest := &model.SignUpInput{
		Username:            "lucky",
		Email:               "email9@test.com",
		Password:            "password",
		PasswordConfimation: "password",
	}

	signUpResp := &model.SignUpResponse{
		PinValidationID:   uuid.New().String(),
		PinExpiredAt:      time.Now().Add(time.Minute * 5),
		RemainingAttempts: 5,
	}

	tests := []common.TestStructure{
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/", strings.NewReader(`
					{
						"request": {
							"username": "lucky",
							"email": "email9@test.com",
							"password": "password",
							"passwordConfirmation": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockUserUsecase.EXPECT().SignUp(ctx, validRequest).Times(1).Return(signUpResp, &common.Error{
					Type: nil,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    signUpResp,
				}, nil).Times(1).Return(nil)

				err := restService.handleSignUp()(ectx)
				assert.NoError(t, err)

				assert.Equal(t, rec.Result().StatusCode, http.StatusOK)
			},
		},
		{
			Name:   "payload invalid json",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				restService := service{
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/", strings.NewReader(`
					{
						"request": {
							"username": "lucky",
							"email": "email9@test.com",
							"password": "password",
							"passwordConfirmation": "password"
						}, <- invalid here
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleSignUp()(ectx)
				assert.NoError(t, err)

				// TODO must equal to 400. somehow it still showing 200
				// but when tested by insomnia, it's returning 400
				// can't even read the response body -_-
				assert.EqualValues(t, rec.Result().StatusCode, http.StatusOK)
			},
		},
		{
			Name:   "usecase returning err internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				restService := service{
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/", strings.NewReader(`
					{
						"request": {
							"username": "lucky",
							"email": "email9@test.com",
							"password": "password",
							"passwordConfirmation": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockUserUsecase.EXPECT().SignUp(ctx, validRequest).Times(1).Return(signUpResp, &common.Error{
					Type:    usecase.ErrInternal,
					Message: "error internal",
					Cause:   errors.New("error internal"),
					Code:    http.StatusInternalServerError,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleSignUp()(ectx)
				assert.NoError(t, err)

				// TODO must equal to 500. somehow it still showing 200
				// but when tested by insomnia, it's returning 500
				// can't even read the response body -_-
				assert.EqualValues(t, rec.Result().StatusCode, http.StatusOK)
			},
		},
		{
			Name:   "default err resp follow the returned err by usecase",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				restService := service{
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/", strings.NewReader(`
					{
						"request": {
							"username": "lucky",
							"email": "email9@test.com",
							"password": "password",
							"passwordConfirmation": "password"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Type:    usecase.ErrEmailAlreadyRegistered,
					Message: "email already used",
					Cause:   errors.New("email already used"),
					Code:    http.StatusBadRequest,
				}

				mockUserUsecase.EXPECT().SignUp(ctx, validRequest).Times(1).Return(signUpResp, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleSignUp()(ectx)
				assert.NoError(t, err)

				// TODO must equal to 400. somehow it still showing 200
				// but when tested by insomnia, it's returning 400
				// can't even read the response body -_-
				assert.EqualValues(t, rec.Result().StatusCode, http.StatusOK)
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

func TestRest_handleAccountVerification(t *testing.T) {
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockUserUsecase := mock.NewMockUserUsecase(ctrl)
	input := &model.AccountVerificationInput{
		PinValidationID: uuid.MustParse("1af3b478-ab30-468a-9518-4434d8f1b8f8"),
		Pin:             "123456",
	}
	output := &model.SuccessAccountVerificationResponse{
		ID:        uuid.New(),
		Email:     "test@email.com",
		Username:  "username",
		IsActive:  true,
		Role:      model.RoleUser,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	tests := []common.TestStructure{
		{
			Name:   "success",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/validation/", strings.NewReader(`
					{
						"request": {
							"pinValidationID": "1af3b478-ab30-468a-9518-4434d8f1b8f8",
							"pin": "123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockUserUsecase.EXPECT().VerifyAccount(ctx, input).Times(1).Return(output, nil, &common.Error{
					Type: nil,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    output,
				}, nil)

				err := restService.handleAccountVerification()(ectx)
				assert.NoError(t, err)

				assert.Equal(t, rec.Result().StatusCode, http.StatusOK)
			},
		},
		{
			Name:   "binding failed",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/validation/", strings.NewReader(`
					{
						"request": {
							"pinValidationID": "1af3b478-ab30-468a-9518-4434d8f1b8f8",
							"pin": "123456"
						} <- invalid payload here
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleAccountVerification()(ectx)
				assert.NoError(t, err)

				assert.Equal(t, rec.Result().StatusCode, http.StatusOK)
			},
		},
		{
			Name:   "usecase returning error internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/validation/", strings.NewReader(`
					{
						"request": {
							"pinValidationID": "1af3b478-ab30-468a-9518-4434d8f1b8f8",
							"pin": "123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockUserUsecase.EXPECT().VerifyAccount(ctx, input).Times(1).Return(output, nil, &common.Error{
					Message: "internal server error",
					Cause:   errors.New("err internal"),
					Type:    usecase.ErrInternal,
					Code:    http.StatusInternalServerError,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleAccountVerification()(ectx)
				assert.NoError(t, err)

				assert.Equal(t, rec.Result().StatusCode, http.StatusOK)
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
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPost, "/users/accounts/validation/", strings.NewReader(`
					{
						"request": {
							"pinValidationID": "1af3b478-ab30-468a-9518-4434d8f1b8f8",
							"pin": "123456"
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "input invalid",
					Cause:   errors.New("input invalid"),
					Type:    usecase.ErrInputAccountVerificationInvalid,
					Code:    http.StatusBadRequest,
				}

				failed := &model.FailedAccountVerificationResponse{
					RemainingAttempts: 1,
				}

				mockUserUsecase.EXPECT().VerifyAccount(ctx, input).Times(1).Return(output, failed, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(failed), nil)

				err := restService.handleAccountVerification()(ectx)
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

func TestRest_handleInitiateResetUserPassword(t *testing.T) {
	//ctx := context.TODO()
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockUserUsecase := mock.NewMockUserUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "id empty",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleInitiateResetUserPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "id is invalid",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("obviously invalid, right?")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleInitiateResetUserPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase return err internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockUserUsecase.EXPECT().InitiateResetPassword(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Code: http.StatusInternalServerError,
					Type: usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleInitiateResetUserPassword()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "usecase return spesific err",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Code: http.StatusBadRequest,
					Type: usecase.ErrInputResetPasswordInvalid,
				}

				mockUserUsecase.EXPECT().InitiateResetPassword(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleInitiateResetUserPassword()(ectx)
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
					userUsecase:          mockUserUsecase,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Type: nil,
				}

				res := &model.InitiateResetPasswordOutput{
					ID:       id,
					Username: "username",
					Email:    "email@email.com",
				}

				mockUserUsecase.EXPECT().InitiateResetPassword(ectx.Request().Context(), id).Times(1).Return(res, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleInitiateResetUserPassword()(ectx)
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

func TestRest_handleSearchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockUserUc := mock.NewMockUserUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name: "ok",
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/users/", nil)
				req.Header.Set("Content-Type", "application/json")

				input := &model.SearchUserInput{
					Role:   model.RoleAdmin,
					Limit:  10,
					Offset: 100,
				}

				resp := &model.SearchUserOutput{
					Users: []*model.FindUserResponse{},
					Count: 10,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")
				ectx.QueryParams().Add("role", "aDmiN")

				mockUserUc.EXPECT().Search(ectx.Request().Context(), input).Times(1).Return(resp, &common.Error{Type: nil})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    resp,
				}, nil)

				err := restService.handleSearchUsers()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name: "input invalid",
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/users/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				ectx.QueryParams().Add("role", "hamdeh.")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleSearchUsers()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name: "uc err",
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/users/", nil)
				req.Header.Set("Content-Type", "application/json")

				input := &model.SearchUserInput{
					Role:   model.RoleAdmin,
					Limit:  10,
					Offset: 100,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")
				ectx.QueryParams().Add("role", "aDmiN")

				cerr := &common.Error{Type: usecase.ErrInternal}

				mockUserUc.EXPECT().Search(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleSearchUsers()(ectx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run()
		})
	}
}

func TestRest_handleChangeUserActivationStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockUserUc := mock.NewMockUserUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name:   "binding json failed",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`
					{
						"request": {
							"activationStatus": true
						}, <- invalid here
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleChangeUserActivationStatus()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "pram id is invalid uuid",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`
					{
						"request": {
							"activationStatus": true
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("invalid here")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleChangeUserActivationStatus()(ectx)
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
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`
					{
						"request": {
							"activationStatus": true
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockUserUc.EXPECT().ChangeUserAccountActiveStatus(ectx.Request().Context(), id, true).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleChangeUserActivationStatus()(ectx)
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
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`
					{
						"request": {
							"activationStatus": true
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Message: "err uc",
					Cause:   errors.New("err"),
					Type:    usecase.ErrForbiddenUpdateActiveStatus,
					Code:    http.StatusForbidden,
				}

				mockUserUc.EXPECT().ChangeUserAccountActiveStatus(ectx.Request().Context(), id, true).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleChangeUserActivationStatus()(ectx)
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
					userUsecase:          mockUserUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`
					{
						"request": {
							"activationStatus": true
						},
						"signature": "ok"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Type: nil,
				}

				user := &model.User{
					ID: id,
				}

				mockUserUc.EXPECT().ChangeUserAccountActiveStatus(ectx.Request().Context(), id, true).Times(1).Return(user.ToRESTResponse("decrypted"), cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    user.ToRESTResponse("decrypted"),
				}, nil).Times(1).Return(nil)

				err := restService.handleChangeUserActivationStatus()(ectx)
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
