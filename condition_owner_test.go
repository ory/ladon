package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectIsOwnerCondition(t *testing.T) {
	c := &SubjectIsOwnerCondition{}
	assert.True(t, c.Fulfills(&Request{
		Subject: "foo",
		Context: &Context{
			Owner: "foo",
		},
	}))

	assert.False(t, c.Fulfills(&Request{
		Subject: "bar",
		Context: &Context{
			Owner: "foo",
		},
	}))
}
func TestSubjectIsNotOwnerCondition(t *testing.T) {
	c := &SubjectIsNotOwnerCondition{}
	assert.False(t, c.Fulfills(&Request{
		Subject: "foo",
		Context: &Context{
			Owner: "foo",
		},
	}))

	assert.True(t, c.Fulfills(&Request{
		Subject: "bar",
		Context: &Context{
			Owner: "foo",
		},
	}))
}
