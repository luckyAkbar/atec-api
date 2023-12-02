package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gorm.io/gorm"
)

// AccessToken represent access_tokens table
type AccessToken struct {
	ID         uuid.UUID
	Token      string
	UserID     uuid.UUID
	ValidUntil time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt
}

// ToLogInOutput convert access token to log in output with plain token
func (at *AccessToken) ToLogInOutput(plainToken string) *LogInOutput {
	return &LogInOutput{
		ID:         at.ID,
		Token:      plainToken,
		UserID:     at.UserID,
		ValidUntil: at.ValidUntil,
		CreatedAt:  at.CreatedAt,
		UpdatedAt:  at.UpdatedAt,
		DeletedAt:  at.DeletedAt,
	}
}

// LogInInput input for log in process
type LogInInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate validate struct
func (lii *LogInInput) Validate() error {
	return validator.Struct(lii)
}

// LogInOutput output of log in process
type LogInOutput struct {
	ID         uuid.UUID      `json:"id"`
	Token      string         `json:"token"`
	UserID     uuid.UUID      `json:"userID"`
	ValidUntil time.Time      `json:"validUntil"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt,omitempty"`
}

// LogOutInput input for log out process
type LogOutInput struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

// Validate validate struct
func (loi *LogOutInput) Validate() error {
	return validator.Struct(loi)
}

// AccessTokenRepository access token repository
type AccessTokenRepository interface {
	Create(ctx context.Context, at *AccessToken) error
	FindByToken(ctx context.Context, token string) (*AccessToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

// AuthUsecase auth usecase
type AuthUsecase interface {
	LogIn(ctx context.Context, input *LogInInput) (*LogInOutput, *common.Error)
	LogOut(ctx context.Context, input *LogOutInput) *common.Error
}
