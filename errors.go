package ladon

import (
	"net/http"

	"github.com/go-errors/errors"
)

type Error struct {
	*errors.Error
	Code int
}

var (
	ErrForbidden = &Error{
		Error: errors.New("Forbidden"),
		Code:  http.StatusForbidden,
	}
)
