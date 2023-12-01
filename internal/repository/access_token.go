package repository

import (
	"context"

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
