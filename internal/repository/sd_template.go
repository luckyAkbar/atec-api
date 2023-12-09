package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type sdRepo struct {
	db *gorm.DB
}

// NewSDTemplateRepository create new SDTemplateRepository
func NewSDTemplateRepository(db *gorm.DB) model.SDTemplateRepository {
	return &sdRepo{db}
}

func (r *sdRepo) Create(ctx context.Context, template *model.SpeechDelayTemplate) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdRepo.Create",
		"input": helper.Dump(template),
	})

	if err := r.db.WithContext(ctx).Create(template).Error; err != nil {
		logger.WithError(err).Error("failed to create test template")
		return err
	}

	return nil
}

func (r *sdRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SpeechDelayTemplate, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdRepo.FindByID",
		"input": helper.Dump(id),
	})

	template := &model.SpeechDelayTemplate{}
	err := r.db.WithContext(ctx).Unscoped().Take(template, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find test template by id")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return template, nil
	}
}
func (r *sdRepo) Search(ctx context.Context, input *model.SearchSDTemplateInput) ([]*model.SpeechDelayTemplate, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdRepo.Search",
		"input": helper.Dump(input),
	})

	var templates []*model.SpeechDelayTemplate

	query := r.db.WithContext(ctx)
	if input.IncludeDeleted {
		query = query.Unscoped()
	}

	where, conds := input.ToWhereQuery()
	for i := 0; i < len(where); i++ {
		query = query.Where(where[i], conds[i])
	}

	err := query.Limit(input.Limit).Offset(input.Offset).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		logger.WithError(err).Error("failed to search sd template from db")
		return []*model.SpeechDelayTemplate{}, err
	}

	return templates, nil
}

func (r *sdRepo) Update(ctx context.Context, template *model.SpeechDelayTemplate, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":     "sdRepo.Update",
		"template": helper.Dump(template),
	})

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Save(template).Error
	if err != nil {
		logger.WithError(err).Error("failed to update speech delay template")
		return err
	}

	return nil
}
