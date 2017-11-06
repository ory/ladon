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

package compiler

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexCompiler(t *testing.T) {
	for k, c := range []struct {
		template       string
		delimiterStart byte
		delimiterEnd   byte
		failCompile    bool
		matchAgainst   string
		failMatch      bool
	}{
		{"urn:foo:{.*}", '{', '}', false, "urn:foo:bar:baz", false},
		{"urn:foo.bar.com:{.*}", '{', '}', false, "urn:foo.bar.com:bar:baz", false},
		{"urn:foo.bar.com:{.*}", '{', '}', false, "urn:foo.com:bar:baz", true},
		{"urn:foo.bar.com:{.*}", '{', '}', false, "foobar", true},
		{"urn:foo.bar.com:{.{1,2}}", '{', '}', false, "urn:foo.bar.com:aa", false},

		{"urn:foo.bar.com:{.*{}", '{', '}', true, "", true},
		{"urn:foo:<.*>", '<', '>', false, "urn:foo:bar:baz", false},

		// Ignoring this case for now...
		//{"urn:foo.bar.com:{.*\\{}", '{', '}', false, "", true},
	} {
		k++
		result, err := CompileRegex(c.template, c.delimiterStart, c.delimiterEnd)
		assert.Equal(t, c.failCompile, err != nil, "Case %d", k)
		if c.failCompile || err != nil {
			continue
		}

		t.Logf("Case %d compiled to: %s", k, result.String())
		ok, err := regexp.MatchString(result.String(), c.matchAgainst)
		assert.Nil(t, err, "Case %d", k)
		assert.Equal(t, !c.failMatch, ok, "Case %d", k)
	}
}
