package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPairsEqualMatch(t *testing.T) {
	for _, c := range []struct {
		pairs interface{}
		pass  bool
	}{
		{pairs: "junk", pass: false},
		{pairs: []interface{}{[]interface{}{}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "1", "2"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "2"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "1"}, []interface{}{"2", "3"}}, pass: false},
		{pairs: []interface{}{}, pass: true},
		{pairs: []interface{}{[]interface{}{"1", "1"}}, pass: true},
		{pairs: []interface{}{[]interface{}{"1", "1"}, []interface{}{"2", "2"}}, pass: true},
	} {
		condition := &StringPairsEqualCondition{}

		assert.Equal(t, c.pass, condition.Fulfills(c.pairs, new(Request)), "%s", c.pairs)
	}
}
