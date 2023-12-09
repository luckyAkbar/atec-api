package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestSDTemplate_Validate(t *testing.T) {
	tests := []common.TestStructure{
		{
			Name:   "name empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name: "",
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "indication threshold empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                "ok",
					IndicationThreshold: 0,
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "positive indication empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                  "ok",
					IndicationThreshold:   10,
					PositiveIndiationText: "",
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "negative indication empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "",
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "sub group len 0",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails:        []SDTemplateSubGroupDetail{},
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "sub group exists, but name empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails: []SDTemplateSubGroupDetail{
						{
							Name: "",
						},
					},
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "sub group exists, but QuestionCount empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails: []SDTemplateSubGroupDetail{
						{
							Name:          "okelah",
							QuestionCount: 0,
						},
					},
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "sub group exists, but AnswerOptionCount empty",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails: []SDTemplateSubGroupDetail{
						{
							Name:              "okelah",
							QuestionCount:     10,
							AnswerOptionCount: 0,
						},
					},
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "sub group exists, but one of them invalid",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails: []SDTemplateSubGroupDetail{
						{
							Name:              "okelah",
							QuestionCount:     10,
							AnswerOptionCount: 3,
						},
						{
							Name: "",
						},
					},
				}
				err := in.Validate()
				assert.Error(t, err)
			},
		},
		{
			Name:   "ok",
			MockFn: func() {},
			Run: func() {
				in := &SDTemplate{
					Name:                   "ok",
					IndicationThreshold:    10,
					PositiveIndiationText:  "ok",
					NegativeIndicationText: "ok jg",
					SubGroupDetails: []SDTemplateSubGroupDetail{
						{
							Name:              "okelah",
							QuestionCount:     10,
							AnswerOptionCount: 3,
						},
						{
							Name:              "okeh juga",
							QuestionCount:     10,
							AnswerOptionCount: 5,
						},
					},
				}
				err := in.Validate()
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestSDTemplate_SearchSDTemplateInput_ToWhereQuery(t *testing.T) {
	trueVal := true
	falseVal := false
	tests := []common.TestStructure{
		{
			Name: "1",
			Run: func() {
				in := &SearchSDTemplateInput{
					CreatedBy: uuid.New(),
					Limit:     -10,
					Offset:    -10009,
				}
				where, conds := in.ToWhereQuery()
				assert.Equal(t, len(where), len(conds))
				assert.Equal(t, in.Limit, 100)
				assert.Equal(t, in.Offset, 0)
			},
		},
		{
			Name: "2",
			Run: func() {
				in := &SearchSDTemplateInput{
					CreatedBy:    uuid.New(),
					CreatedAfter: time.Now().Add(time.Hour * -10).UTC(),
					Limit:        10,
					Offset:       -10009,
				}
				where, conds := in.ToWhereQuery()
				assert.Equal(t, len(where), len(conds))
				assert.Equal(t, in.Limit, 10)
				assert.Equal(t, in.Offset, 0)
			},
		},
		{
			Name: "3",
			Run: func() {
				in := &SearchSDTemplateInput{
					CreatedBy:      uuid.New(),
					CreatedAfter:   time.Now().Add(time.Hour * -10).UTC(),
					Limit:          10,
					IsActive:       &trueVal,
					IsLocked:       &falseVal,
					IncludeDeleted: false,
					Offset:         10009,
				}
				where, conds := in.ToWhereQuery()
				assert.Equal(t, len(where), len(conds))
				assert.Equal(t, in.Limit, 10)
				assert.Equal(t, in.Offset, 10009)

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run()
		})
	}
}
