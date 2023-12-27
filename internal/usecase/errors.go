package usecase

import (
	"errors"

	"github.com/luckyAkbar/atec-api/internal/common"
)

var (
	// ErrInternal is returned when internal error occurs, such as database error, etc
	ErrInternal = errors.New("000000")

	// ErrResourceNotFound is returned when resource is not found, such as user not found, etc
	ErrResourceNotFound = errors.New("000003")

	// ErrEmailInputInvalid is returned when email input is invalid, such as empty subject, 0 receipients, etc
	ErrEmailInputInvalid = errors.New("email input invalid")

	// ErrInputSignUpInvalid will be returned when input is invalid, such as empty email, empty password, etc
	ErrInputSignUpInvalid = errors.New("001001")

	// ErrEmailAlreadyRegistered will be returned when user tried to sign up using an already registered email
	ErrEmailAlreadyRegistered = errors.New("001002")

	// ErrInputAccountVerificationInvalid will be returned when input is invalid
	ErrInputAccountVerificationInvalid = errors.New("001003")

	// ErrPinExpired is returned when the pin is expired by time, or no remaining attempts available
	ErrPinExpired = errors.New("001004")

	// ErrPinInvalid is returned when pin is invalid
	ErrPinInvalid = errors.New("001005")

	// ErrInputResetPasswordInvalid is returned when input is invalid
	ErrInputResetPasswordInvalid = errors.New("001006")

	// ErrForbiddenUpdateActiveStatus will be returned when trying to update Admin status or self updating status
	ErrForbiddenUpdateActiveStatus = errors.New("001007")

	// ErrUserIsBlocked is returned when user is blocked to access this service
	ErrUserIsBlocked = errors.New("002001")

	// ErrInvalidPassword is returned when password is invalid
	ErrInvalidPassword = errors.New("002002")

	// ErrInvalidLoginInput is returned when login input is invalid
	ErrInvalidLoginInput = errors.New("002003")

	// ErrInvalidLogoutInput is returned when login input is invalid
	ErrInvalidLogoutInput = errors.New("002004")

	// ErrAccessTokenExpired is returned when access token is expired
	ErrAccessTokenExpired = errors.New("002005")

	// ErrInvalidValidateChangePasswordSessionInput is returned when input is invalid
	ErrInvalidValidateChangePasswordSessionInput = errors.New("002006")

	// ErrResetPasswordSessionExpired is returned when reset password session is expired
	ErrResetPasswordSessionExpired = errors.New("002007")

	// ErrInvalidResetPasswordInput is returned when input is invalid
	ErrInvalidResetPasswordInput = errors.New("002008")

	// ErrSDTemplateInputInvalid is returned when input is invalid
	ErrSDTemplateInputInvalid = errors.New("003001")

	// ErrSDTemplateAlreadyLocked is returned when sd template is already locked
	ErrSDTemplateAlreadyLocked = errors.New("003002")

	// ErrSDTemplateCantBeActivated will be returned when trying to activate an invalid SD template
	ErrSDTemplateCantBeActivated = errors.New("003003")

	// ErrSDTemplateIsDeactivated will be returned when trying to create package from inactive template
	ErrSDTemplateIsDeactivated = errors.New("003004")

	// ErrSDTemplateIsAlreadyActive will be returned when trying to update an active template
	ErrSDTemplateIsAlreadyActive = errors.New("003005")

	// ErrSDPackageInputInvalid is returned when input is invalid
	ErrSDPackageInputInvalid = errors.New("004001")

	// ErrSDPackageAlreadyLocked is returned when sd package is already locked
	ErrSDPackageAlreadyLocked = errors.New("004002")

	// ErrSDPackageCantBeActivated will be returned when sd package fails on full validation
	ErrSDPackageCantBeActivated = errors.New("004003")

	// ErrSDPackageAlreadyActive will be returned if trying to update an active sd package
	ErrSDPackageAlreadyActive = errors.New("004004")

	// ErrSDPackageAlreadyDeactivated will be returned when the sd package is inactive
	ErrSDPackageAlreadyDeactivated = errors.New("004005")

	// ErrInvalidSDTestAnswer will be returned if any error found when submitting sd test answer
	ErrInvalidSDTestAnswer = errors.New("005001")

	// ErrForbiddenToSubmitSDTestAnswer will be returned when the test is not accepting any answer
	ErrForbiddenToSubmitSDTestAnswer = errors.New("005002")

	// ErrInvalidSubmitKey will be returned if the submit key is invalid
	ErrInvalidSubmitKey = errors.New("005003")

	// ErrInvalidViewHistoriesInput will be returned if the input is invalid
	ErrInvalidViewHistoriesInput = errors.New("005004")
)

var nilErr = &common.Error{
	Type: nil,
}
