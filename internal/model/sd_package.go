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
	SubGroupDetails []SDSubGroupDetail `json:"subGroupDetails" validate:"required,min=1,dive"`
}

// PartialValidation will validate the SD Package. enough to be used for first time creating / just updating the SD Template
func (sdp *SDPackage) PartialValidation() error {
	return validator.Struct(sdp)
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

// SDPackageUsecase interface for SD package usecase
type SDPackageUsecase interface {
	Create(ctx context.Context, input *SDPackage) (*GeneratedSDPackage, *common.Error)
	FindByID(ctx context.Context, id uuid.UUID) (*GeneratedSDPackage, *common.Error)
}

// SDPackageRepository interface for SD package repository
type SDPackageRepository interface {
	Create(ctx context.Context, input *SpeechDelayPackage) error
	FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*SpeechDelayPackage, error)
}
