package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModel_SDPackage_PartialValidation(t *testing.T) {
	tests := []common.TestStructure{
		{
			Name: "all empty",
			Run: func() {
				in := &SDPackage{}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "package name empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "",
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "template id empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "sub group details len is 0",
			Run: func() {
				in := &SDPackage{
					PackageName:     "valid package name",
					TemplateID:      uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "sub group details name is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "",
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists length is 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name:                   "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.Question is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "",
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue length is 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question:        "valid question?",
									AnswersAndValue: []SDAnswerAndValue{},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Text is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text: "",
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is set to 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 0,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is set below 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: -100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is duplicated",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 100,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
		{
			Name: "ok minimalis",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.NoError(t, in.PartialValidation())
			},
		},
		{
			Name: "ok lah",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
						{
							Name: "another valid group name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1001,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
										{
											Text:  "pilihan ketiga, tapi ya begitulah",
											Value: 11,
										},
									},
								},
							},
						},
					},
				}
				assert.NoError(t, in.PartialValidation())
			},
		},
		{
			Name: "sub group name must be deeply unique",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "must be unique here",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
						{
							Name: "must be unique here",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.PartialValidation())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run()
		})
	}
}

func TestSDTemplate_SearchSDPackageInput_ToWhereQuery(t *testing.T) {
	trueVal := true
	falseVal := false
	tests := []common.TestStructure{
		{
			Name: "1",
			Run: func() {
				in := &SearchSDPackageInput{
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
				in := &SearchSDPackageInput{
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
				in := &SearchSDPackageInput{
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
		{
			Name: "3",
			Run: func() {
				in := &SearchSDPackageInput{
					CreatedBy:      uuid.New(),
					CreatedAfter:   time.Now().Add(time.Hour * -10).UTC(),
					Limit:          10,
					IsActive:       &trueVal,
					IsLocked:       &falseVal,
					IncludeDeleted: false,
					Offset:         10009,
					TemplateID:     uuid.New(),
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

func TestModel_SDPackage_FullValidation(t *testing.T) {
	tests := []common.TestStructure{
		{
			Name: "all empty",
			Run: func() {
				in := &SDPackage{}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "package name empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "",
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "template id empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "sub group details len is 0",
			Run: func() {
				in := &SDPackage{
					PackageName:     "valid package name",
					TemplateID:      uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "sub group details name is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "",
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists length is 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name:                   "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.Question is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "",
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue length is 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question:        "valid question?",
									AnswersAndValue: []SDAnswerAndValue{},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Text is empty",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text: "",
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is set to 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 0,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is set below 0",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: -100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "SubGroupDetails.QuestionAndAnswerLists.AnswersAndValue.Value is duplicated",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 100,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(nil))
			},
		},
		{
			Name: "template fail on full validation",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(&SpeechDelayTemplate{Template: &SDTemplate{}}))
			},
		},
		{
			Name: "template inactive",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(&SpeechDelayTemplate{
					IsActive: false,
					Template: &SDTemplate{
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
					}}))
			},
		},
		{
			Name: "template was deleted",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "valid name",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
					},
				}
				assert.Error(t, in.FullValidation(&SpeechDelayTemplate{
					IsActive:  true,
					DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
					Template: &SDTemplate{
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
					}}))
			},
		},
		{
			Name: "at least one sub group on package don't exist on template",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "this dont exists on template",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 99,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 100,
										},
									},
								},
							},
						},
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 2,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "at least one sub group package details exists, but not present on the template")
			},
		},
		{
			Name: "at least one sub group on template don't exist on package",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 2,
							},
							{
								Name:              "dont exists on package :()",
								QuestionCount:     1,
								AnswerOptionCount: 2,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "at least one sub group template details is not present on the package sub group details")
			},
		},
		{
			Name: "a group questions count don't match the specified one",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     2,
								AnswerOptionCount: 2,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "the number of questions on the package is not match with the template, group name: okelah expecting 2 got 1")
			},
		},
		{
			Name: "a group answer count is less than the specified one",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan ketigax",
											Value: 2,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 3,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 2,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "the number of answers on the package is not match with the template, group name: okelah expecting 2 got 3")
			},
		},
		{
			Name: "answer's value list is not complete. e.g 1,2,3 but got 1,2,4",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan ketigax",
											Value: 2,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 4,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 3,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "value list on group okelah was not complete. Missing value 3 and expecting range 1 - 3")
			},
		},
		{
			Name: "answer's value list is not complete. e.g 1,2,3 but got 1,2,4",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "wakuwaku",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 4,
										},

										{
											Text:  "pilihan ketigax",
											Value: 5,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "wakuwaku",
								QuestionCount:     1,
								AnswerOptionCount: 3,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "value list on group wakuwaku was not complete. Missing value 3 and expecting range 1 - 3")
			},
		},
		{
			Name: "ok",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 3,
							},
						},
					},
				})
				assert.NoError(t, err)
			},
		},
		{
			Name: "value list on a few answer and value are not ordered1",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
										{
											Text:  "pilihan ketigax",
											Value: 4, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     2,
								AnswerOptionCount: 3,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "value list on group okelah was not complete. Missing value 3 and expecting range 1 - 3")
			},
		},
		{
			Name: "value list on a few answer and value are not ordered2",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "okelah",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
										{
											Text:  "pilihan ketigax",
											Value: 4, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
						{
							Name: "kepribadian diri",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},
										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
										{
											Text:  "pilihan keempats",
											Value: 4, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     2,
								AnswerOptionCount: 3,
							},
							{
								Name:              "kepribadian diri",
								QuestionCount:     2,
								AnswerOptionCount: 4,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "value list on group okelah was not complete. Missing value 3 and expecting range 1 - 3")
			},
		},
		{
			Name: "value list on a few answer and value are not ordered3",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "kepribadian diri",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
										{
											Text:  "pilihan keempats",
											Value: 4,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 1212,
										},
										{
											Text:  "pilihan ketigax",
											Value: 79,
										},
										{
											Text:  "pilihan ketigax",
											Value: 99, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
						{
							Name: "kepribadian diri 2",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
										{
											Text:  "pilihan keempats",
											Value: 4,
										},
										{
											Text:  "pilihan kelimaax",
											Value: 99,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 3,
										},
										{
											Text:  "pilihan ketigax",
											Value: 2,
										},
										{
											Text:  "pilihan kelimax",
											Value: 5,
										},
										{
											Text:  "pilihan keepmats",
											Value: 4,
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "kepribadian diri",
								QuestionCount:     2,
								AnswerOptionCount: 4,
							},
							{
								Name:              "kepribadian diri 2",
								QuestionCount:     2,
								AnswerOptionCount: 5,
							},
						},
					},
				})
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "value list on group kepribadian diri was not complete. Missing value 4 and expecting range 1 - 4")
			},
		},
		{
			Name: "ok2",
			Run: func() {
				in := &SDPackage{
					PackageName: "valid package name",
					TemplateID:  uuid.New(),
					SubGroupDetails: []SDSubGroupDetail{
						{
							Name: "kepribadian diri",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 2,
										},

										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
										{
											Text:  "pilihan keempats",
											Value: 4,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 4,
										},
										{
											Text:  "pilihan ketigax",
											Value: 3,
										},
										{
											Text:  "pilihan ketigax",
											Value: 2, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
						{
							Name: "kepribadian diri 2",
							QuestionAndAnswerLists: []SDQuestionAndAnswers{
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 5,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 3,
										},

										{
											Text:  "pilihan ketigax",
											Value: 2,
										},
										{
											Text:  "pilihan keempats",
											Value: 4,
										},
										{
											Text:  "pilihan keempats",
											Value: 1,
										},
									},
								},
								{
									Question: "valid question?",
									AnswersAndValue: []SDAnswerAndValue{
										{
											Text:  "pilihan pertama",
											Value: 1,
										},
										{
											Text:  "pilihan kedua, tapi value nya sama",
											Value: 3,
										},
										{
											Text:  "pilihan ketigax",
											Value: 2,
										},
										{
											Text:  "pilihan ketigax",
											Value: 5, // here is jumping value. should be 3 and must cause error on validation
										},
										{
											Text:  "pilihan ketigax",
											Value: 4, // here is jumping value. should be 3 and must cause error on validation
										},
									},
								},
							},
						},
					},
				}
				err := in.FullValidation(&SpeechDelayTemplate{
					IsActive: true,
					Template: &SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []SDTemplateSubGroupDetail{
							{
								Name:              "kepribadian diri",
								QuestionCount:     2,
								AnswerOptionCount: 4,
							},
							{
								Name:              "kepribadian diri 2",
								QuestionCount:     2,
								AnswerOptionCount: 5,
							},
						},
					},
				})
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run()
		})
	}
}
