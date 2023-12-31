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

func TestRest_handleSubmitSDTestAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDTestUc := mock.NewMockSDTestUsecase(ctrl)

	input := &model.SubmitSDTestInput{
		TestID:    uuid.MustParse("f5849b4c-a92e-4d6e-a578-909445c17996"),
		SubmitKey: "key",
		Answers: &model.SDTestAnswer{TestAnswers: []*model.TestAnswer{
			{
				GroupName: "test",
				Answers: []model.Answer{
					{
						Question: "test1",
						Answer:   "test2",
					},
				},
			},
		}},
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
					sdtestUsecase:        mockSDTestUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
					{
						"request": {
							"testID": "f5849b4c-a92e-4d6e-a578-909445c17996",
							"submitKey": "key",
							"answers": {
								"testAnswers": [
									{
										"groupName": "test",
										"answers": [
											{
												"question": "test1",
												"answer": "test2"
											}
										]
									}
								]
							}
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				res := &model.SubmitSDTestOutput{
					ID: uuid.New(),
				}

				mockSDTestUc.EXPECT().Submit(ectx.Request().Context(), input).Times(1).Return(res, &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleSubmitSDTestAnswer()(ectx)
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

				err := restService.handleSubmitSDTestAnswer()(ectx)
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

				err := restService.handleSubmitSDTestAnswer()(ectx)
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
							"testID": "f5849b4c-a92e-4d6e-a578-909445c17996",
							"submitKey": "key",
							"answers": {
								"testAnswers": [
									{
										"groupName": "test",
										"answers": [
											{
												"question": "test1",
												"answer": "test2"
											}
										]
									}
								]
							}
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockSDTestUc.EXPECT().Submit(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleSubmitSDTestAnswer()(ectx)
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
							"testID": "f5849b4c-a92e-4d6e-a578-909445c17996",
							"submitKey": "key",
							"answers": {
								"testAnswers": [
									{
										"groupName": "test",
										"answers": [
											{
												"question": "test1",
												"answer": "test2"
											}
										]
									}
								]
							}
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

				mockSDTestUc.EXPECT().Submit(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleSubmitSDTestAnswer()(ectx)
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

func TestRest_handleViewSDTestHistories(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	sdtUc := mock.NewMockSDTestUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name: "ok",
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				input := &model.ViewHistoriesInput{
					Limit:  10,
					Offset: 100,
				}

				resp := []model.ViewHistoriesOutput{}
				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")

				sdtUc.EXPECT().Histories(ectx.Request().Context(), input).Times(1).Return(resp, &common.Error{Type: nil})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    resp,
				}, nil)

				err := restService.handleViewSDTestHistories()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				ectx.QueryParams().Add("userID", "hamdeh.")

				sdtUc.EXPECT().Histories(ectx.Request().Context(), &model.ViewHistoriesInput{}).Times(1).Return([]model.ViewHistoriesOutput{}, &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    []model.ViewHistoriesOutput{},
				}, nil)
				err := restService.handleViewSDTestHistories()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()
				input := &model.ViewHistoriesInput{
					Limit:  10,
					Offset: 100,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("createdBy", id.String())
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")

				cerr := &common.Error{Type: usecase.ErrInternal}

				sdtUc.EXPECT().Histories(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleViewSDTestHistories()(ectx)
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

func TestRest_handleGetSDTestStatistic(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	sdtUc := mock.NewMockSDTestUsecase(ctrl)

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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("user_id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleGetSDTestStatistic()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("user_id")
				ectx.SetParamValues("obviously invalid, right?")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleGetSDTestStatistic()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("user_id")
				ectx.SetParamValues(id.String())

				sdtUc.EXPECT().Statistic(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Message: "err internal",
					Cause:   errors.New("err internal"),
					Code:    http.StatusInternalServerError,
					Type:    usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleGetSDTestStatistic()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("user_id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Code: http.StatusBadRequest,
					Type: usecase.ErrInputResetPasswordInvalid,
				}

				sdtUc.EXPECT().Statistic(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleGetSDTestStatistic()(ectx)
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
					sdtestUsecase:        sdtUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("user_id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Type: nil,
				}

				res := []model.SDTestStatistic{}

				sdtUc.EXPECT().Statistic(ectx.Request().Context(), id).Times(1).Return(res, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleGetSDTestStatistic()(ectx)
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

func TestRest_handleDownloadTestResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	sdt := mock.NewMockSDTestUsecase(ctrl)

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
					sdtestUsecase:        sdt,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDownloadTestResult()(ectx)
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
					sdtestUsecase:        sdt,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("obviously invalid, right?")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDownloadTestResult()(ectx)
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
					sdtestUsecase:        sdt,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				sdt.EXPECT().DownloadResult(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Message: "err internal",
					Cause:   errors.New("err internal"),
					Code:    http.StatusInternalServerError,
					Type:    usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDownloadTestResult()(ectx)
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
					sdtestUsecase:        sdt,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
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

				sdt.EXPECT().DownloadResult(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleDownloadTestResult()(ectx)
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
					sdtestUsecase:        sdt,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Type: nil,
				}

				res := &model.ImageResult{}

				sdt.EXPECT().DownloadResult(ectx.Request().Context(), id).Times(1).Return(res, cerr)

				err := restService.handleDownloadTestResult()(ectx)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
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
