package rest

import (
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

func TestRest_handleCreateSDTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDTemplateUc := mock.NewMockSDTemplateUsecase(ctrl)

	input := &model.SDTemplate{
		Name:                   "name",
		IndicationThreshold:    10,
		PositiveIndiationText:  "pos",
		NegativeIndicationText: "neg",
		SubGroupDetails: []model.SDTemplateSubGroupDetail{
			{
				Name:              "ok",
				QuestionCount:     99,
				AnswerOptionCount: 12,
			},
		},
	}

	now := time.Now().UTC()
	template := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      input.Name,
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  input,
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
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

				mockSDTemplateUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(template.ToRESTResponse(), &common.Error{
					Type: nil,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    template.ToRESTResponse(),
				}, nil).Times(1).Return(nil)

				err := restService.handleCreateSDTemplate()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
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
							] , <- invalid here
						},
						"signature": "sig"
					}
				`))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDTemplate()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
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

				err := restService.handleCreateSDTemplate()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
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

				mockSDTemplateUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(nil, &common.Error{
					Type: usecase.ErrInternal,
				})

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDTemplate()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPost, "/sdt/templates/", strings.NewReader(`
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

				mockSDTemplateUc.EXPECT().Create(ectx.Request().Context(), input).Times(1).Return(nil, cerr)

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleCreateSDTemplate()(ectx)
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

func TestRest_handleFindSDTemplateByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAPIRespGen := httpMock.NewMockAPIResponseGenerator(ctrl)
	mockSDTemplateUc := mock.NewMockSDTemplateUsecase(ctrl)

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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDTemplateByID()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues("obviously invalid, right?")

				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrBadRequest.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDTemplateByID()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
				}
				req := httptest.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", "application/json")

				id := uuid.New()

				rec := httptest.NewRecorder()
				ectx := e.NewContext(req, rec)
				ectx.SetParamNames("id")
				ectx.SetParamValues(id.String())

				mockSDTemplateUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(nil, &common.Error{
					Message: "err internal",
					Cause:   errors.New("err internal"),
					Code:    http.StatusInternalServerError,
					Type:    usecase.ErrInternal,
				})
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, ErrInternal.GenerateStdlibHTTPResponse(nil), nil)

				err := restService.handleFindSDTemplateByID()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
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

				mockSDTemplateUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(nil, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, cerr.GenerateStdlibHTTPResponse(nil), nil).Times(1).Return(nil)

				err := restService.handleFindSDTemplateByID()(ectx)
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
					sdtemplateUsecase:    mockSDTemplateUc,
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

				res := &model.GeneratedSDTemplate{}

				mockSDTemplateUc.EXPECT().FindByID(ectx.Request().Context(), id).Times(1).Return(res, cerr)
				mockAPIRespGen.EXPECT().GenerateEchoAPIResponse(ectx, &stdhttp.StandardResponse{
					Success: true,
					Message: "success",
					Status:  http.StatusOK,
					Data:    res,
				}, nil).Times(1).Return(nil)

				err := restService.handleFindSDTemplateByID()(ectx)
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
