package apperror

import (
	"github.com/pkg/errors"
)

var (
	ErrNotImplemented = errors.New("Not implemented")
	ErrSystemTryLater = errors.New("System error, please try again later")
	ErrRecordNotFound = errors.New("Record not found")
	ErrAccessDenied   = errors.New("Access denied")
	ErrTimeout        = errors.New("Timeout")
	ErrAlreadyExists  = errors.New("Record already exists")

	ErrBadRequest          = errors.New("Bad Request")
	ErrUnauthorized        = errors.New("Unauthorized")
	ErrInternalServerError = errors.New("Internal Server Error")

	ErrInvalidParameter = errors.New("Invalid parameter")
)
