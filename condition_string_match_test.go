package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringMatch(t *testing.T) {
	for _, c := range []struct {
		matches string
		value   interface{}
		pass    bool
	}{
		{matches: ".*", value: "abc", pass: true},
		{matches: "abc.*", value: "abc", pass: true},
		{matches: "abc.+", value: "abc", pass: false},
	} {
		condition := &StringMatchCondition{
			Matches: c.matches,
		}

		assert.Equal(t, c.pass, condition.Fulfills(c.value, new(Request)), "%s", c.matches)
	}
}
