// Package model holds all the datatype representing database, and core data structure and its behaviour
package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/sweet-go/stdlib/mail"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

// Email represent emails table structure from database
type Email struct {
	ID              uuid.UUID
	Subject         string
	Body            string
	To              pq.StringArray `gorm:"type:varchar(255)[]"`
	Cc              pq.StringArray `gorm:"type:varchar(255)[]"`
	Bcc             pq.StringArray `gorm:"type:varchar(255)[]"`
	SentAt          null.Time
	ClientSignature null.String
	Metadata        null.String
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt
}

// GenericReceipientsTo convert To to model.GenericReceipient
func (e *Email) GenericReceipientsTo() []mail.GenericReceipient {
	var recipients []mail.GenericReceipient
	for _, r := range e.To {
		recipients = append(recipients, mail.GenericReceipient{
			Name:  r,
			Email: r,
		})
	}
	return recipients
}

// GenericReceipientsCc convert Cc to model.GenericReceipient
func (e *Email) GenericReceipientsCc() []mail.GenericReceipient {
	var recipients []mail.GenericReceipient
	for _, r := range e.Cc {
		recipients = append(recipients, mail.GenericReceipient{
			Name:  r,
			Email: r,
		})
	}
	return recipients
}

// GenericReceipientsBcc convert Bcc to model.GenericReceipient
func (e *Email) GenericReceipientsBcc() []mail.GenericReceipient {
	var recipients []mail.GenericReceipient
	for _, r := range e.Bcc {
		recipients = append(recipients, mail.GenericReceipient{
			Name:  r,
			Email: r,
		})
	}
	return recipients
}

// Encrypt is helper function to be called to ensure all necessary
// field in the Email are encrypted before storing it to data storage
func (e *Email) Encrypt(cryptor common.SharedCryptor) error {
	bodyEnc, err := cryptor.Encrypt(e.Body)
	if err != nil {
		return err
	}

	tosEnc := []string{}
	for _, to := range e.To {
		toEnc, err := cryptor.Encrypt(to)
		if err != nil {
			return err
		}

		tosEnc = append(tosEnc, toEnc)
	}

	e.Body = bodyEnc
	e.To = tosEnc

	return nil
}

// Decrypt is a helper function to decrypted all the fields that previously encrypted
// by Encrypt function
func (e *Email) Decrypt(cryptor common.SharedCryptor) error {
	bodyDec, err := cryptor.Decrypt(e.Body)
	if err != nil {
		return err
	}

	tosDec := []string{}
	for _, to := range e.To {
		toDec, err := cryptor.Decrypt(to)
		if err != nil {
			return err
		}

		tosDec = append(tosDec, toDec)
	}

	e.Body = bodyDec
	e.To = tosDec

	return nil
}

// RegisterEmailInput input to register email
type RegisterEmailInput struct {
	Subject string   `validate:"required"`
	Body    string   `validate:"required"`
	To      []string `validate:"min=1,unique,dive,email"`
	Cc      []string `validate:"omitempty,unique,dive,email"`
	Bcc     []string `validate:"omitempty,unique,dive,email"`
}

// Validate run all the validation function to ensure all the input values are following the defined rules here
func (e *RegisterEmailInput) Validate() error {
	return validator.Struct(e)
}

// EmailUsecase represent the email's usecase interface
type EmailUsecase interface {
	Register(ctx context.Context, input *RegisterEmailInput) (*Email, error)
}

// EmailRepository represent the email's repository interface
type EmailRepository interface {
	Create(ctx context.Context, email *Email) error
	FindByID(ctx context.Context, id uuid.UUID) (*Email, error)
	Update(ctx context.Context, email *Email) error
}
