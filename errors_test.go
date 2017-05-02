package ladon

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrResourceNotFound(t *testing.T) {
	assert.EqualError(t, NewErrResourceNotFound(errors.New("not found")), "not found")
}
