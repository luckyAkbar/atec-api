package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestSDTestUsecase_Initiate(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	sdtrRepo := mock.NewMockSDTestRepository(kit.Ctrl)
	sdpRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	sharedCryptor := commonMock.NewMockSharedCryptor(kit.Ctrl)

	ctx := context.Background()
	db := kit.DB
	mockDB := kit.DBmock

	inputPackageID := uuid.New()
	userID := uuid.New()
	pack := &model.SpeechDelayPackage{
		ID:       inputPackageID,
		IsActive: true,
		IsLocked: true,
		Package:  &model.SDPackage{},
	}

	uc := NewSDTestResultUsecase(sdtrRepo, sdpRepo, sharedCryptor, db)

	tests := []common.TestStructure{
		{
			Name: "using defined package id, but repository return not found",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "db error when fetching a defined package id",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "when using defined package id, somehow got inactive package",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       inputPackageID,
					IsActive: false,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageAlreadyDeactivated)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
			},
		},
		{
			Name: "when using defined package id, failure on locking the package must return internal error",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       inputPackageID,
					IsActive: true,
					IsLocked: false,
				}, nil)
				mockDB.ExpectBegin()
				sdpRepo.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("err db"))
				mockDB.ExpectRollback()
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "when using defined package id, success locking but failed when creating submit key must result in internal error",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       inputPackageID,
					IsActive: true,
					IsLocked: false,
				}, nil)
				mockDB.ExpectBegin()
				sdpRepo.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				sharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("", "", errors.New("err db"))
				mockDB.ExpectRollback()
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "when using defined package id, no need to lock the package but fails when creating sd test record",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       inputPackageID,
					IsActive: true,
					IsLocked: true,
				}, nil)
				mockDB.ExpectBegin()
				sharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				sdtrRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("err"))
				mockDB.ExpectRollback()
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "when using defined package id: ok",
			MockFn: func() {
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       inputPackageID,
					Package:  &model.SDPackage{},
					IsActive: true,
					IsLocked: true,
				}, nil)
				mockDB.ExpectBegin()
				sharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				sdtrRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockDB.ExpectCommit()
			},
			Run: func() {
				res, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					PackageID: uuid.NullUUID{UUID: inputPackageID, Valid: true},
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.PackageID, inputPackageID)
				assert.Equal(t, res.SubmitKey, "plain")
			},
		},
		{
			Name: "used by unregistered user must using random active package, when got any error must returning err internal",
			MockFn: func() {
				sdpRepo.EXPECT().FindRandomActivePackage(ctx).Times(1).Return(nil, errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "used by unregistered user must using random active package: ok",
			MockFn: func() {
				sdpRepo.EXPECT().FindRandomActivePackage(ctx).Times(1).Return(pack, nil)
				mockDB.ExpectBegin()
				sharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				sdtrRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockDB.ExpectCommit()
			},
			Run: func() {
				res, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.PackageID, inputPackageID)
				assert.Equal(t, res.SubmitKey, "plain")
			},
		},
		{
			Name: "used by registered user and not specifying any package id: got error from db",
			MockFn: func() {
				sdpRepo.EXPECT().FindLeastUsedPackageIDByUserID(ctx, userID).Times(1).Return(uuid.Nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					UserID: uuid.NullUUID{UUID: userID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "used by registered user and not specifying any package id: somehow got not found hmz",
			MockFn: func() {
				sdpRepo.EXPECT().FindLeastUsedPackageIDByUserID(ctx, userID).Times(1).Return(uuid.Nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					UserID: uuid.NullUUID{UUID: userID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "used by registered user and not specifying any package id: but got error when finding by id??",
			MockFn: func() {
				sdpRepo.EXPECT().FindLeastUsedPackageIDByUserID(ctx, userID).Times(1).Return(inputPackageID, nil)
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					UserID: uuid.NullUUID{UUID: userID, Valid: true},
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "used by registered user and not specifying any package id: ok",
			MockFn: func() {
				sdpRepo.EXPECT().FindLeastUsedPackageIDByUserID(ctx, userID).Times(1).Return(inputPackageID, nil)
				sdpRepo.EXPECT().FindByID(ctx, inputPackageID, false).Times(1).Return(pack, nil)
				mockDB.ExpectBegin()
				sharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				sdtrRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockDB.ExpectCommit()
			},
			Run: func() {
				res, cerr := uc.Initiate(ctx, &model.InitiateSDTestInput{
					UserID: uuid.NullUUID{UUID: userID, Valid: true},
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.PackageID, inputPackageID)
				assert.Equal(t, res.SubmitKey, "plain")
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

func TestSDTestUsecase_Submit(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	sdtrRepo := mock.NewMockSDTestRepository(kit.Ctrl)
	sdpRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	sharedCryptor := commonMock.NewMockSharedCryptor(kit.Ctrl)

	ctx := context.Background()
	db := kit.DB
	tid := uuid.New()
	packID := uuid.New()

	user := model.AuthUser{
		UserID:      uuid.New(),
		AccessToken: "token",
		Role:        model.RoleAdmin,
	}

	authCtx := model.SetUserToCtx(ctx, user)

	uc := NewSDTestResultUsecase(sdtrRepo, sdpRepo, sharedCryptor, db)

	tests := []common.TestStructure{
		{
			Name:   "input invalid",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInvalidSDTestAnswer)
			},
		},
		{
			Name:   "input invalid",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "",
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInvalidSDTestAnswer)
			},
		},
		{
			Name:   "input invalid",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers:   &model.SDTestAnswer{},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInvalidSDTestAnswer)
			},
		},
		{
			Name: "db failed when trying to find sd test result by id",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(ctx, tid).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "test data not found",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(ctx, tid).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "the submitter id set, but submitted by a non registered user",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(ctx, tid).Times(1).Return(&model.SDTest{
					UserID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Submit(ctx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
				assert.Equal(t, cerr.Type, ErrForbiddenToSubmitSDTestAnswer)
			},
		},
		{
			Name: "the submitter id set, but submitted by different user",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
				assert.Equal(t, cerr.Type, ErrForbiddenToSubmitSDTestAnswer)
			},
		},
		{
			Name: "submit key invalid",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("different")
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInvalidSubmitKey)
			},
		},
		{
			Name: "test data is already pass the OpenUntil",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * -1).UTC(),
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
				assert.Equal(t, cerr.Type, ErrForbiddenToSubmitSDTestAnswer)
			},
		},
		{
			Name: "test data's FinishedAt is already set",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:     uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey:  "submitkeyenc",
					OpenUntil:  time.Now().Add(time.Hour * 1).UTC(),
					FinishedAt: null.NewTime(time.Now().UTC(), true),
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
				assert.Equal(t, cerr.Type, ErrForbiddenToSubmitSDTestAnswer)
			},
		},
		{
			Name: "failure on finding package id from db",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * 1).UTC(),
					PackageID: packID,
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
				sdpRepo.EXPECT().FindByID(authCtx, packID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "somehow the package was not found",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * 1).UTC(),
					PackageID: packID,
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
				sdpRepo.EXPECT().FindByID(authCtx, packID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		// error on grading process may not be detailed here. but must be on the model package
		{
			Name: "any failure on grading process",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * 1).UTC(),
					PackageID: packID,
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
				sdpRepo.EXPECT().FindByID(authCtx, packID, false).Times(1).Return(&model.SpeechDelayPackage{
					Package: &model.SDPackage{
						PackageName: "testing",
						TemplateID:  uuid.New(),
						SubGroupDetails: []model.SDSubGroupDetail{
							{
								Name: "test1",
								QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
									{
										Question: "testing?",
										AnswersAndValue: []model.SDAnswerAndValue{
											{
												Text:  "iya",
												Value: 1,
											},
											{
												Text:  "nope",
												Value: 2,
											},
										},
									},
								},
							},
						},
					},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test",
								Answers: []model.Answer{
									{
										Question: "test",
										Answer:   "test",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInvalidSDTestAnswer)
				assert.Equal(t, cerr.Message, "test answer are invalid. details: group test1 is not found on answers list")
			},
		},
		{
			Name: "failed updating the sd test data",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * 1).UTC(),
					PackageID: packID,
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
				sdpRepo.EXPECT().FindByID(authCtx, packID, false).Times(1).Return(&model.SpeechDelayPackage{
					Package: &model.SDPackage{
						PackageName: "testing",
						TemplateID:  uuid.New(),
						SubGroupDetails: []model.SDSubGroupDetail{
							{
								Name: "test1",
								QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
									{
										Question: "testing?",
										AnswersAndValue: []model.SDAnswerAndValue{
											{
												Text:  "iya",
												Value: 1,
											},
											{
												Text:  "nope",
												Value: 2,
											},
										},
									},
								},
							},
						},
					},
				}, nil)
				sdtrRepo.EXPECT().Update(authCtx, gomock.Any(), nil).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test1",
								Answers: []model.Answer{
									{
										Question: "testing?",
										Answer:   "iya",
									},
								},
							},
						},
					},
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				sdtrRepo.EXPECT().FindByID(authCtx, tid).Times(1).Return(&model.SDTest{
					UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
					SubmitKey: "submitkeyenc",
					OpenUntil: time.Now().Add(time.Hour * 1).UTC(),
					PackageID: packID,
				}, nil)
				sharedCryptor.EXPECT().ReverseSecureToken("valid").Times(1).Return("submitkeyenc")
				sdpRepo.EXPECT().FindByID(authCtx, packID, false).Times(1).Return(&model.SpeechDelayPackage{
					Package: &model.SDPackage{
						PackageName: "testing",
						TemplateID:  uuid.New(),
						SubGroupDetails: []model.SDSubGroupDetail{
							{
								Name: "test1",
								QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
									{
										Question: "testing?",
										AnswersAndValue: []model.SDAnswerAndValue{
											{
												Text:  "iya",
												Value: 1,
											},
											{
												Text:  "nope",
												Value: 2,
											},
										},
									},
								},
							},
						},
					},
				}, nil)
				sdtrRepo.EXPECT().Update(authCtx, gomock.Any(), nil).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.Submit(authCtx, &model.SubmitSDTestInput{
					TestID:    tid,
					SubmitKey: "valid",
					Answers: &model.SDTestAnswer{
						TestAnswers: []*model.TestAnswer{
							{
								GroupName: "test1",
								Answers: []model.Answer{
									{
										Question: "testing?",
										Answer:   "iya",
									},
								},
							},
						},
					},
				})
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Result.Result[0].GroupName, "test1")
				assert.Equal(t, res.Result.Result[0].Result, 1)
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

func TestSDTestUsecase_Histories(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	sdtrRepo := mock.NewMockSDTestRepository(kit.Ctrl)
	sdpRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	sharedCryptor := commonMock.NewMockSharedCryptor(kit.Ctrl)

	ctx := context.Background()
	uid := uuid.New()
	adminCtx := model.SetUserToCtx(ctx, model.AuthUser{
		UserID:      uuid.New(),
		AccessToken: "token",
		Role:        model.RoleAdmin,
	})
	userCtx := model.SetUserToCtx(ctx, model.AuthUser{
		UserID:      uid,
		AccessToken: "token",
		Role:        model.RoleUser,
	})
	db := kit.DB
	randUserID := uuid.New()
	tid := uuid.New()
	pid := uuid.New()
	now := time.Now().UTC()

	uc := NewSDTestResultUsecase(sdtrRepo, sdpRepo, sharedCryptor, db)

	tests := []common.TestStructure{
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID: uuid.NullUUID{UUID: randUserID, Valid: true},
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID: uuid.NullUUID{UUID: randUserID, Valid: true},
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID:    uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID: uuid.NullUUID{UUID: pid, Valid: true},
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID:    uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID: uuid.NullUUID{UUID: pid, Valid: true},
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID:       uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:    uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter: null.NewTime(now, true),
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID:       uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:    uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter: null.NewTime(now, true),
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "admin trying to search",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(adminCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: randUserID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "non admin search param user id must be overriden",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uuid.New(), Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "non admin search param user id must be overriden",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return([]*model.SDTest{
					{
						ID: tid,
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Histories(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "empty array?",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return(nil, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, len(res), 0)
			},
		},
		{
			Name: "empty array?",
			MockFn: func() {
				sdtrRepo.EXPECT().Search(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				}).Times(1).Return([]*model.SDTest{}, nil)
			},
			Run: func() {
				res, cerr := uc.Histories(userCtx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: uid, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(now, true),
					IncludeUnfinished: true,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})

				assert.NoError(t, cerr.Type)
				assert.Equal(t, len(res), 0)
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

func TestSDTestUsecase_Statistic(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	sdtrRepo := mock.NewMockSDTestRepository(kit.Ctrl)
	sdpRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	sharedCryptor := commonMock.NewMockSharedCryptor(kit.Ctrl)

	ctx := context.Background()
	uid := uuid.New()
	randUID := uuid.New()
	userCtx := model.SetUserToCtx(ctx, model.AuthUser{
		UserID: uid,
		Role:   model.RoleUser,
	})
	adminCtx := model.SetUserToCtx(ctx, model.AuthUser{
		Role: model.RoleAdmin,
	})

	uc := NewSDTestResultUsecase(sdtrRepo, sdpRepo, sharedCryptor, nil)

	tests := []common.TestStructure{
		{
			Name: "non admin should only be able to view his own statistic",
			MockFn: func() {
				sdtrRepo.EXPECT().Statistic(userCtx, uid).Times(1).Return([]model.SDTestStatistic{}, nil)
			},
			Run: func() {
				_, cerr := uc.Statistic(userCtx, uuid.New())
				assert.NoError(t, cerr.Type)
			},
		},
		{
			Name: "admin can use any user id",
			MockFn: func() {
				sdtrRepo.EXPECT().Statistic(adminCtx, randUID).Times(1).Return([]model.SDTestStatistic{}, nil)
			},
			Run: func() {
				_, cerr := uc.Statistic(adminCtx, randUID)
				assert.NoError(t, cerr.Type)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				sdtrRepo.EXPECT().Statistic(adminCtx, randUID).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Statistic(adminCtx, randUID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "not found on db",
			MockFn: func() {
				sdtrRepo.EXPECT().Statistic(adminCtx, randUID).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Statistic(adminCtx, randUID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
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
