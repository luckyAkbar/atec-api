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
	QuestionCount     int    `json:"questionCount" validate:"required"`
	AnswerOptionCount int    `json:"answerOptionCount" validate:"required"`
}

// SDTemplate define what the full SD test template will look like
type SDTemplate struct {
	Name                   string                     `json:"name" validate:"required,max=255"`
	IndicationThreshold    int                        `json:"indicationThreshold" validate:"required"`
	PositiveIndiationText  string                     `json:"positiveIndicationText" validate:"required"`
	NegativeIndicationText string                     `json:"negativeIndicationText" validate:"required"`
	SubGroupDetails        []SDTemplateSubGroupDetail `json:"subGroupDetails" validate:"min=1,dive"`
}

// Validate validate struct. For now, it's enough using built in validation rules provided by validation package
func (csdti *SDTemplate) Validate() error {
	return validator.Struct(csdti)
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

// SDTemplateUsecase speech delay test template usecase
type SDTemplateUsecase interface {
	Create(ctx context.Context, input *SDTemplate) (*GeneratedSDTemplate, *common.Error)
}

// SDTemplateRepository speech delay test template repository
type SDTemplateRepository interface {
	Create(ctx context.Context, template *SpeechDelayTemplate) error
}
