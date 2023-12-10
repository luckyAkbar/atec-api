package model

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// SDTemplateSubGroupDetail define what the details of every sub group used by this template
type SDTemplateSubGroupDetail struct {
	Name              string `json:"name" validate:"required"`
	QuestionCount     int    `json:"questionCount" validate:"required,min=1"`
	AnswerOptionCount int    `json:"answerOptionCount" validate:"required,min=2"`
}

// SDTemplate define what the full SD test template will look like
type SDTemplate struct {
	Name                   string                     `json:"name" validate:"required,max=255"`
	IndicationThreshold    int                        `json:"indicationThreshold" validate:"required,min=0"`
	PositiveIndiationText  string                     `json:"positiveIndicationText" validate:"required"`
	NegativeIndicationText string                     `json:"negativeIndicationText" validate:"required"`
	SubGroupDetails        []SDTemplateSubGroupDetail `json:"subGroupDetails" validate:"min=1,dive"`
}

// PartialValidation will validate the SD Template. enough to be used for first time creating / just updating the SD Template
func (csdti *SDTemplate) PartialValidation() error {
	return validator.Struct(csdti)
}

// CountMaximumPoint will count the maximum point possible that can be achieved by this SD Template.
func (csdti *SDTemplate) CountMaximumPoint() int {
	var maximumPoint int
	for _, subGroupDetail := range csdti.SubGroupDetails {
		maximumPoint += subGroupDetail.QuestionCount * subGroupDetail.AnswerOptionCount
	}

	return maximumPoint
}

// CountMinimumPoint will count the minimum point possible that can be achieved by this SD Template.
// Will always equal to the length of SubGroupDetails
func (csdti *SDTemplate) CountMinimumPoint() int {
	return len(csdti.SubGroupDetails)
}

// FullValidation will validate the SD Template to ensure all rules are satisfied. Suitable to be used to activate the SD Template
func (csdti *SDTemplate) FullValidation() error {
	if err := csdti.PartialValidation(); err != nil {
		return err
	}

	if csdti.IndicationThreshold < csdti.CountMinimumPoint() {
		return fmt.Errorf("indicationThreshold must be greater than or equal to the number of sub group details (min: %d)", csdti.CountMinimumPoint())
	}

	if csdti.IndicationThreshold > csdti.CountMaximumPoint() {
		return fmt.Errorf("indicationThreshold must be less than or equal to the maximum point (max: %d)", csdti.CountMaximumPoint())
	}

	return nil
}

// Scan is a function to scan database value to CreateSDTemplateInput
func (csdti *SDTemplate) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue interface{}) (err error) {
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

	if err = json.Unmarshal(bytes, csdti); err != nil {
		return
	}

	return
}

// Value is a function to convert CreateSDTemplateInput to json
func (csdti SDTemplate) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	return json.Marshal(fieldValue)
}

// GeneratedSDTemplate will be used to define the generated SD template as the returned value as REST API responses
type GeneratedSDTemplate struct {
	ID        uuid.UUID      `json:"id"`
	CreatedBy uuid.UUID      `json:"createdBy"`
	Name      string         `json:"name"`
	Template  *SDTemplate    `json:"template"`
	IsActive  bool           `json:"isActive"`
	IsLocked  bool           `json:"isLocked"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty"`
}

// SpeechDelayTemplate will represent speech delay only test templates on db table.
// Because the actual table's column datatype for tempalte is JSONB, every ATEC test type must
// define a specific struct and make sure the Template is support JSONB implementation.
// Also make sure to customize function TableName to return "test_templates".
type SpeechDelayTemplate struct {
	ID        uuid.UUID
	CreatedBy uuid.UUID
	Name      string
	IsActive  bool
	IsLocked  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Template  *SDTemplate
}

// TableName define the table name for gorm
func (sdt SpeechDelayTemplate) TableName() string {
	return "test_templates"
}

// ToRESTResponse convert SpeechDelayTemplate to GeneratedSDTemplate which ease rest response generation
func (sdt *SpeechDelayTemplate) ToRESTResponse() *GeneratedSDTemplate {
	return &GeneratedSDTemplate{
		ID:        sdt.ID,
		CreatedBy: sdt.CreatedBy,
		Name:      sdt.Name,
		Template:  sdt.Template,
		IsActive:  sdt.IsActive,
		IsLocked:  sdt.IsLocked,
		CreatedAt: sdt.CreatedAt,
		UpdatedAt: sdt.UpdatedAt,
		DeletedAt: sdt.DeletedAt,
	}
}

// SearchSDTemplateInput input for searching SDTemplate input
type SearchSDTemplateInput struct {
	CreatedBy      uuid.UUID `query:"createdBy"`
	CreatedAfter   time.Time `query:"createdAfter"`
	IsActive       *bool     `query:"isActive"`
	IsLocked       *bool     `query:"isLocked"`
	IncludeDeleted bool      `query:"includeDeleted"`
	Limit          int       `query:"limit"`
	Offset         int       `query:"offset"`
}

// ToWhereQuery convert SearchSDTemplateInput to where query and conditions. If limit is unset / set over 100, will be set to 100.
// If offset is unset / set under 0, will be set to 0.
func (sdti *SearchSDTemplateInput) ToWhereQuery() ([]interface{}, []interface{}) {
	var whereQuery []interface{}
	var conds []interface{}

	if sdti.Limit < 0 || sdti.Limit > 100 {
		sdti.Limit = 100
	}

	if sdti.Offset < 0 {
		sdti.Offset = 0
	}

	if sdti.CreatedBy != uuid.Nil {
		whereQuery = append(whereQuery, "created_by = ?")
		conds = append(conds, sdti.CreatedBy)
	}

	if !reflect.ValueOf(sdti.CreatedAfter).IsZero() {
		whereQuery = append(whereQuery, "created_at > ?")
		conds = append(conds, sdti.CreatedAfter.UTC())
	}

	if sdti.IsActive != nil {
		whereQuery = append(whereQuery, "is_active = ?")
		conds = append(conds, *sdti.IsActive)
	}

	if sdti.IsLocked != nil {
		whereQuery = append(whereQuery, "is_locked = ?")
		conds = append(conds, *sdti.IsLocked)
	}

	return whereQuery, conds
}

// SearchSDTemplateOutput output for searching SD Template
type SearchSDTemplateOutput struct {
	Templates []*GeneratedSDTemplate `json:"templates"`
	Count     int                    `json:"count"`
}

// SDTemplateUsecase speech delay test template usecase
type SDTemplateUsecase interface {
	Create(ctx context.Context, input *SDTemplate) (*GeneratedSDTemplate, *common.Error)
	FindByID(ctx context.Context, id uuid.UUID) (*GeneratedSDTemplate, *common.Error)
	Search(ctx context.Context, input *SearchSDTemplateInput) (*SearchSDTemplateOutput, *common.Error)
	Update(ctx context.Context, id uuid.UUID, input *SDTemplate) (*GeneratedSDTemplate, *common.Error)
	Delete(ctx context.Context, id uuid.UUID) (*GeneratedSDTemplate, *common.Error)
	UndoDelete(ctx context.Context, id uuid.UUID) (*GeneratedSDTemplate, *common.Error)
	ChangeSDTemplateActiveStatus(ctx context.Context, id uuid.UUID, isActive bool) (*GeneratedSDTemplate, *common.Error)
}

// SDTemplateRepository speech delay test template repository
type SDTemplateRepository interface {
	Create(ctx context.Context, template *SpeechDelayTemplate) error
	FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*SpeechDelayTemplate, error)
	Search(ctx context.Context, input *SearchSDTemplateInput) ([]*SpeechDelayTemplate, error)
	Update(ctx context.Context, template *SpeechDelayTemplate, tx *gorm.DB) error
	Delete(ctx context.Context, id uuid.UUID) (*SpeechDelayTemplate, error)
	UndoDelete(ctx context.Context, id uuid.UUID) (*SpeechDelayTemplate, error)
}
