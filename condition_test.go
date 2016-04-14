package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConditionsAppend(t *testing.T) {
	cs := Conditions{}
	c := &SubjectIsOwnerCondition{}
	cs.AddCondition(c)
	assert.Equal(t, c, cs[0])
}
