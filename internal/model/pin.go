package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

// Pin represent "pins" table
type Pin struct {
	ID          uuid.UUID
	Pin         string
	UserID      uuid.UUID
	ExpiredAt   time.Time
	FailedCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   null.Time
}

// PinRepository pin's repository
type PinRepository interface {
	Create(ctx context.Context, pin *Pin, tx *gorm.DB) error
	FindByID(ctx context.Context, id uuid.UUID) (*Pin, error)
}
