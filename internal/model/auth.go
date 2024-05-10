package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"gopkg.in/guregu/null.v4"
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

// IsExpired reports whether the access token is expired, either the ValidUntil time is in the past, or the DeletedAt is not null.
func (at *AccessToken) IsExpired() bool {
	return at.ValidUntil.Before(time.Now().UTC()) || at.DeletedAt.Valid
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

type authCtxKey string

var (
	authUserCtxKey authCtxKey = "github.com/luckyAkbar/atec-api/internal/model:AuthUser"
)

// AuthUser represents all the necessary data to be passed on context
type AuthUser struct {
	UserID      uuid.UUID
	AccessToken string
	Role        Role
}

// IsAdmin return whether Role is RoleAdmin
func (a *AuthUser) IsAdmin() bool {
	return a.Role == RoleAdmin
}

// SetUserToCtx set user to context
func SetUserToCtx(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, authUserCtxKey, user)
}

// GetUserFromCtx get user from context
func GetUserFromCtx(ctx context.Context) *AuthUser {
	user, ok := ctx.Value(authUserCtxKey).(AuthUser)
	if !ok {
		return nil
	}
	return &user
}

// ResetPasswordInput input for reset password process
type ResetPasswordInput struct {
	Key                 string `json:"key" validate:"required"`
	Password            string `json:"password" validate:"required,min=8"`
	PasswordConfimation string `json:"passwordConfirmation" validate:"required,min=8,eqfield=Password"`
}

// Validate validates struct
func (s *ResetPasswordInput) Validate() error {
	return validator.Struct(s)
}

// ResetPasswordResponse will be returned when reset password success
type ResetPasswordResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"isActive"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt null.Time `json:"deletedAt,omitempty"`
}

// AccessTokenRepository access token repository
type AccessTokenRepository interface {
	Create(ctx context.Context, at *AccessToken) error
	FindByToken(ctx context.Context, token string) (*AccessToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByIDs(ctx context.Context, ids []uuid.UUID, hardDelete bool) error
	FindCredentialByToken(ctx context.Context, token string) (*AccessToken, *User, error)
	DeleteByUserID(ctx context.Context, id uuid.UUID, tx *gorm.DB) error
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]AccessToken, error)
	DeleteCredentialsFromCache(ctx context.Context, tokens []string) error
}

// AuthUsecase auth usecase
type AuthUsecase interface {
	LogIn(ctx context.Context, input *LogInInput) (*LogInOutput, *common.Error)
	LogOut(ctx context.Context) *common.Error
	ValidateAccess(ctx context.Context, token string) (*AuthUser, *common.Error)
	ValidateResetPasswordSession(ctx context.Context, key string) *common.Error
	ResetPassword(ctx context.Context, input *ResetPasswordInput) (*ResetPasswordResponse, *common.Error)
}
