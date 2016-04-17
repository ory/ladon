package ladon

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	c := NewContext(&http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"127.0.0.1"},
		},
	}, "foobar")

	assert.Equal(t, "127.0.0.1", c.ClientIP)
}
