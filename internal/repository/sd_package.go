package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sdpRepo struct {
	db *gorm.DB
}

// NewSDPackageRepository will create new an sdpRepo object representation of model.SDPackageRepository interface
func NewSDPackageRepository(db *gorm.DB) model.SDPackageRepository {
	return &sdpRepo{
		db: db,
	}
}

func (r *sdpRepo) Create(ctx context.Context, input *model.SpeechDelayPackage) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Create",
		"input": helper.Dump(input),
	})
	if err := r.db.WithContext(ctx).Create(input).Error; err != nil {
		logger.WithError(err).Error("failed to create sd package")
		return err
	}

	return nil
}

func (r *sdpRepo) FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdpRepo.FindByID",
		"id":   id.String(),
	})

	query := r.db.WithContext(ctx)
	if includeDeleted {
		query = query.Unscoped()
	}

	sdp := &model.SpeechDelayPackage{}
	err := query.Take(sdp, "id = ?", id).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd package")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return sdp, nil
	}
}

func (r *sdpRepo) Search(ctx context.Context, input *model.SearchSDPackageInput) ([]*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Search",
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

	var packs []*model.SpeechDelayPackage
	err := query.Limit(input.Limit).Offset(input.Offset).Order("created_at desc").Find(&packs).Error
	if err != nil {
		logger.WithError(err).Error("failed to search sd package from db")
		return []*model.SpeechDelayPackage{}, err
	}

	return packs, nil
}

func (r *sdpRepo) Update(ctx context.Context, pack *model.SpeechDelayPackage, tx *gorm.DB) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":    "sdpRepo.Update",
		"package": helper.Dump(pack),
	})

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Save(pack).Error
	if err != nil {
		logger.WithError(err).Error("failed to update speech delay package")
		return err
	}

	return nil
}

func (r *sdpRepo) Delete(ctx context.Context, id uuid.UUID) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Delete",
		"input": helper.Dump(id),
	})

	deleted := &model.SpeechDelayPackage{}
	err := r.db.WithContext(ctx).Clauses(clause.Returning{}).Delete(deleted, "id = ?", id).Error
	if err != nil {
		logger.WithError(err).Error("failed to delete speech delay package")
		return nil, err
	}

	return deleted, nil
}

func (r *sdpRepo) UndoDelete(ctx context.Context, id uuid.UUID) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.Delete",
		"input": helper.Dump(id),
	})

	pack := &model.SpeechDelayPackage{}
	err := r.db.WithContext(ctx).Model(pack).Unscoped().Clauses(clause.Returning{}).
		Where("id = ?", id).Update("deleted_at", gorm.DeletedAt{Valid: false}).Error
	if err != nil {
		logger.WithError(err).Error("failed to delete speech delay package")
		return nil, err
	}

	return pack, nil
}

func (r *sdpRepo) FindRandomActivePackage(ctx context.Context) (*model.SpeechDelayPackage, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdpRepo.FindRandomActivePackage",
	})

	pack := &model.SpeechDelayPackage{}
	err := r.db.WithContext(ctx).
		Order("RANDOM()").
		Limit(1).
		Take(pack, "is_active = ?", true).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find random active package")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return pack, nil
	}
}

type testPackageCount struct {
	PackageID uuid.UUID `gorm:"column:package_id"`
	Count     int       `gorm:"column:count"`
}

func (r *sdpRepo) FindLeastUsedPackageIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.FindLeastUsedPackageIDByUserID",
		"input": helper.Dump(userID),
	})

	var activePackageIDs []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.SpeechDelayPackage{}).
		Select("id").
		Where("is_active = ?", true).
		Scan(&activePackageIDs).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find active sd package")
		return uuid.Nil, err
	case gorm.ErrRecordNotFound:
		return uuid.Nil, ErrNotFound
	case nil:
		if len(activePackageIDs) == 0 {
			return uuid.Nil, ErrNotFound
		}
	}

	var tpc []testPackageCount
	err = r.db.WithContext(ctx).
		Model(&model.SDTest{}).
		Where("user_id = ?", userID).
		Select("package_id, count(package_id) as count").
		Group("package_id").
		Scan(&tpc).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find sd test result")
		return uuid.Nil, err
	case gorm.ErrRecordNotFound:
		return uuid.Nil, ErrNotFound
	case nil:
		break
	}

	// here we will find a package id that doesn't exists in
	// the ids listed in tpc
	// if not found, it means the user never used that, so return the id immediatly
	for _, v := range activePackageIDs {
		found := false
		for _, x := range tpc {
			if v == x.PackageID {
				found = true
			}
		}

		if !found {
			return v, nil
		}
	}

	return findTheLeastUsedPackage(tpc)
}

func (r *sdpRepo) GetTemplateByPackageID(ctx context.Context, packageID uuid.UUID) (*model.SpeechDelayTemplate, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "sdpRepo.GetTemplateByPackageID",
		"input": helper.Dump(packageID),
	})

	pack := &model.SpeechDelayPackage{}
	tem := &model.SpeechDelayTemplate{}
	err := r.db.WithContext(ctx).
		Where("id = (?)", r.db.Table(pack.TableName()).Select("template_id").Where("id = (?)", packageID)).
		Take(&tem).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find template by sd package id")
		return nil, err
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	case nil:
		return tem, nil
	}
}

// will be used to find the least used package
// originally made to reduce the code complexity in order to pass the linter
func findTheLeastUsedPackage(tpc []testPackageCount) (uuid.UUID, error) {
	// here we will find the least used package by user
	arr := []int{}
	for _, v := range tpc {
		arr = append(arr, v.Count)
	}

	min := slices.Min(arr)
	for _, v := range tpc {
		if v.Count == min {
			return v.PackageID, nil
		}
	}

	return uuid.Nil, errors.New("unexpected error when trying to find least used package")
}
