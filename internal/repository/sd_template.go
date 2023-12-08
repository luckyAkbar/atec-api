package repository

import (
	"context"

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
