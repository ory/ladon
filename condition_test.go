package ladon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultCondition(t *testing.T) {
	for k, c := range []DefaultCondition{
		{Operator: "op"},
		{},
		{Operator: "op", Extra: map[string]interface{}{}},
		{Operator: "op", Extra: map[string]interface{}{"foo": "bar"}},
	} {
		assert.Equal(t, c.Operator, c.GetOperator(), "Case %d", k)
		if c.Extra == nil {
			assert.NotNil(t, c.GetExtra())
		} else {
			assert.Equal(t, c.Extra, c.GetExtra())
		}
	}
}
