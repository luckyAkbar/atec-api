package rest

import (
	"context"
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

func TestRest_handleInitiateSDTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDTestUc := mock.NewMockSDTestUsecase(ctrl)

	input := &model.InitiateSDTestInput{}

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
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
					{
						"request": {
							
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				res := &model.InitiateSDTestOutput{
					ID: uuid.New(),
				}

				mockSDTestUc.EXPECT().Initiate(ectx.Request().Context(), input).Times(1).Return(res, &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ok - with user from ctx",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
					{
						"request": {
							
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				user := model.AuthUser{
					UserID: uuid.New(),
				}
				authCtx := model.SetUserToCtx(context.Background(), user)
				ectx.SetRequest(ectx.Request().WithContext(authCtx))

				res := &model.InitiateSDTestOutput{
					ID: uuid.New(),
				}

				inputWithCtx := &model.InitiateSDTestInput{
					UserID: uuid.NullUUID{UUID: user.UserID, Valid: true},
				}

				mockSDTestUc.EXPECT().Initiate(ectx.Request().Context(), inputWithCtx).Times(1).Return(res, &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "binding json failed",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/tests/", strings.NewReader(`
					{
						"request": {
							] , <- invalid here
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "empty request / nil",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/tests/", strings.NewReader(`
					{
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
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
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/tests/", strings.NewReader(`
					{
						"request": {
							"name": "name",
							"indicationThreshold": 10,
							"positiveIndicationText": "pos",
							"negativeIndicationText": "neg",
							"subGroupDetails": [
								{
									"name": "ok",
									"questionCount": 99,
									"answerOptionCount": 12
								}
							]
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockSDTestUc.EXPECT().Initiate(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
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
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/tests/", strings.NewReader(`
					{
						"request": {
							"name": "name",
							"indicationThreshold": 10,
							"positiveIndicationText": "pos",
							"negativeIndicationText": "neg",
							"subGroupDetails": [
								{
									"name": "ok",
									"questionCount": 99,
									"answerOptionCount": 12
								}
							]
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "hmz",
					Cause:   errors.New("err apajalah"),
					Code:    http.StatusBadRequest,
					Type:    usecase.ErrSDTemplateInputInvalid,
				}

				mockSDTestUc.EXPECT().Initiate(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleInitiateSDTest()(ectx)
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
