package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

// Role is enum for available role
type Role string

// list available roles
const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

// User represent "users" table
type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Username  string
	IsActive  bool
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// SignUpInput will be the request format to sign up
type SignUpInput struct {
	Username            string `json:"username" validate:"required"`
	Email               string `json:"email" validate:"required,email"`
	Password            string `json:"password" validate:"required,min=8"`
	PasswordConfimation string `json:"passwordConfirmation" validate:"required,min=8,eqfield=Password"`
}

// Validate validates struct
func (s *SignUpInput) Validate() error {
	return validator.Struct(s)
}

// SignUpResponse will be the returned response format when success signup
type SignUpResponse struct {
	PinValidationID   string    `json:"pinValidationID"`
	PinExpiredAt      time.Time `json:"pinExpiredAt"`
	RemainingAttempts int       `json:"remainingAttempts"`
}

// UserUsecase user's usecase
type UserUsecase interface {
	SignUp(ctx context.Context, input *SignUpInput) (*SignUpResponse, *common.Error)
}

// UserRepository user's repository
type UserRepository interface {
	Create(ctx context.Context, user *User, tx *gorm.DB) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}
