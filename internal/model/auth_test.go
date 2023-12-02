package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthModel_LogInInput_Validate(t *testing.T) {
	t.Run("invalid email", func(t *testing.T) {
		in := LogInInput{
			Email:    "invalid . format",
			Password: "pwpwpwpwpwpwpw",
		}

		assert.Error(t, in.Validate())
	})

	t.Run("pw not long enough", func(t *testing.T) {
		in := LogInInput{
			Email:    "valid.format@gmail.test",
			Password: "2short",
		}

		assert.Error(t, in.Validate())
	})

	t.Run("ok", func(t *testing.T) {
		in := LogInInput{
			Email:    "valid.format@gmail.test",
			Password: "not2short",
		}

		assert.NoError(t, in.Validate())
	})
}
