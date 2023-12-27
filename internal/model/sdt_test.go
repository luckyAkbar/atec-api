package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestSDT_SDTest_IsStillAcceptingAnswer(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		sdt := &SDTest{
			OpenUntil: time.Now().Add(time.Hour * -1),
		}
		err := sdt.IsStillAcceptingAnswer()
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "the test is already expired")
	})

	t.Run("already answered", func(t *testing.T) {
		sdt := &SDTest{
			OpenUntil:  time.Now().Add(time.Hour * 1),
			FinishedAt: null.NewTime(time.Now(), true),
		}
		err := sdt.IsStillAcceptingAnswer()
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "the test is already answered")
	})

	t.Run("ok", func(t *testing.T) {
		sdt := &SDTest{
			OpenUntil: time.Now().Add(time.Hour * 1),
		}
		err := sdt.IsStillAcceptingAnswer()
		assert.NoError(t, err)
	})
}

func TestSDT_SDTestAnswer_DoGradingProcess(t *testing.T) {
	p := &SDPackage{
		PackageName: "test",
		TemplateID:  uuid.New(),
		SubGroupDetails: []SDSubGroupDetail{
			{
				Name: "kepribadian diri",
				QuestionAndAnswerLists: []SDQuestionAndAnswers{
					{
						Question: "apakah anak percaya diri?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat percaya diri",
								Value: 1,
							},
							{
								Text:  "cukup percaya diri",
								Value: 2,
							},
							{
								Text:  "tidak percaya diri",
								Value: 3,
							},
						},
					},
					{
						Question: "apakah anak memiliki jati diri?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat punya jati diri",
								Value: 1,
							},
							{
								Text:  "cukup punya jati diri",
								Value: 2,
							},
							{
								Text:  "tidak punya jati diri",
								Value: 3,
							},
						},
					},
				},
			},
			{
				Name: "kemampuan memasak",
				QuestionAndAnswerLists: []SDQuestionAndAnswers{
					{
						Question: "apakah anak bisa memasak nasi?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat bisa",
								Value: 1,
							},
							{
								Text:  "cukup bisa",
								Value: 2,
							},
							{
								Text:  "tidak bisa",
								Value: 3,
							},
						},
					},
					{
						Question: "apakah anak bisa memasak bubur?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat bisa",
								Value: 1,
							},
							{
								Text:  "cukup bisa",
								Value: 2,
							},
							{
								Text:  "tidak bisa",
								Value: 3,
							},
						},
					},
				},
			},
			{
				Name: "kemampuan berhitung",
				QuestionAndAnswerLists: []SDQuestionAndAnswers{
					{
						Question: "apakah anak bisa menghitung hingga 100?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat bisa",
								Value: 1,
							},
							{
								Text:  "cukup bisa",
								Value: 2,
							},
							{
								Text:  "tidak bisa",
								Value: 3,
							},
						},
					},
					{
						Question: "apakah anak bisa kali kali-an?",
						AnswersAndValue: []SDAnswerAndValue{
							{
								Text:  "sangat bisa",
								Value: 1,
							},
							{
								Text:  "cukup bisa",
								Value: 2,
							},
							{
								Text:  "tidak bisa",
								Value: 3,
							},
						},
					},
				},
			},
		},
	}

	t.Run("struct invalid", func(t *testing.T) {
		s := &SDTestAnswer{}
		_, err := s.DoGradingProcess(&SDPackage{})
		assert.Error(t, err)
	})
	t.Run("struct invalid: empty array", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{},
		}
		_, err := s.DoGradingProcess(&SDPackage{})
		assert.Error(t, err)
	})
	t.Run("1 sub group is not present / answered", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "group kepribadian diri is not found on answers list")
	})

	t.Run("2 sub group is not present / answered", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "cukup punya jati diri",
						},
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "group kemampuan memasak is not found on answers list")
	})
	t.Run("all groups present, but some are not required on package", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "apa ini ya?",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "cukup punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "unknown group: apa ini ya? is not required on package")
	})

	t.Run("all groups present, but a question is not answered", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "cukup punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						// missing here
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "question apakah anak bisa memasak bubur? is still not answered")
	})

	t.Run("all groups present, but a group question is not required on package", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "cukup punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
						{
							Question: "dilarang ada pertanyaan tambahan ye",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "question dilarang ada pertanyaan tambahan ye is not found on package")
	})

	t.Run("all groups present, but a group question's answer is not present on package", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "jawaban ini tuh gaada di package",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		_, err := s.DoGradingProcess(p)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "answer jawaban ini tuh gaada di package is not found on package")
	})

	t.Run("ok1", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "sangat percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "cukup punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		res, err := s.DoGradingProcess(p)
		assert.NoError(t, err)

		assert.Equal(t, res[0].GroupName, "kepribadian diri")
		assert.Equal(t, res[0].Result, 3)

		assert.Equal(t, res[1].GroupName, "kemampuan berhitung")
		assert.Equal(t, res[1].Result, 2)

		assert.Equal(t, res[2].GroupName, "kemampuan memasak")
		assert.Equal(t, res[2].Result, 2)
	})

	t.Run("ok2", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "tidak percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "tidak punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "sangat bisa",
						},
					},
				},
			},
		}
		res, err := s.DoGradingProcess(p)
		assert.NoError(t, err)

		assert.Equal(t, res[0].GroupName, "kepribadian diri")
		assert.Equal(t, res[0].Result, 6)

		assert.Equal(t, res[1].GroupName, "kemampuan berhitung")
		assert.Equal(t, res[1].Result, 2)

		assert.Equal(t, res[2].GroupName, "kemampuan memasak")
		assert.Equal(t, res[2].Result, 2)
	})

	t.Run("ok3", func(t *testing.T) {
		s := &SDTestAnswer{
			TestAnswers: []*TestAnswer{
				{
					GroupName: "kepribadian diri",
					Answers: []Answer{
						{
							Question: "apakah anak percaya diri?",
							Answer:   "tidak percaya diri",
						},
						{
							Question: "apakah anak memiliki jati diri?",
							Answer:   "tidak punya jati diri",
						},
					},
				},
				{
					GroupName: "kemampuan berhitung",
					Answers: []Answer{
						{
							Question: "apakah anak bisa menghitung hingga 100?",
							Answer:   "sangat bisa",
						},
						{
							Question: "apakah anak bisa kali kali-an?",
							Answer:   "sangat bisa",
						},
					},
				},
				{
					GroupName: "kemampuan memasak",
					Answers: []Answer{
						{
							Question: "apakah anak bisa memasak nasi?",
							Answer:   "tidak bisa",
						},
						{
							Question: "apakah anak bisa memasak bubur?",
							Answer:   "tidak bisa",
						},
					},
				},
			},
		}
		res, err := s.DoGradingProcess(p)
		assert.NoError(t, err)

		assert.Equal(t, res[0].GroupName, "kepribadian diri")
		assert.Equal(t, res[0].Result, 6)

		assert.Equal(t, res[1].GroupName, "kemampuan berhitung")
		assert.Equal(t, res[1].Result, 2)

		assert.Equal(t, res[2].GroupName, "kemampuan memasak")
		assert.Equal(t, res[2].Result, 6)
	})
}
