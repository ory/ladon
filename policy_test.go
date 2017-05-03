package ladon_test

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	for k, c := range policyCases {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			var cc = DefaultPolicy{
				Conditions: make(Conditions),
			}
			data, err := json.Marshal(c)
			RequireError(t, false, err)

			json.Unmarshal(data, &cc)
			RequireError(t, false, err)
			assert.Equal(t, c, &cc)
		})
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

func RequireError(t *testing.T, expectError bool, err error, args ...interface{}) {
	if err != nil && !expectError {
		t.Logf("Unexpected error: %s\n", err.Error())
		t.Logf("Arguments: %v\n", args)
		t.Logf("\n\n")
	}
	require.Equal(t, expectError, err != nil)
}
