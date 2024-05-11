package repository

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type accessTokenRepo struct {
	db     *gorm.DB
	cacher model.Cacher
}

// NewAccessTokenRepository returns a new AccessTokenRepository
func NewAccessTokenRepository(db *gorm.DB, cacher model.Cacher) model.AccessTokenRepository {
	return &accessTokenRepo{
		db:     db,
		cacher: cacher,
	}
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

func (r *accessTokenRepo) DeleteByIDs(ctx context.Context, ids []uuid.UUID, hardDelete bool) error {
	query := r.db.WithContext(ctx)
	if hardDelete {
		query = query.Unscoped()
	}

	err := query.Delete(&model.AccessToken{}, ids).Error
	if err != nil {
		return err
	}

	return nil
}

type credential struct {
	model.AccessToken
	model.User
}

func (r *accessTokenRepo) FindCredentialByToken(ctx context.Context, token string) (*model.AccessToken, *model.User, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.FindCredentialByToken",
		"data": token,
	})

	var result credential

	result, err := r.getCreadentialsTokenFromCache(ctx, token)
	switch err {
	default:
		break
	case ErrNotFound:
		return nil, nil, ErrNotFound
	case nil:
		return &result.AccessToken, &result.User, nil
	}

	err = r.db.WithContext(ctx).Model(&model.AccessToken{}).
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
		_ = r.setNilCredentialsToCache(ctx, token)
		return nil, nil, ErrNotFound
	}

	_ = r.setCredentialsToCache(ctx, result)

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

func (r *accessTokenRepo) getCreadentialsTokenFromCache(ctx context.Context, token string) (credential, error) {
	cache, err := r.cacher.Get(ctx, token)
	switch err {
	default:
		return credential{}, err
	case redis.Nil:
		return credential{}, redis.Nil
	case nil:
		break
	}

	// when this happend, the caller should immedieatly know that the value
	// is purposefully set to be nil, thus prevent refetching from db for
	// value that never exists
	if cache == model.NilKey {
		return credential{}, ErrNotFound
	}

	res := credential{}
	if err := json.Unmarshal([]byte(cache), &res); err != nil {
		return credential{}, err
	}

	if reflect.ValueOf(res.AccessToken).IsZero() || reflect.ValueOf(res.User).IsZero() {
		return credential{}, ErrNotFound
	}

	return res, nil
}

func (r *accessTokenRepo) setCredentialsToCache(ctx context.Context, creds credential) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "accessTokenRepo.setCredentialsToCache",
	})

	logger.Info("start setting creds to cache")

	val, err := json.Marshal(creds)
	if err != nil {
		logger.WithError(err).Error("failed to marshal creds for caching")
		return err
	}

	now := time.Now().UTC()
	if err := r.cacher.Set(ctx, creds.Token, string(val), creds.ValidUntil.Sub(now)); err != nil {
		logger.WithError(err).Error("failed to set cache for creds")
		return err
	}

	return nil
}

func (r *accessTokenRepo) setNilCredentialsToCache(ctx context.Context, key string) error {
	if err := r.cacher.Set(ctx, key, model.NilKey, config.AccessTokenActiveDuration()); err != nil {
		return err
	}

	return nil
}

func (r *accessTokenRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.AccessToken, error) {
	if limit <= 0 {
		limit = 1
	}

	accessTokens := []model.AccessToken{}
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Find(&accessTokens).Error
	switch err {
	default:
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		if len(accessTokens) == 0 {
			return nil, ErrNotFound
		}

		return accessTokens, nil
	}
}

func (r *accessTokenRepo) DeleteCredentialsFromCache(ctx context.Context, tokens []string) error {
	if err := r.cacher.Del(ctx, tokens); err != nil {
		return err
	}

	return nil
}
