package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// SDAnswerAndValue sd answer and value
type SDAnswerAndValue struct {
	Text  string `json:"text" validate:"required"`
	Value int    `json:"value" validate:"required,min=1"`
}

// SDQuestionAndAnswers sd question and answer
type SDQuestionAndAnswers struct {
	Question        string             `json:"question" validate:"required"`
	AnswersAndValue []SDAnswerAndValue `json:"answerAndValue" validate:"required,min=1,unique=Value,dive"`
}

// SDSubGroupDetail sd sub group detail
type SDSubGroupDetail struct {
	Name                   string                 `json:"name" validate:"required"`
	QuestionAndAnswerLists []SDQuestionAndAnswers `json:"questionAndAnswerLists" validate:"required,min=1,dive"`
}

// SDPackage sd package
type SDPackage struct {
	PackageName     string             `json:"packageName" validate:"required"`
	TemplateID      uuid.UUID          `json:"templateID" validate:"required"`
	SubGroupDetails []SDSubGroupDetail `json:"subGroupDetails" validate:"required,min=1,unique=Name,dive"`
}

// PartialValidation will validate the SD Package. enough to be used for first time creating / just updating the SD Template
func (sdp *SDPackage) PartialValidation() error {
	return validator.Struct(sdp)
}

// FullValidation will ensure that the SDPackage satisfy all the rules set by supplied SpeechDelayTemplate
func (sdp *SDPackage) FullValidation(t *SpeechDelayTemplate) error {
	if err := sdp.PartialValidation(); err != nil {
		return err
	}

	// safety check
	if err := t.Template.FullValidation(); err != nil {
		return err
	}

	if !t.IsActive || t.DeletedAt.Valid {
		return errors.New("template is not active or already deleted")
	}

	if err := sdp.ensureSubGroupPackageExistsOnTemplate(t); err != nil {
		return err
	}

	if err := sdp.ensureSubGroupTemplateExistsOnPackage(t); err != nil {
		return err
	}

	return sdp.ensureAllQuestionsAndAnwerMatchToTemplate(t)
}

// this process ensure all the sub group on package also exists on template
// prevent an unregistered sub group on package
func (sdp *SDPackage) ensureSubGroupPackageExistsOnTemplate(t *SpeechDelayTemplate) error {
	matchCount := 0
	for _, s := range sdp.SubGroupDetails {
		for _, q := range t.Template.SubGroupDetails {
			if s.Name == q.Name {
				matchCount++
			}
		}
	}

	if matchCount != len(sdp.SubGroupDetails) {
		return errors.New("at least one sub group package details exists, but not present on the template")
	}

	return nil
}

// this ensure that all sub group details on template also present on the package
func (sdp *SDPackage) ensureSubGroupTemplateExistsOnPackage(t *SpeechDelayTemplate) error {
	matchCount := 0
	for _, q := range t.Template.SubGroupDetails {
		for _, s := range sdp.SubGroupDetails {
			if s.Name == q.Name {
				matchCount++
			}
		}
	}

	if matchCount != len(t.Template.SubGroupDetails) {
		return errors.New("at least one sub group template details is not present on the package sub group details")
	}

	return nil
}

// matcher only be used to easily pair the sub group from package with sub group from template
type matcher struct {
	template SDTemplateSubGroupDetail
	pack     SDSubGroupDetail
}

func (sdp *SDPackage) ensureAllQuestionsAndAnwerMatchToTemplate(t *SpeechDelayTemplate) error {
	// trying to map the matching template sub group to actual package sub group
	m := []matcher{}
	for _, tt := range t.Template.SubGroupDetails {
		for _, tq := range sdp.SubGroupDetails {
			if tt.Name == tq.Name {
				m = append(m, matcher{
					template: tt,
					pack:     tq,
				})
			}
		}
	}

	return sdp.validateQuestionCountAndAnswerCount(m)
}

func (sdp *SDPackage) validateQuestionCountAndAnswerCount(m []matcher) error {
	for _, d := range m {
		// ensure that the number of question match the number defined in template
		if len(d.pack.QuestionAndAnswerLists) != d.template.QuestionCount {
			return fmt.Errorf("the number of questions on the package is not match with the template, group name: %s expecting %d got %d", d.template.Name, d.template.QuestionCount, len(d.pack.QuestionAndAnswerLists))
		}

		// ensure that every question's answer match the number defined in template
		for _, q := range d.pack.QuestionAndAnswerLists {
			if len(q.AnswersAndValue) != d.template.AnswerOptionCount {
				return fmt.Errorf("the number of answers on the package is not match with the template, group name: %s expecting %d got %d", d.template.Name, d.template.AnswerOptionCount, len(q.AnswersAndValue))
			}
		}

		if err := sdp.ensureAllValuesAreOrdered(d); err != nil {
			return err
		}
	}

	return nil
}

// this func ensure that the value list are ordered, from 1 to the number of specified question count
// e.g ensure 1, 2, 3, 4 are the Value when AnswerOptionCount are 4, not 1, 2, 3, 5, 6
func (sdp *SDPackage) ensureAllValuesAreOrdered(d matcher) error {
	for _, a := range d.pack.QuestionAndAnswerLists {
		for _, b := range a.AnswersAndValue {
			found := false
			missing := 0
			for i := 1; i <= d.template.AnswerOptionCount; i++ {
				if b.Value == i {
					found = true
					break
				}
				missing = i
			}

			if !found {
				return fmt.Errorf("value list on group %s was not complete. Missing value %d and expecting range %d - %d", d.template.Name, missing, 1, d.template.AnswerOptionCount)
			}
		}
	}

	return nil
}

// Scan is a function to scan database value to CreateSDTemplateInput
func (sdp *SDPackage) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue interface{}) (err error) {
	if dbValue == nil {
		return
	}

	var bytes []byte
	switch v := dbValue.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: %#v", dbValue)
	}

	if err = json.Unmarshal(bytes, sdp); err != nil {
		return
	}

	return
}

// Value is a function to convert CreateSDTemplateInput to json
func (sdp SDPackage) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	return json.Marshal(fieldValue)
}

// GeneratedSDPackage will be used to define the generated SD package as the returned value as REST API responses
type GeneratedSDPackage struct {
	ID         uuid.UUID      `json:"id"`
	TemplateID uuid.UUID      `json:"templateID"`
	Name       string         `json:"name"`
	CreatedBy  uuid.UUID      `json:"createdBy"`
	Package    *SDPackage     `json:"package"`
	IsActive   bool           `json:"isActive"`
	IsLocked   bool           `json:"isLocked"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
}

// SpeechDelayPackage will represent speech delay only test packages on db table.
// Because the actual table's column datatype for package is JSONB, every ATEC test type must
// define a specific struct and make sure the Package is support JSONB implementation.
// Also make sure to customize function TableName to return "test_packages".
type SpeechDelayPackage struct {
	ID         uuid.UUID
	TemplateID uuid.UUID
	Name       string
	CreatedBy  uuid.UUID
	Package    *SDPackage
	IsActive   bool
	IsLocked   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt
}

// TableName define the table name for gorm
func (sdp SpeechDelayPackage) TableName() string {
	return "test_packages"
}

// ToRESTResponse convert SpeechDelayPackage to GeneratedSDPackage which ease rest response generation
func (sdp *SpeechDelayPackage) ToRESTResponse() *GeneratedSDPackage {
	return &GeneratedSDPackage{
		ID:         sdp.ID,
		TemplateID: sdp.TemplateID,
		Name:       sdp.Name,
		CreatedBy:  sdp.CreatedBy,
		Package:    sdp.Package,
		IsActive:   sdp.IsActive,
		IsLocked:   sdp.IsLocked,
		CreatedAt:  sdp.CreatedAt,
		UpdatedAt:  sdp.UpdatedAt,
		DeletedAt:  sdp.DeletedAt,
	}
}

// SearchSDPackageInput input to search sd package
type SearchSDPackageInput struct {
	TemplateID     uuid.UUID `query:"templateID"`
	CreatedBy      uuid.UUID `query:"createdBy"`
	CreatedAfter   time.Time `query:"createdAfter"`
	IsActive       *bool     `query:"isActive"`
	IsLocked       *bool     `query:"isLocked"`
	IncludeDeleted bool      `query:"includeDeleted"`
	Limit          int       `query:"limit"`
	Offset         int       `query:"offset"`
}

// ToWhereQuery convert SearchSDPackageInput to where query and conditions. If limit is unset / set over 100, will be set to 100.
// If offset is unset / set under 0, will be set to 0.
func (sdpi *SearchSDPackageInput) ToWhereQuery() ([]interface{}, []interface{}) {
	var whereQuery []interface{}
	var conds []interface{}

	if sdpi.Limit < 0 || sdpi.Limit > 100 {
		sdpi.Limit = 100
	}

	if sdpi.Offset < 0 {
		sdpi.Offset = 0
	}

	if sdpi.TemplateID != uuid.Nil {
		whereQuery = append(whereQuery, "template_id = ?")
		conds = append(conds, sdpi.TemplateID)
	}

	if sdpi.CreatedBy != uuid.Nil {
		whereQuery = append(whereQuery, "created_by = ?")
		conds = append(conds, sdpi.CreatedBy)
	}

	if !reflect.ValueOf(sdpi.CreatedAfter).IsZero() {
		whereQuery = append(whereQuery, "created_at > ?")
		conds = append(conds, sdpi.CreatedAfter.UTC())
	}

	if sdpi.IsActive != nil {
		whereQuery = append(whereQuery, "is_active = ?")
		conds = append(conds, *sdpi.IsActive)
	}

	if sdpi.IsLocked != nil {
		whereQuery = append(whereQuery, "is_locked = ?")
		conds = append(conds, *sdpi.IsLocked)
	}

	return whereQuery, conds
}

// SearchPackageOutput output search sd package
type SearchPackageOutput struct {
	Packages []*GeneratedSDPackage `json:"packages"`
	Count    int                   `json:"count"`
}

// SDPackageUsecase interface for SD package usecase
type SDPackageUsecase interface {
	Create(ctx context.Context, input *SDPackage) (*GeneratedSDPackage, *common.Error)
	FindByID(ctx context.Context, id uuid.UUID) (*GeneratedSDPackage, *common.Error)
	Search(ctx context.Context, input *SearchSDPackageInput) (*SearchPackageOutput, *common.Error)
	Update(ctx context.Context, id uuid.UUID, input *SDPackage) (*GeneratedSDPackage, *common.Error)
	Delete(ctx context.Context, id uuid.UUID) (*GeneratedSDPackage, *common.Error)
	UndoDelete(ctx context.Context, id uuid.UUID) (*GeneratedSDPackage, *common.Error)
	ChangeSDPackageActiveStatus(ctx context.Context, id uuid.UUID, isActive bool) (*GeneratedSDPackage, *common.Error)
}

// SDPackageRepository interface for SD package repository
type SDPackageRepository interface {
	Create(ctx context.Context, input *SpeechDelayPackage) error
	FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*SpeechDelayPackage, error)
	Search(ctx context.Context, input *SearchSDPackageInput) ([]*SpeechDelayPackage, error)
	Update(ctx context.Context, pack *SpeechDelayPackage, tx *gorm.DB) error
	Delete(ctx context.Context, id uuid.UUID) (*SpeechDelayPackage, error)
	UndoDelete(ctx context.Context, id uuid.UUID) (*SpeechDelayPackage, error)
}
