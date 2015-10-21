package policy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var cases = []*DefaultPolicy{
	&DefaultPolicy{"1", "description", []string{"user"}, AllowAccess, []string{"articles:[0-9]+"}, []string{"create", "update"}},
	&DefaultPolicy{"2", "description", []string{"user"}, DenyAccess, []string{"articles:[0-9]+"}, []string{"create", "update"}},
	&DefaultPolicy{"1", "description", []string{"user"}, "foobar", []string{"articles:[0-9]+"}, []string{"create", "update"}},
}

func TestHasAccess(t *testing.T) {
	assert.True(t, cases[0].HasAccess())
	assert.False(t, cases[1].HasAccess())
	assert.False(t, cases[2].HasAccess())
}

func TestGetters(t *testing.T) {
	for _, c := range cases {
		assert.Equal(t, c.ID, c.GetID())
		assert.Equal(t, c.Description, c.GetDescription())
		assert.Equal(t, c.Resources, c.GetResources())
		assert.Equal(t, c.Subjects, c.GetSubjects())
	}
}
