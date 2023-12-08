package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
)

type sdtUc struct {
	sdtRepo model.SDTemplateRepository
}

// NewSDTemplateUsecase create SDTemplateUsecase
func NewSDTemplateUsecase(sdtRepo model.SDTemplateRepository) model.SDTemplateUsecase {
	return &sdtUc{sdtRepo: sdtRepo}
}

func (uc *sdtUc) Create(ctx context.Context, input *model.SDTemplate) (*model.GeneratedSDTemplate, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtUc.Create",
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, &common.Error{
			Message: err.Error(),
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrSDTemplateInputInvalid,
		}
	}

	requester := model.GetUserFromCtx(ctx)
	now := time.Now().UTC()
	template := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: requester.UserID,
		Name:      input.Name,
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  input,
	}

	if err := uc.sdtRepo.Create(ctx, template); err != nil {
		logger.WithError(err).Error("failed to create template")
		return nil, &common.Error{
			Message: "failed to create template",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return template.ToRESTResponse(), nilErr
}
