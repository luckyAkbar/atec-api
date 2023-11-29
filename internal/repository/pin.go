package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type pinRepo struct {
	db *gorm.DB
}

// NewPinRepository satisfy model.PinRepository
func NewPinRepository(db *gorm.DB) model.PinRepository {
	return &pinRepo{db: db}
}

func (r *pinRepo) Create(ctx context.Context, pin *model.Pin, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "pinRepo.Create",
		"data": helper.Dump(pin),
	})

	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(pin).Error; err != nil {
		logger.WithError(err).Error("failed to write pins data to db")
		return err
	}

	return nil
}

func (r *pinRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Pin, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "pinRepo.FindByID",
		"data": helper.Dump(id),
	})

	pin := &model.Pin{}
	err := r.db.WithContext(ctx).Take(pin, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to read pin data from db")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return pin, nil
	}
}
