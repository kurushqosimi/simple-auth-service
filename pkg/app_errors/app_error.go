package app_errors

type AppError struct {
	Code string
	Err  error
}

func (e *AppError) Error() string {
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Helper constructor
func NewAppError(code string, err error) *AppError {
	return &AppError{Code: code, Err: err}
}

// Optional: a quick check function
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}
