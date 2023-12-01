package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Pin represent "pins" table
type Pin struct {
	ID                uuid.UUID
	Pin               string
	UserID            uuid.UUID
	ExpiredAt         time.Time
	RemainingAttempts int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
}

// IsExpired checks is the p is expired by time, or has zero or lower RemainingAttempts
func (p *Pin) IsExpired() bool {
	return p.ExpiredAt.Before(time.Now()) || p.RemainingAttempts <= 0
}

// PinRepository pin's repository
type PinRepository interface {
	Create(ctx context.Context, pin *Pin, tx *gorm.DB) error
	FindByID(ctx context.Context, id uuid.UUID) (*Pin, error)
	DecrementRemainingAttempts(ctx context.Context, id uuid.UUID) error
}
