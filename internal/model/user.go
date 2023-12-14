package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"encoding/json"

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

// UnmarshalText implements Text Unmarshaler
func (r *Role) UnmarshalText(text []byte) error {
	str := string(text)
	switch strings.ToUpper(str) {
	default:
		return fmt.Errorf("role: %s is not a valid role", str)
	case "ADMIN":
		*r = RoleAdmin
	case "USER":
		*r = RoleUser
	}

	return nil
}

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

// IsBlocked decide if the user is blocked or not by DeletedAt and IsActive attributes
func (u *User) IsBlocked() bool {
	return u.DeletedAt.Valid || !u.IsActive
}

// IsAdmin return true if Role is RoleAdmin, false otherwise
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// ToRESTResponse convert User to FindUserResponse which ease rest response generation
func (u *User) ToRESTResponse(plainEmail string) *FindUserResponse {
	return &FindUserResponse{
		ID:        u.ID,
		Email:     plainEmail,
		Password:  u.Password,
		Username:  u.Username,
		IsActive:  u.IsActive,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: u.DeletedAt,
	}
}

// FindUserResponse can be used to return the User data to API response. It's better to use this than directly return the User struct
type FindUserResponse struct {
	ID        uuid.UUID      `json:"id"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	Username  string         `json:"username"`
	IsActive  bool           `json:"isActive"`
	Role      Role           `json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty"`
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

// AccountVerificationInput will be the request format to verify account pin
type AccountVerificationInput struct {
	PinValidationID uuid.UUID `json:"pinValidationID" validate:"required,uuid4"`
	Pin             string    `json:"pin" validate:"required"`
}

// Validate validates struct
func (a *AccountVerificationInput) Validate() error {
	return validator.Struct(a)
}

// SuccessAccountVerificationResponse will be returned when pin varification is successfull
type SuccessAccountVerificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"isActive"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt null.Time `json:"deletedAt,omitempty"`
}

// FailedAccountVerificationResponse will be returned when pin varification is failed
type FailedAccountVerificationResponse struct {
	RemainingAttempts int `json:"remainingAttempts"`
}

// ChangePasswordSession will hold the data saved to validate change password
type ChangePasswordSession struct {
	UserID    uuid.UUID
	ExpiredAt time.Time
	CreatedAt time.Time
	CreatedBy uuid.UUID
}

// ToJSONString convert struct to json string
func (cps *ChangePasswordSession) ToJSONString() string {
	res, err := json.Marshal(cps)
	if err != nil {
		// will this ever be real?
		return ""
	}
	return string(res)
}

// IsExpired check whether the ChangePasswordSession is expired by its ExpiredAt against time.Now
func (cps *ChangePasswordSession) IsExpired() bool {
	return cps.ExpiredAt.Before(time.Now().UTC())
}

// InitiateResetPasswordOutput will be returned when initiate reset password is successfull
type InitiateResetPasswordOutput struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// SearchUserInput input
type SearchUserInput struct {
	Username       string `query:"username"`
	Email          string `query:"email"`
	IsActive       *bool  `query:"isActive"`
	IncludeDeleted bool   `query:"includeDeleted"`
	Role           Role   `query:"role"`
	Limit          int    `query:"limit"`
	Offset         int    `query:"offset"`
}

// ToWhereQuery convert SearchUserInput to where query and conditions. If limit is unset / set over 100, will be set to 100.
// If offset is unset / set under 0, will be set to 0.
func (sui *SearchUserInput) ToWhereQuery() ([]interface{}, []interface{}) {
	var whereQuery []interface{}
	var conds []interface{}

	if sui.Limit < 0 || sui.Limit > 100 {
		sui.Limit = 100
	}

	if sui.Offset < 0 {
		sui.Offset = 0
	}

	if sui.Username != "" {
		whereQuery = append(whereQuery, "username ILIKE ?")
		conds = append(conds, "%"+sui.Username+"%")
	}

	if sui.Email != "" {
		whereQuery = append(whereQuery, "email ILIKE ?")
		conds = append(conds, "%"+sui.Email+"%")
	}

	if sui.IsActive != nil {
		whereQuery = append(whereQuery, "is_active = ?")
		conds = append(conds, *sui.IsActive)
	}

	if sui.Role != "" {
		whereQuery = append(whereQuery, "role = ?")
		conds = append(conds, sui.Role)
	}

	return whereQuery, conds
}

// SearchUserOutput output
type SearchUserOutput struct {
	Users []*FindUserResponse `json:"users"`
	Count int                 `json:"count"`
}

// UserUsecase user's usecase
type UserUsecase interface {
	SignUp(ctx context.Context, input *SignUpInput) (*SignUpResponse, *common.Error)
	VerifyAccount(ctx context.Context, input *AccountVerificationInput) (*SuccessAccountVerificationResponse, *FailedAccountVerificationResponse, *common.Error)
	InitiateResetPassword(ctx context.Context, userID uuid.UUID) (*InitiateResetPasswordOutput, *common.Error)
	Search(ctx context.Context, input *SearchUserInput) (*SearchUserOutput, *common.Error)
}

// UserRepository user's repository
type UserRepository interface {
	Create(ctx context.Context, user *User, tx *gorm.DB) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateActiveStatus(ctx context.Context, id uuid.UUID, status bool) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateChangePasswordSession(ctx context.Context, key string, expiry time.Duration, session *ChangePasswordSession) error
	FindChangePasswordSession(ctx context.Context, key string) (*ChangePasswordSession, error)
	Update(ctx context.Context, user *User, tx *gorm.DB) error
	Search(ctx context.Context, input *SearchUserInput) ([]*User, error)
}
