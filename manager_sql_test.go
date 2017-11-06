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

package ladon_test

import (
	"fmt"
	"testing"

	"github.com/ory/ladon"
)

// This test is skipped because the method was deprecated
//
func TestFindPoliciesForSubject(t *testing.T) {

	for k, s := range map[string]ladon.Manager{
		"postgres": managers["postgres"],
		"mysql":    managers["mysql"],
	} {
		t.Run(fmt.Sprintf("manager=%s", k), ladon.TestHelperFindPoliciesForSubject(k, s))
	}
}
