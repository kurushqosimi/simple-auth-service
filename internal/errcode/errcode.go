// Package errcode holds all error codes and their default messages.
// Typically, you’d place this in something like `internal/errcode/errcode.go` or `pkg/errcode/errcode.go`.
package errcode

const (
	ErrInvalidRequest     = "invalid_request"
	ErrUnauthorized       = "unauthorized"
	ErrForbidden          = "forbidden"
	ErrNotFound           = "not_found"
	ErrConflict           = "conflict"
	ErrInternal           = "internal_error"
	ErrEmailAlreadyExists = "email_already_exists"
	ErrAccountCreated     = "account_created"
	ErrOTPNotFound        = "otp_not_found"
	ErrOTPInvalid         = "invalid_otp"
	ErrInvalidPassword    = "invalid_password"
	ErrLoginRedirect      = "login_redirect"
)

var errorMessages = map[string]string{
	ErrInvalidRequest:     "The request is invalid or malformed",
	ErrUnauthorized:       "Missing or invalid authentication credentials",
	ErrForbidden:          "You do not have permission to access this resource",
	ErrNotFound:           "The requested resource was not found",
	ErrConflict:           "A resource conflict occurred (e.g., duplicate data)",
	ErrInternal:           "An unexpected server error occurred",
	ErrEmailAlreadyExists: "A user with this email already exists. Please try a different email.",
	ErrAccountCreated:     "An account was created, but email with activation code was not sent. Please, contact support.",
	ErrOTPNotFound:        "For this user code was not found. Please try again.",
	ErrOTPInvalid:         "The code you provided is invalid",
	ErrInvalidPassword:    "The password you provided is incorrect",
}

// GetErrorMessage returns a standard “user-friendly” message
// for a given error code. If the code is unknown, it returns a default.
func GetErrorMessage(errorCode string) string {
	if msg, ok := errorMessages[errorCode]; ok {
		return msg
	}
	return "Unexpected error" // fallback for unknown codes
}
