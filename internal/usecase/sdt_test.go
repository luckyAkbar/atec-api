package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/stretchr/testify/assert"
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
