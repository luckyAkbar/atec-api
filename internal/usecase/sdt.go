package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gopkg.in/guregu/null.v4"
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

	return sdtest.ToInitiateSDTestOutput(submitKeyPlain, pack.Name, pack.Package.RenderTestQuestions()), nilErr
}

func (uc *sdtrUc) Submit(ctx context.Context, input *model.SubmitSDTestInput) (*model.SubmitSDTestOutput, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, &common.Error{
			Message: fmt.Sprintf("invalid input to submit test answer: %s", err.Error()),
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrInvalidSDTestAnswer,
		}
	}

	testData, err := uc.sdtrRepo.FindByID(ctx, input.TestID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd test by id")
		return nil, &common.Error{
			Message: "failed to find sd test data",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "sd test not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if testData.UserID.Valid {
		requester := model.GetUserFromCtx(ctx)
		if requester == nil || testData.UserID.UUID != requester.UserID {
			return nil, &common.Error{
				Message: "forbidden to submit other people sd test answer",
				Cause:   errors.New("forbidden to submit other people sd test answer"),
				Code:    http.StatusForbidden,
				Type:    ErrForbiddenToSubmitSDTestAnswer,
			}
		}
	}

	if uc.sharedCryptor.ReverseSecureToken(input.SubmitKey) != testData.SubmitKey {
		return nil, &common.Error{
			Message: "invalid submit key",
			Cause:   errors.New("invalid submit key"),
			Code:    http.StatusBadRequest,
			Type:    ErrInvalidSubmitKey,
		}
	}

	if err := testData.IsStillAcceptingAnswer(); err != nil {
		return nil, &common.Error{
			Message: err.Error(),
			Cause:   err,
			Code:    http.StatusForbidden,
			Type:    ErrForbiddenToSubmitSDTestAnswer,
		}
	}

	pack, err := uc.sdpRepo.FindByID(ctx, testData.PackageID, false)
	switch err {
	default:
		return nil, &common.Error{
			Message: "failed to find sd package data",
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

	grade, err := input.Answers.DoGradingProcess(pack.Package)
	if err != nil {
		return nil, &common.Error{
			Message: fmt.Sprintf("test answer are invalid. details: %s", err.Error()),
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrInvalidSDTestAnswer,
		}
	}

	testData.Result = model.SDTestResult{
		Result: grade,
	}
	testData.Answer = *input.Answers
	now := time.Now().UTC()
	testData.UpdatedAt = now
	testData.FinishedAt = null.NewTime(now, true)
	if err := uc.sdtrRepo.Update(ctx, testData, nil); err != nil {
		return nil, &common.Error{
			Message: "failed to save test result",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return testData.ToSubmitTestOutput(pack.Name, input.SubmitKey, pack.Package.RenderTestQuestions()), nilErr
}

func (uc *sdtrUc) Histories(ctx context.Context, input *model.ViewHistoriesInput) ([]model.ViewHistoriesOutput, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtrUc.Histories",
		"input": helper.Dump(input),
	})

	searchInput := input
	requester := model.GetUserFromCtx(ctx)
	if !requester.IsAdmin() {
		searchInput.UserID = uuid.NullUUID{UUID: requester.UserID, Valid: true}
	}

	res, err := uc.sdtrRepo.Search(ctx, searchInput)
	if err != nil {
		logger.WithError(err).Error("failed to search sd test histories")
		return nil, &common.Error{
			Message: "failed to search sd test histories",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	resp := []model.ViewHistoriesOutput{}
	for _, v := range res {
		resp = append(resp, v.ToViewHistoriesOutput())
	}

	return resp, nilErr
}

func (uc *sdtrUc) Statistic(ctx context.Context, userID uuid.UUID) ([]model.SDTestStatistic, *common.Error) {
	logger := logrus.WithFields(logrus.Fields{
		"func":  "sdtrUc.Statistic",
		"input": userID,
	})

	requester := model.GetUserFromCtx(ctx)
	if !requester.IsAdmin() {
		userID = requester.UserID
	}

	res, err := uc.sdtrRepo.Statistic(ctx, userID)
	switch err {
	default:
		logger.WithError(err).Error("failed to get sd test statistic")
		return nil, &common.Error{
			Message: "failed to get sd test statistic",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "no statistic found for this user",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		return res, nilErr
	}
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
