package rest

import (
	"errors"
	"fmt"
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

func TestRest_handleCreateSDPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDPUc := mock.NewMockSDPackageUsecase(ctrl)

	input := &model.SDPackage{
		PackageName: "testing package",
		TemplateID:  uuid.New(),
		SubGroupDetails: []model.SDSubGroupDetail{
			{
				Name: "valid name",
				QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
					{
						Question: "valid question?",
						AnswersAndValue: []model.SDAnswerAndValue{
							{
								Text:  "pilihan pertama",
								Value: 99,
							},
							{
								Text:  "pilihan kedua, tapi value nya sama",
								Value: 99,
							},
						},
					},
				},
			},
		},
	}

	now := time.Now().UTC()
	pack := &model.SpeechDelayPackage{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      input.PackageName,
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Package:   input,
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
					sdpackageUsecase:     mockSDPUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "testing package",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "valid name",
									"questionAndAnswerLists": [
										{
											"question": "valid question?",
											"answerAndValue": [
												{
													"text": "pilihan pertama",
													"value": 99
												},
												{
													"text": "pilihan kedua, tapi value nya sama",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, input.TemplateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockSDPUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(pack.ToRESTResponse(), &common.Error{Type: nil})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    pack.ToRESTResponse(),
				}, nil).Times(1).Return(nil)

				err := restService.handleCreateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPUc,
				}

				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "testing package",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "valid name",
									"questionAndAnswerLists": [
										{
											"question": "valid question?",
											"answerAndValue": [
												{
													"text": "pilihan pertama",
													"value": 99
												},
												{
													"text": "pilihan kedua, tapi value nya sama",
													"value": 99
												}
											] , <- invalid here
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, input.TemplateID.String())))

				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
					{
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "testing package",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "valid name",
									"questionAndAnswerLists": [
										{
											"question": "valid question?",
											"answerAndValue": [
												{
													"text": "pilihan pertama",
													"value": 99
												},
												{
													"text": "pilihan kedua, tapi value nya sama",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, input.TemplateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockSDPUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "testing package",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "valid name",
									"questionAndAnswerLists": [
										{
											"question": "valid question?",
											"answerAndValue": [
												{
													"text": "pilihan pertama",
													"value": 99
												},
												{
													"text": "pilihan kedua, tapi value nya sama",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, input.TemplateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				cerr := &common.Error{
					Message: "hmz",
					Cause:   errors.New("err apajalah"),
					Code:    http.StatusBadRequest,
					Type:    usecase.ErrSDPackageInputInvalid,
				}

				mockSDPUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDPackage()(ectx)
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

func TestRest_handleFindSDPackageByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDPackageUc := mock.NewMockSDPackageUsecase(ctrl)

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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDPackageByID()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("obviously invalid, right?")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDPackageByID()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockSDPackageUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Message: "err internal",
					Cause:   errors.New("err internal"),
					Code:    http.StatusInternalServerError,
					Type:    usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDPackageByID()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
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

				mockSDPackageUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleFindSDPackageByID()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
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

				res := &model.GeneratedSDPackage{}

				mockSDPackageUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(res, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleFindSDPackageByID()(ectx)
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

func TestRest_handleSearchSDPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDPackageUc := mock.NewMockSDPackageUsecase(ctrl)

	tests := []common.TestStructure{
		{
			Name: "ok",
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()
				input := &model.SearchSDPackageInput{
					CreatedBy: id,
					Limit:     10,
					Offset:    100,
				}

				resp := &model.SearchPackageOutput{
					Packages: []*model.GeneratedSDPackage{},
					Count:    10,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("createdBy", id.String())
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")

				mockSDPackageUc.EXPECT().Search(ectx.Request().Context(), input).Times(1).Return(resp, &common.Error{Type: nil})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    resp,
				}, nil)

				err := restService.handleSearchSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				ectx.QueryParams().Add("createdBy", "hamdeh.")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleSearchSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodGet, "/std/templates/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()
				input := &model.SearchSDPackageInput{
					CreatedBy: id,
					Limit:     10,
					Offset:    100,
				}

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.QueryParams().Add("createdBy", id.String())
				ectx.QueryParams().Add("limit", "10")
				ectx.QueryParams().Add("offset", "100")

				cerr := &common.Error{Type: usecase.ErrInternal}

				mockSDPackageUc.EXPECT().Search(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleSearchSDPackage()(ectx)
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

func TestRest_handleUpdateSDPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDPackageUc := mock.NewMockSDPackageUsecase(ctrl)

	templateID := uuid.New()
	validInput := &model.SDPackage{
		PackageName: "name",
		TemplateID:  templateID,
		SubGroupDetails: []model.SDSubGroupDetail{
			{
				Name: "okname",
				QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
					{
						Question: "question",
						AnswersAndValue: []model.SDAnswerAndValue{
							{
								Text:  "text",
								Value: 99,
							},
						},
					},
				},
			},
		},
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPut, "/sdt/packages/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "name",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "okname",
									"questionAndAnswerLists": [
										{
											"question": "question",
											"answerAndValue": [
												{
													"text": "text",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, templateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				res := &model.GeneratedSDPackage{}

				mockSDPackageUc.EXPECT().Update(ectx.Request().Context(), id, validInput).Times(1).Return(res, &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleUpdateSDPackage()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "invalid json",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPut, "/sdt/packages/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "name",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "okname",
									"questionAndAnswerLists": [
										{
											"question": "question",
											"answerAndValue": [
												{
													"text": "text",
													"value": 99
												}
											]
										}
									] <- invalid here
								}
							]
						},
						"signature": ""
					}
				`, templateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleUpdateSDPackage()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "invalid uuid",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPut, "/sdt/packages/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "name",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "okname",
									"questionAndAnswerLists": [
										{
											"question": "question",
											"answerAndValue": [
												{
													"text": "text",
													"value": 99
												}
											]
										}
									] <- invalid here
								}
							]
						},
						"signature": ""
					}
				`, templateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				ectx.SetParamNames("id")
				ectx.SetParamValues("invalid uuid")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleUpdateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPut, "/sdt/packages/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "name",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "okname",
									"questionAndAnswerLists": [
										{
											"question": "question",
											"answerAndValue": [
												{
													"text": "text",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, templateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{Type: usecase.ErrInternal}

				mockSDPackageUc.EXPECT().Update(ectx.Request().Context(), id, validInput).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleUpdateSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPut, "/sdt/packages/", strings.NewReader(fmt.Sprintf(`
					{
						"request": {
							"packageName": "name",
							"templateID": "%s",
							"subGroupDetails": [
								{
									"name": "okname",
									"questionAndAnswerLists": [
										{
											"question": "question",
											"answerAndValue": [
												{
													"text": "text",
													"value": 99
												}
											]
										}
									]
								}
							]
						},
						"signature": ""
					}
				`, templateID.String())))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				id := uuid.New()
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{Type: usecase.ErrSDPackageInputInvalid}

				mockSDPackageUc.EXPECT().Update(ectx.Request().Context(), id, validInput).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleUpdateSDPackage()(ectx)
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

func TestRest_handleDeleteSDPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDPackageUc := mock.NewMockSDPackageUsecase(ctrl)

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
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDeleteSDPackage()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "uc return err internal",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockSDPackageUc.EXPECT().Delete(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDeleteSDPackage()(ectx)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "uc return specific err",
			MockFn: func() {},
			Run: func() {
				e := echo.New()
				group := e.Group("")
				restService := service{
					rootGroup:            group,
					apiResponseGenerator: mockAPIRespGen,
					sdpackageUsecase:     mockSDPackageUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				cerr := &common.Error{
					Code: http.StatusForbidden,
					Type: usecase.ErrSDTemplateAlreadyLocked,
				}

				mockSDPackageUc.EXPECT().Delete(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleDeleteSDPackage()(ectx)
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
					sdpackageUsecase:     mockSDPackageUc,
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

				now := time.Now().UTC()
				p := &model.SpeechDelayPackage{
					ID:         uuid.New(),
					CreatedBy:  uuid.New(),
					Name:       "ne",
					IsActive:   false,
					IsLocked:   false,
					CreatedAt:  now,
					UpdatedAt:  now,
					TemplateID: uuid.New(),
					Package:    &model.SDPackage{},
				}

				mockSDPackageUc.EXPECT().Delete(ectx.Request().Context(), id).Times(1).Return(p.ToRESTResponse(), cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    p.ToRESTResponse(),
				}, nil)

				err := restService.handleDeleteSDPackage()(ectx)
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
