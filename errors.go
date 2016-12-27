package ladon

import (
	"github.com/pkg/errors"
)

var (
	// ErrRequestDenied is returned when an access request can not be satisfied by any policy.
	ErrRequestDenied = errors.New("Request was denied by default")

	// ErrRequestForcefullyDenied is returned when an access request is explicitly denied by a policy.
	ErrRequestForcefullyDenied = errors.New("Request was forcefully denied")
)
