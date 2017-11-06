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
	"github.com/stretchr/testify/require"
)

func TestSQLManagerMigrateFromMajor0Minor6ToMajor0Minor7(t *testing.T) {
	// Setting up the Migration is easy:
	//  var db = getSqlDatabaseFromSomewhere()
	//  s := NewSQLManager(db, nil)
	//
	//  if _, err := s.CreateSchemas(); err != nil {
	//  	log.Fatalf("Could not create mysql schema: %v", err)
	//  }
	//
	//  migrator = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
	//	  DB:db,
	//	  SQLManager:s,
	//  }

	for k, s := range map[string]ladon.ManagerMigrator{
		"postgres": migrators["postgres"],
		"mysql":    migrators["mysql"],
	} {
		t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {

			// This create part is only necessary to populate the data store with some values. If you
			// migrate you won't need this
			for _, c := range ladon.TestManagerPolicies {
				t.Run(fmt.Sprintf("create=%s", k), func(t *testing.T) {
					require.NoError(t, s.Create(c))
				})
			}

			require.NoError(t, s.Migrate())

			for _, c := range ladon.TestManagerPolicies {
				t.Run(fmt.Sprintf("fetch=%s", k), func(t *testing.T) {
					get, err := s.GetManager().Get(c.GetID())
					require.NoError(t, err)

					ladon.AssertPolicyEqual(t, c, get)

					require.NoError(t, s.GetManager().Delete(c.GetID()))
				})
			}
		})
	}
}
