// Package model holds all the datatype representing database, and core data structure and its behaviour
package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	Deadline        null.Int
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

// IsAlreadyPastDeadline will report whether since e.CreatedTime until now already passed
// e.Deadline amount of seconds. If e.Deadline is left unset, will default to return false
func (e *Email) IsAlreadyPastDeadline() bool {
	if e.Deadline.Valid {
		now := time.Now().UTC()
		if now.Sub(e.CreatedAt) > time.Duration(e.Deadline.Int64)*time.Second {
			return true
		}
	}

	return false
}

// RegisterEmailInput input to register email
type RegisterEmailInput struct {
	Subject        string   `validate:"required"`
	Body           string   `validate:"required"`
	To             []string `validate:"min=1,unique,dive,email"`
	Cc             []string `validate:"omitempty,unique,dive,email"`
	Bcc            []string `validate:"omitempty,unique,dive,email"`
	DeadlineSecond int64
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
