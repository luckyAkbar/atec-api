package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type sdpRepo struct {
	db *gorm.DB
}

// NewSDPackageRepository will create new an sdpRepo object representation of model.SDPackageRepository interface
func NewSDPackageRepository(db *gorm.DB) model.SDPackageRepository {
	return &sdpRepo{
		db: db,
	}
}

func (r *sdpRepo) Create(ctx context.Context, input *model.SpeechDelayPackage) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Create",
		"input": helper.Dump(input),
	})
	if err := r.db.WithContext(ctx).Create(input).Error; err != nil {
		logger.WithError(err).Error("failed to create sd package")
		return err
	}

	return nil
}

func (r *sdpRepo) FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdpRepo.FindByID",
		"id":   id.String(),
	})

	query := r.db.WithContext(ctx)
	if includeDeleted {
		query = query.Unscoped()
	}

	sdp := &model.SpeechDelayPackage{}
	err := query.Take(sdp, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd package")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return sdp, nil
	}
}
