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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditionsAppend(t *testing.T) {
	cs := Conditions{}
	c := &CIDRCondition{}
	cs.AddCondition("clientIP", c)
	assert.Equal(t, c, cs["clientIP"])
}

func TestMarshalUnmarshalNative(t *testing.T) {
	css := &Conditions{
		"clientIP": &CIDRCondition{CIDR: "127.0.0.1/0"},
		"owner":    &EqualsSubjectCondition{},
	}
	out, err := json.Marshal(css)
	require.Nil(t, err)
	t.Logf("%s", out)

	cs := Conditions{}
	require.Nil(t, cs.UnmarshalJSON(out))
}

func TestMarshalUnmarshal(t *testing.T) {
	css := &Conditions{
		"clientIP": &CIDRCondition{CIDR: "127.0.0.1/0"},
		"owner":    &EqualsSubjectCondition{},
		"role":     &StringMatchCondition{Matches: ".*"},
	}
	out, err := json.Marshal(css)
	require.Nil(t, err)
	t.Logf("%s", out)

	cs := Conditions{}
	require.Nil(t, json.Unmarshal([]byte(`{
	"owner": {
		"type": "EqualsSubjectCondition"
	},
	"clientIP": {
		"type": "CIDRCondition",
		"options": {
			"cidr": "127.0.0.1/0"
		}
	},
	"role": {
		"type": "StringMatchCondition",
		"options": {
			"matches": ".*"
		}
	},
	"resourceFilter": {
			"type": "ResourceContainsCondition"
		}
}`), &cs))

	require.Len(t, cs, 4)
	assert.IsType(t, &EqualsSubjectCondition{}, cs["owner"])
	assert.IsType(t, &CIDRCondition{}, cs["clientIP"])
	assert.IsType(t, &StringMatchCondition{}, cs["role"])
	assert.IsType(t, &ResourceContainsCondition{}, cs["resourceFilter"])
}

func TestUnmarshalFails(t *testing.T) {
	cs := Conditions{}
	require.NotNil(t, json.Unmarshal([]byte(`{
	"somefield": {
		"type": "DoesntExist"
	}
}`), &cs))
}
