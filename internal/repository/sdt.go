package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// uuidList a helper type to Scan database value from ARRAY_AGG from database.
// If only needed here, no need to move it to other package
type uuidList []uuid.UUID

// Scan implements the sql.Scanner interface.
func (ul *uuidList) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		uuidStrings := strings.Split(strings.Trim(src, "{}"), ",")
		for _, u := range uuidStrings {
			parsedUUID, err := uuid.Parse(u)
			if err != nil {
				return err
			}
			*ul = append(*ul, parsedUUID)
		}

		return nil
	default:
		return errors.New("unsupported type to scan uuidList")
	}
}

// intList a helper type to Scan database value from ARRAY_AGG from database.
// If only needed here, no need to move it to other package
type intList []int

// Scan implements the sql.Scanner interface.
func (il *intList) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		intString := strings.Split(strings.Trim(src, "{}"), ",")
		for _, u := range intString {
			val, err := strconv.ParseInt(u, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value for int: %s. err: %s", u, err.Error())
			}

			*il = append(*il, int(val))
		}

		return nil
	default:
		return errors.New("unsupported type to scan intList")
	}
}

// stringList a helper type to Scan database value from ARRAY_AGG from database.
// If only needed here, no need to move it to other package
type stringList []string

// Scan implements the sql.Scanner interface.
func (sl *stringList) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		strString := strings.Split(strings.Trim(src, "{}"), ",")
		for _, u := range strString {
			s := strings.TrimSuffix(u, "\"")
			s = strings.TrimPrefix(s, "\"")
			*sl = append(*sl, s)
		}

		return nil
	default:
		return errors.New("unsupported type to scan stringList")
	}
}

// timeList a helper type to Scan database value from ARRAY_AGG from database.
// If only needed here, no need to move it to other package
type timeList []time.Time

// Scan implements the sql.Scanner interface.
func (tl *timeList) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		timeString := strings.Split(strings.Trim(src, "{}"), ",")
		for _, u := range timeString {
			s := strings.Trim(u, "\"")
			t, err := time.Parse("2006-01-02 15:04:05.999999-07", s)
			if err != nil {
				return fmt.Errorf("invalid value for time: %s. err: %s", u, err.Error())
			}

			*tl = append(*tl, t)
		}

		return nil
	default:
		return errors.New("unsupported type to scan timeList")
	}
}

type rawTemplateStatistic struct {
	TemplateID             uuid.UUID  `gorm:"column:template_id"`
	TemplateName           string     `gorm:"column:template_name"`
	IndicationThreshold    int        `gorm:"column:indication_threshold"`
	PositiveIndiationText  string     `gorm:"column:positive_indication_text"`
	NegativeIndicationText string     `gorm:"column:negative_indication_text"`
	TestResultID           uuidList   `gorm:"column:test_id"`
	PackageID              uuidList   `gorm:"column:package_id"`
	TotalPoint             intList    `gorm:"column:total_point"`
	PackageName            stringList `gorm:"column:package_name"`
	TestFinishedAt         timeList   `gorm:"column:finished_at"`
}

func (r *sdtrRepo) Statistic(ctx context.Context, userID uuid.UUID) ([]model.SDTestStatistic, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "sdtrRepo.Statistic",
		"id":   userID.String(),
	})

	var res []rawTemplateStatistic
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
			tt.id AS template_id,
			tt."name"  AS template_name,
			tt."template" -> 'indicationThreshold' AS indication_threshold,
			tt."template" -> 'negativeIndicationText' AS negative_indication_text,
			tt."template" -> 'positiveIndicationText' AS positive_indication_text,
			ARRAY_AGG(tr.id) AS test_id,
			ARRAY_AGG(tr.package_id) AS package_id,
			ARRAY_AGG(tr."result"-> 'total') AS total_point,
			ARRAY_AGG(tp."name") AS package_name,
			ARRAY_AGG(tr.finished_at ORDER BY tr.finished_at ASC)  AS finished_at
				FROM test_results tr
					JOIN test_packages tp ON tr.package_id = tp.id
					JOIN test_templates tt ON tp.template_id = tt.id 
						WHERE tr.user_id  = ?
						AND tr.finished_at IS NOT NULL
						GROUP BY tt.id;
		`, userID).Scan(&res).Error

	if err != nil {
		logger.WithError(err).Error("failed to get test result statistic")
		return nil, err
	}

	if len(res) == 0 {
		return nil, ErrNotFound
	}

	stats := []model.SDTestStatistic{}

	for _, v := range res {
		stats = append(stats, toSDTestStatistic(v))
	}

	return stats, nil
}

func toSDTestStatistic(res rawTemplateStatistic) model.SDTestStatistic {
	s := model.SDTestStatistic{}
	s.TemplateID = res.TemplateID
	s.TemplateName = res.TemplateName
	s.IndicationThreshold = res.IndicationThreshold
	s.NegativeIndicationText = strings.TrimSuffix(strings.TrimPrefix(res.NegativeIndicationText, "\""), "\"")
	s.PositiveIndiationText = strings.TrimSuffix(strings.TrimPrefix(res.PositiveIndiationText, "\""), "\"")

	stats := []model.StatsComponent{}
	for i := 0; i < len(res.TotalPoint); i++ {
		stats = append(stats, model.StatsComponent{
			TestResultID:   res.TestResultID[i],
			PackageID:      res.PackageID[i],
			ResultPoint:    res.TotalPoint[i],
			PackageName:    res.PackageName[i],
			TestFinishedAt: res.TestFinishedAt[i],
		})
	}

	s.Stats = stats

	return s
}
