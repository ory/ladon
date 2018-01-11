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
		matches   string
		delimiter string
		resource  string
		value     string
		pass      bool
	}{
		{delimiter: ":", value: "GroupId:2", resource: "resources:sensor-data:GroupId:2:deviceIds:a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{delimiter: "_", value: "GroupId_2", resource: "resources_sensordata_GroupId_2_deviceIds_a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{delimiter: "_", value: "deviceIds_a9b", resource: "resources_sensordata_GroupId_2_deviceIds_a9b", pass: true},
		{delimiter: "_", value: "deviceIds_a9b", resource: "resources_sensordata_GroupId_2_deviceIds_a9b6", pass: false},
		{delimiter: ":", value: "GroupId:2", resource: "resources:sensor-data:GroupId:2:deviceIds:*", pass: true},
		{delimiter: ":", value: "GroupId:2", resource: "abc", pass: false},
		{delimiter: ":", value: "", resource: "abc", pass: false},
		{delimiter: "", value: "test", resource: "abc", pass: false},
	} {
		condition := &ResourceCondition{}

		ctx := make(Context)

		request := &Request{Resource: c.resource, Subject: "users:arneanka", Context: ctx}
		resourceFilter := &ResourceFilter{Delimiter: c.delimiter, Value: c.value}

		assert.Equal(t, c.pass, condition.Fulfills(resourceFilter, request), "%s", c.matches)
		assert.Equal(t, condition.GetName(), "ResourceCondition", "should be called ResourceCondition")
	}
}
