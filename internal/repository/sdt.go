package repository

import (
	"context"

	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type sdtrRepo struct {
	db *gorm.DB
}

// NewSDTestResultRepository create new SDTestRepository
func NewSDTestResultRepository(db *gorm.DB) model.SDTestRepository {
	return &sdtrRepo{db}
}

func (r *sdtrRepo) Create(ctx context.Context, test *model.SDTest, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtrRepo.Create",
		"input": helper.Dump(test),
	})

	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(test).Error; err != nil {
		logger.WithError(err).Error("failed to create test result")
		return err
	}

	return nil
}
