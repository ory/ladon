package ladon_test

import (
	"testing"

	. "github.com/ory-am/ladon"
	"github.com/stretchr/testify/assert"
)

var policyConditions = Conditions{
	&SubjectIsOwnerCondition{},
}

var policyCases = []*DefaultPolicy{
	{
		ID:          "1",
		Description: "description",
		Subjects:    []string{"user"},
		Effect:      AllowAccess,
		Resources:   []string{"articles:<[0-9]+>"},
		Actions:     []string{"create", "update"},
		Conditions:  policyConditions,
	},
	{
		Effect: DenyAccess,
	},
}

func TestHasAccess(t *testing.T) {
	assert.True(t, policyCases[0].AllowAccess())
	assert.False(t, policyCases[1].AllowAccess())
}

func TestGetters(t *testing.T) {
	for _, c := range policyCases {
		assert.Equal(t, c.ID, c.GetID())
		assert.Equal(t, c.Description, c.GetDescription())
		assert.Equal(t, c.Resources, c.GetResources())
		assert.Equal(t, c.Subjects, c.GetSubjects())
		assert.Equal(t, len(c.Conditions), len(c.GetConditions()))
		assert.Equal(t, c.Effect, c.GetEffect())
		assert.Equal(t, c.Actions, c.GetActions())
		assert.Equal(t, byte('<'), c.GetStartDelimiter())
		assert.Equal(t, byte('>'), c.GetEndDelimiter())
	}
}
