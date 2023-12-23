package model

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// SDTest the actual db table representation
type SDTest struct {
	ID         uuid.UUID
	PackageID  uuid.UUID
	UserID     uuid.NullUUID
	Answer     SDTestAnswer
	Result     SDTestResult
	FinishedAt null.Time
	OpenUntil  time.Time
	SubmitKey  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt
}

// ToRESTResponse will convert to sd test to api response
func (sdt *SDTest) ToRESTResponse(plainSubmitKey, packageName string, testQuestion map[string][]SDTestQuestion) *InitiateSDTestOutput {
	return &InitiateSDTestOutput{
		ID:           sdt.ID,
		PackageID:    sdt.PackageID,
		PackageName:  packageName,
		UserID:       sdt.UserID,
		OpenUntil:    sdt.OpenUntil,
		SubmitKey:    plainSubmitKey,
		CreatedAt:    sdt.CreatedAt,
		UpdatedAt:    sdt.UpdatedAt,
		TestQuestion: testQuestion,
		DeletedAt:    sdt.DeletedAt,
	}
}

// TableName must be implemented to correctly safe the sd test result to test_result table
func (sdt SDTest) TableName() string {
	return "test_results"
}

// Answer singulare answer per question
type Answer struct {
	Question string
	Answer   string
}

// TestAnswer will hold per group test answer
type TestAnswer struct {
	GroupName string   `json:"groupName"`
	Answers   []Answer `json:"answers"`
}

// SDTestAnswer will hold total sd test answer
type SDTestAnswer struct {
	TestAnswers []TestAnswer `json:"testAnswers"`
}

// Scan is a function to scan database value to CreateSDTemplateInput
func (sdta *SDTestAnswer) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue interface{}) (err error) {
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

	if err = json.Unmarshal(bytes, sdta); err != nil {
		return
	}

	return
}

// Value is a function to convert CreateSDTemplateInput to json
func (sdta SDTestAnswer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	return json.Marshal(fieldValue)
}

// SDTestGroupResultstruct sd test result per group
type SDTestGroupResultstruct struct {
	GroupName string `json:"groupName"`
	Result    int    `json:"result"`
}

// SDTestResult will hold the total result of sd test
type SDTestResult struct {
	Result []SDTestGroupResultstruct `json:"result"`
}

// Scan is a function to scan database value to CreateSDTemplateInput
func (sdtr *SDTestResult) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue interface{}) (err error) {
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

	if err = json.Unmarshal(bytes, sdtr); err != nil {
		return
	}

	return
}

// Value is a function to convert CreateSDTemplateInput to json
func (sdtr SDTestResult) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	return json.Marshal(fieldValue)
}

// InitiateSDTestInput input when initiating the sd test
type InitiateSDTestInput struct {
	UserID          uuid.NullUUID `json:"userID,omitempty"`
	PackageID       uuid.NullUUID `json:"packageID"`
	DurationMinutes time.Duration `json:"durationMinutes"`
}

// InitiateSDTestOutput output when initiating the sd test
type InitiateSDTestOutput struct {
	ID           uuid.UUID                   `json:"id"`
	PackageID    uuid.UUID                   `json:"packageID"`
	PackageName  string                      `json:"packageName"`
	UserID       uuid.NullUUID               `json:"userID,omitempty"`
	OpenUntil    time.Time                   `json:"openUntil"`
	SubmitKey    string                      `json:"submitKey"`
	CreatedAt    time.Time                   `json:"createdAt"`
	UpdatedAt    time.Time                   `json:"updatedAt"`
	TestQuestion map[string][]SDTestQuestion `json:"testQuestion"`
	DeletedAt    gorm.DeletedAt              `json:"deletedAt,omitempty"`
}

// SDTestUsecase usecase
type SDTestUsecase interface {
	Initiate(ctx context.Context, input *InitiateSDTestInput) (*InitiateSDTestOutput, *common.Error)
}

// SDTestRepository repository
type SDTestRepository interface {
	Create(ctx context.Context, test *SDTest, tx *gorm.DB) error
}
