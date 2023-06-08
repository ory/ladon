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

package ladon_test

import (
	"fmt"
	"testing"

	. "github.com/ory/ladon"
	. "github.com/ory/ladon/manager/memory"
)

var managers = map[string]Manager{}

func TestMain(m *testing.M) {
	connectMEM()
}

func connectMEM() {
	managers["memory"] = NewMemoryManager()
}

func TestManagers(t *testing.T) {
	t.Run("type=get errors", func(t *testing.T) {
		for k, s := range managers {
			t.Run("manager="+k, HelperTestGetErrors(s))
		}
	})

	t.Run("type=CRUD", func(t *testing.T) {
		for k, s := range managers {
			t.Run(fmt.Sprintf("manager=%s", k), HelperTestCreateGetDelete(s))
		}
	})

	t.Run("type=find", func(t *testing.T) {
		for k, s := range map[string]Manager{
			"postgres": managers["postgres"],
			"mysql":    managers["mysql"],
		} {
			t.Run(fmt.Sprintf("manager=%s", k), HelperTestFindPoliciesForSubject(k, s))
			t.Run(fmt.Sprintf("manager=%s", k), HelperTestFindPoliciesForResource(k, s))
		}
	})
}
