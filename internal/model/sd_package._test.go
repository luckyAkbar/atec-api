package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/stretchr/testify/assert"
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
			Name: "ok juga",
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
