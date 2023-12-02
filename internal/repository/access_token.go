package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type accessTokenRepo struct {
	db *gorm.DB
}

// NewAccessTokenRepository returns a new AccessTokenRepository
func NewAccessTokenRepository(db *gorm.DB) model.AccessTokenRepository {
	return &accessTokenRepo{db: db}
}

func (r *accessTokenRepo) Create(ctx context.Context, at *model.AccessToken) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.Create",
		"data": helper.Dump(at),
	})

	if err := r.db.WithContext(ctx).Create(at).Error; err != nil {
		logger.WithError(err).Error("failed to write access token data to db")
		return err
	}

	return nil
}

func (r *accessTokenRepo) FindByToken(ctx context.Context, token string) (*model.AccessToken, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.FindByToken",
		"data": token,
	})

	at := &model.AccessToken{}
	err := r.db.WithContext(ctx).Take(at, "token = ?", token).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to read access token data from db")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return at, nil
	}
}

func (r *accessTokenRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.FindByToken",
		"data": helper.Dump(id),
	})

	if err := r.db.Unscoped().WithContext(ctx).Delete(&model.AccessToken{}, id).Error; err != nil {
		logger.WithError(err).Error("failed to delete access token data from db")
		return err
	}

	return nil
}
