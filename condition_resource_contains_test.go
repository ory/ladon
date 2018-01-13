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
		//Correct matching within resource string without delimiter:
		{matches: "Exact resource component match in resource string, with no delimiter", value: ":GroupId:2:", resource: "resources:sensor-data:GroupId:2:deviceIds:a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{matches: "Exact resource component match in resource string, with no delimiter", value: "myResource", resource: "myResource", pass: true},
		//Better to use regex for string, to avoid false positives:
		{matches: "Exact resource component match in resource string, with no delimiter", value: "(notMyResource$)", resource: "myResource", pass: false},
		{matches: "Exact resource component match with resource string, with no delimiter", value: "(^myResource$)", resource: "notMyResource", pass: false},
		{matches: "Exact resource component match with resource string, with no delimiter", value: "(^myResource$)", resource: "MyResourceNot", pass: false},
		{matches: "Exact resource component match with resource string, with no delimiter", value: "(^myResource$)", resource: "myResource", pass: true},
	
		//Correct matching within resource string:
		{matches: "Exact resource component match in resource string, with : delimiter", delimiter: ":", value: "GroupId:2", resource: "resources:sensor-data:GroupId:2:deviceIds:a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{matches: "Not exact resource component match in resource string, with : delimiter", delimiter: "_", value: "GroupId:2", resource: "resources:sensordata:GroupId:22:deviceIds:a9b6", pass: false},
		{matches: "Exact resource component match in resource string, with _ delimiter", delimiter: "_", value: "GroupId_2", resource: "resources_sensordata_GroupId_2_deviceIds_a9b576e8-7419-4eed-a010-7f68ec0ff588", pass: true},
		{matches: "Not exact resource component match in resource string, with : delimiter", delimiter: ":", value: "roupId:22", resource: "resources:sensordata:GroupId:22:deviceIds:a9b6", pass: false},

		//Correct matching at the end of the resource string:
		{matches: "Exact resource component match at end of resource string, with _ delimiter", delimiter: "_", value: "deviceIds_a9b", resource: "resources_sensordata_GroupId_2_deviceIds_a9b", pass: true},
		{matches: "Not exact resource component match at end of resource string, with _ delimiter", delimiter: "_", value: "deviceIds_a9b", resource: "resources_sensordata_GroupId_2_deviceIds_a9b6", pass: false},
		{matches: "Not found in resource string", delimiter: ":", value: "GroupId:2", resource: "abc", pass: false},
		//More advanced regex:
		{matches: "Regex resource component match at end of resource string, with _ delimiter", delimiter: "_", value: "deviceIds_a9.*", resource: "resources_sensordata_GroupId_2_deviceIds_a9b", pass: true},
		{matches: "Regex resource component match at end of resource string, with : delimiter", delimiter: ":", value: "deviceIds:*", resource: "resources:sensordata:GroupId:2:deviceIds:a9b", pass: true},

		//Erroneous requests:
		{matches: "value missing from request", delimiter: ":", value: "", resource: "abc", pass: false},
		{matches: "delimiter missing from request", delimiter: "", value: "test", resource: "abc", pass: false},
	} {
		condition := &ResourceContainsCondition{}

		ctx := make(Context)
		request := &Request{Resource: c.resource, Subject: "users:arneanka", Context: ctx}

		//The context:
		resourceFilter := make(map[string]interface{})
		if(len(c.delimiter) > 0){
			resourceFilter["delimiter"] = c.delimiter
		}
		resourceFilter["value"] = c.value

		assert.Equal(t, c.pass, condition.Fulfills(resourceFilter, request), "%s", c.matches)
		assert.Equal(t, condition.GetName(), "ResourceContainsCondition", "should be called ResourceContainsCondition")
	}
}
