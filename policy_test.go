package ladon_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/ory-am/ladon"
)

var policyConditions = []DefaultCondition{
	{Operator: "foo", Extra: map[string]interface{}{"bar": "baz"}},
}

var policyCases = []*DefaultPolicy{
	{"1", "description", []string{"user"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"create", "update"}, policyConditions},
	{"2", "description", []string{"user"}, DenyAccess, []string{"articles:<[0-9]+>"}, []string{"create", "update"}, nil},
	{"1", "description", []string{"user"}, "foobar", []string{"articles:<[0-9]+>"}, []string{"create", "update"}, policyConditions},
}

func TestHasAccess(t *testing.T) {
	assert.True(t, policyCases[0].HasAccess())
	assert.False(t, policyCases[1].HasAccess())
	assert.False(t, policyCases[2].HasAccess())
}

func TestGetters(t *testing.T) {
	for _, c := range policyCases {
		assert.Equal(t, c.ID, c.GetID())
		assert.Equal(t, c.Description, c.GetDescription())
		assert.Equal(t, c.Resources, c.GetResources())
		assert.Equal(t, c.Subjects, c.GetSubjects())
		assert.Equal(t, len(c.Conditions), len(c.GetConditions()))
		assert.Equal(t, c.Effect, c.GetEffect())
		assert.Equal(t, c.Permissions, c.GetPermissions())
		assert.Equal(t, byte('<'), c.GetStartDelimiter())
		assert.Equal(t, byte('>'), c.GetEndDelimiter())
	}
}
