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
)

type sdpUc struct {
	sdpRepo model.SDPackageRepository
	sdtRepo model.SDTemplateRepository
}

// NewSDPackageUsecase will create new an sdpUc object representation of model.SDPackageUsecase interface
func NewSDPackageUsecase(sdpRepo model.SDPackageRepository, sdtRepo model.SDTemplateRepository) model.SDPackageUsecase {
	return &sdpUc{
		sdpRepo: sdpRepo,
		sdtRepo: sdtRepo,
	}
}

func (uc *sdpUc) Create(ctx context.Context, input *model.SDPackage) (*model.GeneratedSDPackage, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpuc.Create",
		"input": helper.Dump(input),
	})

	if err := input.PartialValidation(); err != nil {
		return nil, &common.Error{
			Message: fmt.Sprintf("invalid input for create SD package because: %s", err.Error()),
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrSDPackageInputInvalid,
		}
	}

	template, err := uc.sdtRepo.FindByID(ctx, input.TemplateID, false)
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd template by id")
		return nil, &common.Error{
			Message: "failed to find sd template by id",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "sd template not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if !template.IsActive {
		return nil, &common.Error{
			Message: "sd template is not active",
			Cause:   err,
			Code:    http.StatusForbidden,
			Type:    ErrSDTemplateIsDeactivated,
		}
	}

	requester := model.GetUserFromCtx(ctx)
	now := time.Now().UTC()
	sdpackage := &model.SpeechDelayPackage{
		ID:         uuid.New(),
		TemplateID: input.TemplateID,
		Name:       input.PackageName,
		CreatedBy:  requester.UserID,
		Package:    input,
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.sdpRepo.Create(ctx, sdpackage); err != nil {
		logger.WithError(err).Error("failed to create sd package")
		return nil, &common.Error{
			Message: "failed to create sd package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return sdpackage.ToRESTResponse(), nilErr
}

func (uc *sdpUc) FindByID(ctx context.Context, id uuid.UUID) (*model.GeneratedSDPackage, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdpuc.FindByID",
		"id":   id.String(),
	})

	pack, err := uc.sdpRepo.FindByID(ctx, id, true)
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd package")
		return nil, &common.Error{
			Message: "failed to find sd package",
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
		return pack.ToRESTResponse(), nilErr
	}
}

func (uc *sdpUc) Search(ctx context.Context, input *model.SearchSDPackageInput) (*model.SearchPackageOutput, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpuc.Search",
		"input": helper.Dump(input),
	})

	res, err := uc.sdpRepo.Search(ctx, input)
	if err != nil {
		logger.WithError(err).Error("failed to find sd package")
		return nil, &common.Error{
			Message: "failed to find sd package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	response := []*model.GeneratedSDPackage{}
	for _, v := range res {
		response = append(response, v.ToRESTResponse())
	}

	return &model.SearchPackageOutput{
		Packages: response,
		Count:    len(res),
	}, nilErr
}

func (uc *sdpUc) Update(ctx context.Context, id uuid.UUID, input *model.SDPackage) (*model.GeneratedSDPackage, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtUc.Update",
		"id":    id.String(),
		"input": helper.Dump(input),
	})

	if err := input.PartialValidation(); err != nil {
		return nil, &common.Error{
			Message: err.Error(),
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrSDPackageInputInvalid,
		}
	}

	template, err := uc.sdtRepo.FindByID(ctx, input.TemplateID, false)
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd template")
		return nil, &common.Error{
			Message: "failed to find speech delay template",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "speech delay template not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if !template.IsActive {
		return nil, &common.Error{
			Message: "speech delay template is not active",
			Cause:   errors.New("speech delay template is not active"),
			Code:    http.StatusForbidden,
			Type:    ErrSDTemplateIsDeactivated,
		}

	}

	pack, err := uc.sdpRepo.FindByID(ctx, id, false)
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd package")
		return nil, &common.Error{
			Message: "failed to find speech delay package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "speech delay package not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if pack.IsLocked {
		return nil, &common.Error{
			Message: "speech delay package is already locked",
			Cause:   errors.New("speech delay package is already locked"),
			Code:    http.StatusForbidden,
			Type:    ErrSDPackageAlreadyLocked,
		}
	}

	pack.UpdatedAt = time.Now().UTC()
	pack.Name = input.PackageName
	pack.TemplateID = input.TemplateID
	pack.Package = input

	if err := uc.sdpRepo.Update(ctx, pack, nil); err != nil {
		logger.WithError(err).Error("failed to update speech delay package")
		return nil, &common.Error{
			Message: "failed to update speech delay package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return pack.ToRESTResponse(), nilErr
}

func (uc *sdpUc) Delete(ctx context.Context, id uuid.UUID) (*model.GeneratedSDPackage, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdpUc.Delete",
		"id":   id.String(),
	})

	pack, err := uc.sdpRepo.FindByID(ctx, id, false)
	switch err {
	default:
		logger.WithError(err).Error("failed to find speech delay package")
		return nil, &common.Error{
			Message: "failed to find speech delay package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "speech delay package not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if pack.IsLocked {
		return nil, &common.Error{
			Message: "speech delay package is already locked",
			Cause:   errors.New("speech delay package is already locked"),
			Code:    http.StatusForbidden,
			Type:    ErrSDPackageAlreadyLocked,
		}
	}

	deleted, err := uc.sdpRepo.Delete(ctx, pack.ID)
	if err != nil {
		return nil, &common.Error{
			Message: "failed to delete speech delay package",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return deleted.ToRESTResponse(), nilErr
}
