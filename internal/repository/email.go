// Package repository holds all data related operation
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type emailRepo struct {
	db *gorm.DB
}

// NewEmailRepository satisfy model.EmailRepository
func NewEmailRepository(db *gorm.DB) model.EmailRepository {
	return &emailRepo{db: db}
}

func (r *emailRepo) Create(ctx context.Context, email *model.Email) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "emailRepo.Create",
		"data": helper.Dump(email),
	})

	if err := r.db.WithContext(ctx).Create(email).Error; err != nil {
		logger.WithError(err).Error("failed to write emails data to db")
		return err
	}

	return nil
}

func (r *emailRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Email, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "emailRepo.FindByID",
		"data": helper.Dump(id),
	})

	email := &model.Email{}
	err := r.db.WithContext(ctx).Take(email, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to read email data from db")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return email, nil
	}
}

func (r *emailRepo) Update(ctx context.Context, email *model.Email) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "emailRepo.Update",
		"data": helper.Dump(email),
	})

	if err := r.db.WithContext(ctx).Save(email).Error; err != nil {
		logger.WithError(err).Error("failed to update email data to db")
		return err
	}

	return nil
}
