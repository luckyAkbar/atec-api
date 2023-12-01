package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository create a new user repository. Satisfy model.UserRepository interface
func NewUserRepository(db *gorm.DB) model.UserRepository {
	return &userRepo{db}
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
