package ladon_test

import (
	"testing"

	"encoding/json"

	"github.com/ory-am/hydra/pkg"
	. "github.com/ory-am/ladon"
	"github.com/stretchr/testify/assert"
)

var policyConditions = Conditions{
	"owner": &EqualsSubjectCondition{},
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
		Effect:     DenyAccess,
		Conditions: make(Conditions),
	},
}

func TestHasAccess(t *testing.T) {
	assert.True(t, policyCases[0].AllowAccess())
	assert.False(t, policyCases[1].AllowAccess())
}

func TestMarshalling(t *testing.T) {
	for _, c := range policyCases {
		var cc = DefaultPolicy{
			Conditions: make(Conditions),
		}
		data, err := json.Marshal(c)
		pkg.RequireError(t, false, err)

		t.Logf("Got data: %s\n", data)
		json.Unmarshal(data, &cc)
		pkg.RequireError(t, false, err)
		assert.Equal(t, c, &cc)
	}
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
