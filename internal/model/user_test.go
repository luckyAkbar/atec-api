package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignUpInput_Validate(t *testing.T) {
	t.Run("empty username", func(t *testing.T) {
		in := SignUpInput{
			Username: "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("empty email", func(t *testing.T) {
		in := SignUpInput{
			Username: "username",
			Email:    "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid email", func(t *testing.T) {
		in := SignUpInput{
			Username: "username",
			Email:    "invalid email @gmail.com",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		in := SignUpInput{
			Username: "username",
			Email:    "email@gmail.com",
			Password: "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("empty password confirmation", func(t *testing.T) {
		in := SignUpInput{
			Username:            "username",
			Email:               "email@gmail.com",
			Password:            "abcdefgabcdefgabcdefg",
			PasswordConfimation: "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("password mismatch", func(t *testing.T) {
		in := SignUpInput{
			Username:            "username",
			Email:               "email@gmail.com",
			Password:            "abcdefgabcdefgabcdefg",
			PasswordConfimation: "123abcdefgabcdefgabcdefg",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("password less than 8 chars", func(t *testing.T) {
		in := SignUpInput{
			Username:            "username",
			Email:               "email@gmail.com",
			Password:            "abc",
			PasswordConfimation: "abc",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		in := SignUpInput{
			Username:            "username",
			Email:               "email@gmail.com",
			Password:            "abc1238chars",
			PasswordConfimation: "abc1238chars",
		}

		err := in.Validate()
		assert.NoError(t, err)
	})
}
