package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *sdpRepo) Search(ctx context.Context, input *model.SearchSDPackageInput) ([]*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Search",
		"input": helper.Dump(input),
	})

	query := r.db.WithContext(ctx)
	if input.IncludeDeleted {
		query = query.Unscoped()
	}

	where, conds := input.ToWhereQuery()
	for i := 0; i < len(where); i++ {
		query = query.Where(where[i], conds[i])
	}

	var packs []*model.SpeechDelayPackage
	err := query.Limit(input.Limit).Offset(input.Offset).Order("created_at desc").Find(&packs).Error
	if err != nil {
		logger.WithError(err).Error("failed to search sd template from db")
		return []*model.SpeechDelayPackage{}, err
	}

	return packs, nil
}

func (r *sdpRepo) Update(ctx context.Context, pack *model.SpeechDelayPackage, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":    "sdpRepo.Update",
		"package": helper.Dump(pack),
	})

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Save(pack).Error
	if err != nil {
		logger.WithError(err).Error("failed to update speech delay package")
		return err
	}

	return nil
}

func (r *sdpRepo) Delete(ctx context.Context, id uuid.UUID) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Delete",
		"input": helper.Dump(id),
	})

	deleted := &model.SpeechDelayPackage{}
	err := r.db.WithContext(ctx).Clauses(clause.Returning{}).Delete(deleted, "id = ?", id).Error
	if err != nil {
		logger.WithError(err).Error("failed to delete speech delay template")
		return nil, err
	}

	return deleted, nil
}
