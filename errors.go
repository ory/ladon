package ladon

import (
	"net/http"

	"github.com/go-errors/errors"
)

// Error is an error object.
type Error struct {
	*errors.Error

	// Code is the error's http status code.
	Code int
}

var (
	// ErrForbidden is returned when access is forbidden.
	ErrForbidden = &Error{
		Error: errors.New("Forbidden"),
		Code:  http.StatusForbidden,
	}
)
