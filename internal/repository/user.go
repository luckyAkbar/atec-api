package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepo struct {
	db     *gorm.DB
	cacher model.Cacher
}

// NewUserRepository create a new user repository. Satisfy model.UserRepository interface
func NewUserRepository(db *gorm.DB, cacher model.Cacher) model.UserRepository {
	return &userRepo{db, cacher}
}

func (r *userRepo) Create(ctx context.Context, user *model.User, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "userRepo.Create",
	})

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Create(user).Error
	if err != nil {
		logger.WithError(err).Error("failed to create user")
		return err
	}

	return nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "userRepo.FindByEmail",
	})

	user := &model.User{}
	err := r.db.WithContext(ctx).Take(user, "email = ?", email).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by email")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return user, nil
	}
}

func (r *userRepo) UpdateActiveStatus(ctx context.Context, id uuid.UUID, status bool) (*model.User, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "userRepo.UpdateActiveStatus",
	})

	user := &model.User{}
	err := r.db.WithContext(ctx).
		Model(user).Clauses(clause.Returning{}).
		Where("id = ? AND deleted_at IS NULL", id).Update("is_active", status).Error

	switch err {
	default:
		logger.WithError(err).Error("failed to update user active status")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return user, nil
	}
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "userRepo.FindByID",
	})

	user := &model.User{}
	err := r.db.WithContext(ctx).Take(user, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by id")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return user, nil
	}
}

func (r *userRepo) CreateChangePasswordSession(ctx context.Context, key string, expiry time.Duration, session *model.ChangePasswordSession) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":    "userRepo.CreateChangePasswordSession",
		"key":     key,
		"session": helper.Dump(session),
	})

	err := r.cacher.Set(ctx, key, session.ToJSONString(), expiry)
	if err != nil {
		logger.WithError(err).Error("failed to write change password session to cache")
		return err
	}

	return nil
}
