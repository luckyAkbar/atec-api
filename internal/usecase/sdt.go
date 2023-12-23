package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type sdtrUc struct {
	sdtrRepo      model.SDTestRepository
	sdpRepo       model.SDPackageRepository
	sharedCryptor common.SharedCryptor
	tx            *gorm.DB
}

// NewSDTestResultUsecase create new sd test usecase. satisfy model.SDTestUsecase
func NewSDTestResultUsecase(sdtrRepo model.SDTestRepository, sdpRepo model.SDPackageRepository, sharedCryptor common.SharedCryptor, tx *gorm.DB) model.SDTestUsecase {
	return &sdtrUc{
		sdtrRepo:      sdtrRepo,
		sdpRepo:       sdpRepo,
		sharedCryptor: sharedCryptor,
		tx:            tx,
	}
}

func (uc *sdtrUc) Initiate(ctx context.Context, input *model.InitiateSDTestInput) (*model.InitiateSDTestOutput, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtrUc.Initiate",
		"input": helper.Dump(input),
	})

	if input.DurationMinutes == 0 {
		input.DurationMinutes = time.Minute * 60
	}

	pack, cerr := uc.validateAndFetchPackageID(ctx, input.PackageID, input.UserID)
	if cerr.Type != nil {
		logger.WithError(cerr.Cause).Error("failed to fetch sd package to initiate sd test: ", cerr.Message)
		return nil, cerr
	}

	// safety check
	if !pack.IsActive {
		return nil, &common.Error{
			Message: "sd package is not active",
			Cause:   errors.New("sd package is not active"),
			Code:    http.StatusBadRequest,
			Type:    ErrSDPackageAlreadyDeactivated,
		}
	}

	dbTrx := uc.tx.Begin()

	if !pack.IsLocked {
		pack.IsLocked = true
		pack.UpdatedAt = time.Now().UTC()
		if err := uc.sdpRepo.Update(ctx, pack, dbTrx); err != nil {
			dbTrx.Rollback()
			return nil, &common.Error{
				Message: "failed to lock sd package",
				Cause:   err,
				Code:    http.StatusInternalServerError,
				Type:    ErrInternal,
			}
		}
	}

	submitKeyPlain, submitKeyEnc, err := uc.sharedCryptor.CreateSecureToken()
	if err != nil {
		dbTrx.Rollback()
		return nil, &common.Error{
			Message: "failed to create submit key",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	sdtest := &model.SDTest{
		ID:        uuid.New(),
		PackageID: pack.ID,
		UserID:    input.UserID,
		OpenUntil: time.Now().UTC().Add(input.DurationMinutes),
		SubmitKey: submitKeyEnc,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.sdtrRepo.Create(ctx, sdtest, dbTrx); err != nil {
		dbTrx.Rollback()
		return nil, &common.Error{
			Message: "failed to create sd test",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	dbTrx.Commit()

	return sdtest.ToRESTResponse(submitKeyPlain, pack.Name, pack.Package.RenderTestQuestions()), nilErr
}

func (uc *sdtrUc) validateAndFetchPackageID(ctx context.Context, packageID, userID uuid.NullUUID) (*model.SpeechDelayPackage, *common.Error) {
	if packageID.Valid {
		pack, err := uc.sdpRepo.FindByID(ctx, packageID.UUID, false)
		switch err {
		default:
			return nil, &common.Error{
				Message: "failed to fetch sd package",
				Cause:   err,
				Code:    http.StatusInternalServerError,
				Type:    ErrInternal,
			}
		case repository.ErrNotFound:
			return nil, &common.Error{
				Message: "no package found",
				Cause:   err,
				Code:    http.StatusNotFound,
				Type:    ErrResourceNotFound,
			}
		case nil:
			return pack, nilErr
		}
	}

	if !userID.Valid {
		pack, err := uc.sdpRepo.FindRandomActivePackage(ctx)
		if err != nil {
			return nil, &common.Error{
				Message: "failed to fetch sd package",
				Cause:   err,
				Code:    http.StatusInternalServerError,
				Type:    ErrInternal,
			}
		}

		return pack, nilErr
	}

	packID, err := uc.sdpRepo.FindLeastUsedPackageIDByUserID(ctx, userID.UUID)
	switch err {
	default:
		return nil, &common.Error{
			Message: "failed to fetch sd package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "sd package not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	pack, err := uc.sdpRepo.FindByID(ctx, packID, false)
	if err != nil {
		return nil, &common.Error{
			Message: "failed to fetch sd package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return pack, nilErr
}
