// Copyright © 2017 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
