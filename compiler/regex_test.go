/*
 * Copyright © 2016-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright 	2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

package compiler

import (
	"testing"

	"github.com/dlclark/regexp2"
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

		{`urn:foo:<user=(?!admin).*>`, '<', '>', false, "urn:foo:user=john", false},
		{`urn:foo:<user=(?!admin).*>`, '<', '>', false, "urn:foo:user=admin", true},

		{`urn:foo:user=<[[:digit:]]*>`, '<', '>', false, "urn:foo:user=admin", true},
		{`urn:foo:user=<[[:digit:]]*>`, '<', '>', false, "urn:foo:user=62235", false},

		{`urn:foo:user={(?P<id>\d{3})}`, '{', '}', false, "urn:foo:user=622", false},
		{`urn:foo:user=<(?P<id>\d{3})>`, '<', '>', false, "urn:foo:user=622", false},
		{`urn:foo:user=<(?P<id>\d{3})>`, '<', '>', false, "urn:foo:user=aaa", true},

		// Ignoring this case for now...
		// {"urn:foo.bar.com:{.*\\{}", '{', '}', false, "", true},
	} {
		k++
		result, err := CompileRegex(c.template, c.delimiterStart, c.delimiterEnd)
		assert.Equal(t, c.failCompile, err != nil, "Case %d", k)
		if c.failCompile || err != nil {
			continue
		}

		t.Logf("Case %d compiled to: %s", k, result.String())
		re := regexp2.MustCompile(result.String(), regexp2.RE2)
		ok, err := re.MatchString(c.matchAgainst)
		assert.Nil(t, err, "Case %d", k)
		assert.Equal(t, !c.failMatch, ok, "Case %d", k)

	}
}
