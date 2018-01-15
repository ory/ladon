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

func TestResourceContains(t *testing.T) {
	for _, c := range []struct {
		matches   string
		delimiter string
		resource  string
		value     string
		pass      bool
	}{
		//Correct matching within resource string with delimiter:
		{matches: "Find value fails incomplete resource string parts.", delimiter: ":", value: "foo:ba", resource: "foo:bar", pass: false},

		{matches: "Find value as equal to a resource string", delimiter: ":", value: "foo:bar", resource: "foo:bar", pass: true},
		{matches: "Find value as prefix of resource string", delimiter: ":", value: "foo:bar", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being prefix of resource string", delimiter: ":", value: "foo:baz", resource: "foo:bar:baz", pass: false},
		{matches: "Find value as afix of resource string", delimiter: ":", value: "bar:baz", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being afix of resource string", delimiter: ":", value: "foo:baz", resource: "foo:bar:baz", pass: false},
		{matches: "Find value as in middle of resource string", delimiter: ":", value: "bar", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being in middle of resource string", delimiter: ":", value: "bar", resource: "baz:foo:baz", pass: false},

		//Correct matching within resource string without delimiter:
		{matches: "Find value does not work for part incomplete resource string parts without delimiter.", value: "foo:ba", resource: "foo:bar", pass: true},
		{matches: "Find value as equal to a resource string", value: "foo:bar", resource: "foo:bar", pass: true},
		{matches: "Find value as prefix of resource string", value: "foo:bar", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being prefix of resource string", value: "foo:baz", resource: "foo:bar:baz", pass: false},
		{matches: "Find value as afix of resource string", value: "bar:baz", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being afix of resource string", value: "foo:baz", resource: "foo:bar:baz", pass: false},
		{matches: "Find value as in middle of resource string", value: "bar", resource: "foo:bar:baz", pass: true},
		{matches: "Find value not being in middle of resource string", value: "bar", resource: "baz:foo:baz", pass: false},

		//Erroneous requests:
		{matches: "value missing from request", delimiter: ":", value: "", resource: "abc", pass: false},
	} {
		condition := &ResourceContainsCondition{}

		ctx := make(Context)
		request := &Request{Resource: c.resource, Subject: "users:arneanka", Context: ctx}

		//The context:
		resourceFilter := make(map[string]interface{})
		if len(c.delimiter) > 0 {
			resourceFilter["delimiter"] = c.delimiter
		}
		resourceFilter["value"] = c.value

		assert.Equal(t, c.pass, condition.Fulfills(resourceFilter, request), "%s", c.matches)
		assert.Equal(t, condition.GetName(), "ResourceContainsCondition", "should be called ResourceContainsCondition")
	}
}
