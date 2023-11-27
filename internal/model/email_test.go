package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailModel_RegisterEmailInput_Validate(t *testing.T) {
	t.Run("subject empty", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("body empty", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "",
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("receipient to is empty", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("receipient is not a valid email", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"invalid email", "invalid jg"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("receipient cc is not a valid email", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid@email.com", "valid1@email.com"},
			Cc:      []string{"invalid email", "invalid jg"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("receipient bcc is not a valid email", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid@email.com", "valid1@email.com"},
			Cc:      []string{"valid@email.com", "valid1@email.com"},
			Bcc:     []string{"invalid email", "invalid jg"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("duplicate to", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"duplicate@email.com", "duplicate@email.com"},
			Cc:      []string{"valid@email.com", "valid1@email.com"},
			Bcc:     []string{"valid@email.com", "valid1@email.com"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("duplicate cc", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid@email.com", "valid1@email.com"},
			Cc:      []string{"duplicate@email.com", "duplicate@email.com"},
			Bcc:     []string{"valid@email.com", "valid1@email.com"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("duplicate bcc", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid@email.com", "valid1@email.com"},
			Cc:      []string{"valid@email.com", "valid1@email.com"},
			Bcc:     []string{"duplicate@email.com", "duplicate@email.com"},
		}

		err := in.Validate()
		assert.Error(t, err)
	})

	t.Run("ok1", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid@email.com"},
		}

		err := in.Validate()
		assert.NoError(t, err)
	})

	t.Run("ok2", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid.to2@email.com"},
			Cc:      []string{"valid.cc1@email.com", "okemail.22@gmail.com"},
		}

		err := in.Validate()
		assert.NoError(t, err)
	})

	t.Run("ok3", func(t *testing.T) {
		in := RegisterEmailInput{
			Subject: "ada isi",
			Body:    "ada jg",
			To:      []string{"valid.to2@email.com"},
			Cc:      []string{"valid.cc1@email.com", "okemail.22@gmail.com"},
			Bcc:     []string{"valid.cc1@email.com", "okemail.22@gmail.com"},
		}

		err := in.Validate()
		assert.NoError(t, err)
	})
}
