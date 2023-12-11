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
