package repository

import (
	"context"
	"reflect"

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

func (r *accessTokenRepo) FindCredentialByToken(ctx context.Context, token string) (*model.AccessToken, *model.User, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.FindCredentialByToken",
		"data": token,
	})

	type res struct {
		model.AccessToken
		model.User
	}

	var result res
	err := r.db.WithContext(ctx).Model(&model.AccessToken{}).
		Select(`
			"access_tokens"."id",
			"access_tokens"."token",
			"access_tokens"."user_id",
			"access_tokens"."valid_until",
			"access_tokens"."created_at",
			"access_tokens"."updated_at",
			"access_tokens"."deleted_at",
			"users"."id",
			"users"."email",
			"users"."password",
			"users"."username",
			"users"."is_active",
			"users"."role"
		`).Joins(`FULL JOIN "users" ON "access_tokens"."user_id" = "users"."id"`).
		Where(`"access_tokens"."token" = ?`, token).Scan(&result).Error

	if err != nil {
		logger.WithError(err).Error("failed to read credentials data from db")
		return nil, nil, err
	}

	if reflect.ValueOf(result.AccessToken).IsZero() || reflect.ValueOf(result.User).IsZero() {
		return nil, nil, ErrNotFound
	}

	return &result.AccessToken, &result.User, nil
}

func (r *accessTokenRepo) DeleteByUserID(ctx context.Context, id uuid.UUID, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.DeleteByUserID",
		"data": helper.Dump(id),
	})

	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Where("user_id = ?", id).Delete(&model.AccessToken{}).Error; err != nil {
		logger.WithError(err).Error("failed to delete access token data from db")
		return err
	}

	return nil
}
