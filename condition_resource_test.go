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

func TestResourceMatch(t *testing.T) {
	for _, c := range []struct {
		matches        string
		resource       string
		resourceRule   string
		resourceFilter string
		pass           bool
	}{
		{resourceFilter: "tenantIds:2", resource: "resources:sensor-data:tenantIds:2:deviceIds:a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{resourceFilter: "tenantIds:2", resource: "resources:sensor-data:tenantIds:2:deviceIds:*", pass: true},
		{resourceFilter: "tenantIds:2", resource: "abc", pass: false},
	} {
		condition := &ResourceCondition{
		//Matches: c.matches,
		}

		ctx := make(Context)
		ctx["resourceFilter"] = c.resourceFilter

		request := &Request{Resource: c.resource, Subject: "users:arneanka", Context: ctx}

		assert.Equal(t, c.pass, condition.Fulfills(c.resourceFilter, request), "%s", c.matches)
	}
}
