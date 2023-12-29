package model

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"errors"

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

// IsStillAcceptingAnswer will return error if the OpenUntil is pass now
// or the FinishedAt is already set. Return nil otherwise
func (sdt *SDTest) IsStillAcceptingAnswer() error {
	if sdt.OpenUntil.Before(time.Now()) {
		return errors.New("the test is already expired")
	}

	if sdt.FinishedAt.Valid {
		return errors.New("the test is already answered")
	}

	return nil
}

// ToInitiateSDTestOutput will convert to sd test to api response
func (sdt *SDTest) ToInitiateSDTestOutput(plainSubmitKey, packageName string, testQuestion map[string][]SDTestQuestion) *InitiateSDTestOutput {
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

// ToSubmitTestOutput will convert to SubmitSDTestOutput
func (sdt *SDTest) ToSubmitTestOutput(packageName, plainSubmitKey string, testQuestion map[string][]SDTestQuestion) *SubmitSDTestOutput {
	return &SubmitSDTestOutput{
		ID:           sdt.ID,
		PackageID:    sdt.PackageID,
		PackageName:  packageName,
		UserID:       sdt.UserID,
		Answer:       sdt.Answer,
		Result:       sdt.Result,
		OpenUntil:    sdt.OpenUntil,
		SubmitKey:    plainSubmitKey,
		FinishedAt:   sdt.FinishedAt.Time.UTC(),
		CreatedAt:    sdt.CreatedAt,
		UpdatedAt:    sdt.UpdatedAt,
		TestQuestion: testQuestion,
		DeletedAt:    sdt.DeletedAt,
	}
}

// ToViewHistoriesOutput convert SDTest to ViewHistoriesOutput
func (sdt *SDTest) ToViewHistoriesOutput() ViewHistoriesOutput {
	return ViewHistoriesOutput{
		ID:         sdt.ID,
		PackageID:  sdt.PackageID,
		UserID:     sdt.UserID,
		OpenUntil:  sdt.OpenUntil,
		FinishedAt: sdt.FinishedAt.Time.UTC(),
		CreatedAt:  sdt.CreatedAt,
		UpdatedAt:  sdt.UpdatedAt,
		DeletedAt:  sdt.DeletedAt,
		Answer:     sdt.Answer,
		Result:     sdt.Result,
	}

}

// TableName must be implemented to correctly safe the sd test result to test_result table
func (sdt SDTest) TableName() string {
	return "test_results"
}

// Answer singular answer per question
type Answer struct {
	Question string `json:"question" validate:"required"`
	Answer   string `json:"answer" validate:"required"`

	options []SDAnswerAndValue `json:"-"`
}

// getAnswerValue will return the answer value from the options.
// if the answer is not found on the options, will return error
func (a *Answer) getAnswerValue() (int, error) {
	for _, o := range a.options {
		if o.Text == a.Answer {
			return o.Value, nil
		}
	}

	return 0, fmt.Errorf("answer %s is not found on package", a.Answer)
}

// TestAnswer will hold per group test answer
type TestAnswer struct {
	GroupName string   `json:"groupName"  validate:"required"`
	Answers   []Answer `json:"answers"  validate:"required,dive"`

	sdqna []SDQuestionAndAnswers `json:"-"`
}

// gradeGroupResult will try to validate the questions to ensure all the packages question are answered,
// and also validate the answer to ensure the answer is exists on the options.
// On success, will return the group's result. On error, will return error.
func (ta *TestAnswer) gradeGroupResult() (SDTestGroupResult, error) {
	if err := ta.ensureAllQuestionOnPackageAnswered(); err != nil {
		return SDTestGroupResult{}, err
	}

	// ensure all question are exists on package and validate the answer also exists on the options
	result := 0
	for _, a := range ta.Answers {
		found := false
		for _, qna := range ta.sdqna {
			if qna.Question == a.Question {
				found = true
				a.options = qna.AnswersAndValue
				val, err := a.getAnswerValue()
				if err != nil {
					return SDTestGroupResult{}, err
				}

				result += val
				break
			}
		}

		if !found {
			return SDTestGroupResult{}, fmt.Errorf("question %s is not found on package", a.Question)
		}
	}

	return SDTestGroupResult{
		GroupName: ta.GroupName,
		Result:    result,
	}, nil
}

func (ta *TestAnswer) ensureAllQuestionOnPackageAnswered() error {
	// ensure all question on package are exists
	for _, qna := range ta.sdqna {
		found := false
		for _, a := range ta.Answers {
			if qna.Question == a.Question {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("question %s is still not answered", qna.Question)
		}
	}

	return nil
}

// SDTestAnswer will hold total sd test answer
type SDTestAnswer struct {
	TestAnswers []*TestAnswer `json:"testAnswers" validate:"required,min=1,dive"`
}

// DoGradingProcess will validate the answer to make sure all the questions are answered.
func (sdta *SDTestAnswer) DoGradingProcess(p *SDPackage) ([]SDTestGroupResult, error) {
	if err := validator.Struct(sdta); err != nil {
		return nil, err
	}

	if err := sdta.ensureAllSubGroupArePresent(p); err != nil {
		return nil, err
	}

	// ensure only defined groups are present
	// also ensure that all the questions are answered with valid answer option
	for _, ta := range sdta.TestAnswers {
		found := false
		for _, g := range p.SubGroupDetails {
			if ta.GroupName == g.Name {
				found = true
				ta.sdqna = g.QuestionAndAnswerLists
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("unknown group: %s is not required on package", ta.GroupName)
		}
	}

	// ensure all group's questions are present and at least answered
	groupResults := make([]SDTestGroupResult, 0)
	for _, ta := range sdta.TestAnswers {
		r, err := ta.gradeGroupResult()
		if err != nil {
			return nil, err
		}

		groupResults = append(groupResults, r)
	}

	return groupResults, nil
}

func (sdta *SDTestAnswer) ensureAllSubGroupArePresent(p *SDPackage) error {
	// ensure that all the sub group details are present
	for _, g := range p.SubGroupDetails {
		found := false
		for _, ta := range sdta.TestAnswers {
			if ta.GroupName == g.Name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("group %s is not found on answers list", g.Name)
		}
	}

	return nil
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

// SDTestGroupResult sd test result per group
type SDTestGroupResult struct {
	GroupName string `json:"groupName"`
	Result    int    `json:"result"`
}

// SDTestResult will hold the total result of sd test
type SDTestResult struct {
	Result []SDTestGroupResult `json:"result"`
	Total  int                 `json:"total"`
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

// SubmitSDTestInput input to submit sd test
type SubmitSDTestInput struct {
	TestID    uuid.UUID     `json:"testID" validate:"required"`
	SubmitKey string        `json:"submitKey" validate:"required"`
	Answers   *SDTestAnswer `json:"answers" validate:"required"`
}

// Validate validate struct
func (sdtti *SubmitSDTestInput) Validate() error {
	return validator.Struct(sdtti)
}

// SubmitSDTestOutput output from submit sd test
type SubmitSDTestOutput struct {
	ID           uuid.UUID                   `json:"id"`
	PackageID    uuid.UUID                   `json:"packageID"`
	PackageName  string                      `json:"packageName"`
	UserID       uuid.NullUUID               `json:"userID,omitempty"`
	Answer       SDTestAnswer                `json:"answer"`
	Result       SDTestResult                `json:"result"`
	OpenUntil    time.Time                   `json:"openUntil"`
	FinishedAt   time.Time                   `json:"finishedAt"`
	SubmitKey    string                      `json:"submitKey"`
	CreatedAt    time.Time                   `json:"createdAt"`
	UpdatedAt    time.Time                   `json:"updatedAt"`
	TestQuestion map[string][]SDTestQuestion `json:"testQuestion"`
	DeletedAt    gorm.DeletedAt              `json:"deletedAt,omitempty"`
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

// ViewHistoriesInput input
type ViewHistoriesInput struct {
	UserID            uuid.NullUUID `query:"userID"`
	PackageID         uuid.NullUUID `query:"packageID"`
	CreatedAfter      null.Time     `query:"createdAfter"`
	IncludeUnfinished bool          `query:"includeUnfinished"`
	IncludeDeleted    bool          `query:"includeDeleted"`
	Limit             int           `query:"limit"`
	Offset            int           `query:"offset"`
}

// ToWhereQuery convert input to search query
func (vhi *ViewHistoriesInput) ToWhereQuery() ([]interface{}, []interface{}) {
	var whereQuery []interface{}
	var conds []interface{}

	if vhi.Limit < 0 || vhi.Limit > 100 {
		vhi.Limit = 100
	}

	if vhi.Offset < 0 {
		vhi.Offset = 0
	}

	if vhi.UserID.Valid {
		whereQuery = append(whereQuery, "user_id = ?")
		conds = append(conds, vhi.UserID)
	}

	if vhi.PackageID.Valid {
		whereQuery = append(whereQuery, "package_id = ?")
		conds = append(conds, vhi.PackageID)
	}

	if vhi.CreatedAfter.Valid {
		whereQuery = append(whereQuery, "created_at > ?")
		conds = append(conds, vhi.CreatedAfter.Time)
	}

	return whereQuery, conds
}

// ViewHistoriesOutput output from submit sd test
type ViewHistoriesOutput struct {
	ID         uuid.UUID      `json:"id"`
	PackageID  uuid.UUID      `json:"packageID"`
	UserID     uuid.NullUUID  `json:"userID,omitempty"`
	Answer     SDTestAnswer   `json:"answer"`
	Result     SDTestResult   `json:"result"`
	OpenUntil  time.Time      `json:"openUntil"`
	FinishedAt time.Time      `json:"finishedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt,omitempty"`
}

// StatsComponent will define what will be the statistic component
type StatsComponent struct {
	TestResultID   uuid.UUID `json:"testResultID"`
	PackageID      uuid.UUID `json:"packageID"`
	ResultPoint    int       `json:"resultPoint"`
	PackageName    string    `json:"packageName"`
	TestFinishedAt time.Time `json:"testFinishedAt"`
}

// SDTestStatistic will hold the structure of sd test statistic
type SDTestStatistic struct {
	TemplateID             uuid.UUID        `json:"templateID"`
	TemplateName           string           `json:"templateName"`
	IndicationThreshold    int              `json:"indicationThreshold"`
	PositiveIndiationText  string           `json:"positiveIndicationText"`
	NegativeIndicationText string           `json:"negativeIndicationText"`
	Stats                  []StatsComponent `json:"stats"`
}

// SDTestUsecase usecase
type SDTestUsecase interface {
	Initiate(ctx context.Context, input *InitiateSDTestInput) (*InitiateSDTestOutput, *common.Error)
	Submit(ctx context.Context, input *SubmitSDTestInput) (*SubmitSDTestOutput, *common.Error)
	Histories(ctx context.Context, input *ViewHistoriesInput) ([]ViewHistoriesOutput, *common.Error)
	Statistic(ctx context.Context, userID uuid.UUID) ([]SDTestStatistic, *common.Error)
}

// SDTestRepository repository
type SDTestRepository interface {
	Create(ctx context.Context, test *SDTest, tx *gorm.DB) error
	FindByID(ctx context.Context, id uuid.UUID) (*SDTest, error)
	Update(ctx context.Context, test *SDTest, tx *gorm.DB) error
	Search(ctx context.Context, input *ViewHistoriesInput) ([]*SDTest, error)
	Statistic(ctx context.Context, userID uuid.UUID) ([]SDTestStatistic, error)
}
