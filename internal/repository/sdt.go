package repository

import (
	"context"

	"github.com/google/uuid"
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

func (r *sdtrRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SDTest, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdtrRepo.FindByID",
		"id":   id.String(),
	})

	sdt := &model.SDTest{}
	err := r.db.WithContext(ctx).Take(sdt, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd test result")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return sdt, nil
	}
}

func (r *sdtrRepo) Update(ctx context.Context, tr *model.SDTest, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":       "sdtrRepo.Update",
		"testResult": helper.Dump(tr),
	})

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Save(tr).Error
	if err != nil {
		logger.WithError(err).Error("failed to update test result")
		return err
	}

	return nil
}

func (r *sdtrRepo) Search(ctx context.Context, input *model.ViewHistoriesInput) ([]*model.SDTest, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdtrRepo.Search",
		"input": helper.Dump(input),
	})

	query := r.db.WithContext(ctx)
	if input.IncludeDeleted {
		query = query.Unscoped()
	}

	where, conds := input.ToWhereQuery()
	for i := 0; i < len(where); i++ {
		query = query.Where(where[i], conds[i])
	}

	if !input.IncludeUnfinished {
		query = query.Where("finished_at IS NOT NULL")
	}

	var sdt []*model.SDTest
	err := query.Limit(input.Limit).Offset(input.Offset).Order("created_at DESC").Find(&sdt).Error
	if err != nil {
		logger.WithError(err).Error("failed to search test result")
		return nil, err
	}

	return sdt, nil
}
