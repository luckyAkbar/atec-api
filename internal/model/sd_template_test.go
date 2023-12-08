package model

import (
	"testing"

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
