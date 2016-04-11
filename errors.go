package ladon

import (
	"github.com/go-errors/errors"
	"net/http"
)

type Error struct {
	*errors.Error
	Code int
}

var (
	ErrForbidden = &Error{
		Error: errors.New("Forbidden"),
		Code: http.StatusForbidden,
	}
)
