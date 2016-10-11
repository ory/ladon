package ladon

import (
	"github.com/pkg/errors"
)

var (
	// ErrForbidden is returned when access is forbidden.
	ErrForbidden = errors.New("Forbidden")
)
