package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"encoding/json"
	"github.com/stretchr/testify/require"
)

func TestConditionsAppend(t *testing.T) {
	cs := Conditions{}
	c := &SubjectIsOwnerCondition{}
	cs.AddCondition(c)
	assert.Equal(t, c, cs[0])
}

func TestMarshalUnmarshal(t *testing.T) {
	css := &Conditions{
		&SubjectIsOwnerCondition{},
		&CIDRCondition{CIDR:"127.0.0.1/0"},
	}
	out, err := json.Marshal(css)
	require.Nil(t, err)
	t.Logf("%s", out)


	cs := Conditions{}
	require.Nil(t, json.Unmarshal([]byte(`[
	{
		"name": "SubjectIsOwnerCondition",
		"options": {}
	},
	{
		"name": "CIDRCondition",
		"options": {
			"cidr": "127.0.0.1/0"
		}
	}
]`), &cs))

	require.Len(t, cs, 2)
	assert.IsType(t, &SubjectIsOwnerCondition{}, cs[0])
	assert.IsType(t, &CIDRCondition{}, cs[1])
}
