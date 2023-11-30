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
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
}
}

// PinRepository pin's repository
type PinRepository interface {
	Create(ctx context.Context, pin *Pin, tx *gorm.DB) error
	FindByID(ctx context.Context, id uuid.UUID) (*Pin, error)
}
