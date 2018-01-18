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
	"log"
	"os"
	"sync"
	"testing"

	. "github.com/ory/ladon"
	"github.com/ory/ladon/integration"
	. "github.com/ory/ladon/manager/memory"
	. "github.com/ory/ladon/manager/sql"
	"github.com/stretchr/testify/require"
)

var managers = map[string]Manager{}
var migrators = map[string]ManagerMigrator{}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(3)
	connectMySQL(&wg)
	connectPG(&wg)
	connectMEM(&wg)
	wg.Wait()

	s := m.Run()
	integration.KillAll()
	os.Exit(s)
}

func connectMEM(wg *sync.WaitGroup) {
	defer wg.Done()
	managers["memory"] = NewMemoryManager()
}

func connectPG(wg *sync.WaitGroup) {
	defer wg.Done()
	var db = integration.ConnectToPostgres("ladon")
	s := NewSQLManager(db, nil)
	if _, err := s.CreateSchemas("", ""); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
	migrators["postgres"] = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
		DB:         db,
		SQLManager: s,
	}
}

func connectMySQL(wg *sync.WaitGroup) {
	defer wg.Done()
	var db = integration.ConnectToMySQL()
	s := NewSQLManager(db, nil)
	if _, err := s.CreateSchemas("", ""); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
	migrators["mysql"] = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
		DB:         db,
		SQLManager: s,
	}
}

func TestManagers(t *testing.T) {
	t.Run("type=get errors", func(t *testing.T) {
		for k, s := range managers {
			t.Run("manager="+k, TestHelperGetErrors(s))
		}
	})

	t.Run("type=CRUD", func(t *testing.T) {
		for k, s := range managers {
			t.Run(fmt.Sprintf("manager=%s", k), TestHelperCreateGetDelete(s))
		}
	})

	t.Run("type=find", func(t *testing.T) {
		for k, s := range map[string]Manager{
			"postgres": managers["postgres"],
			"mysql":    managers["mysql"],
		} {
			t.Run(fmt.Sprintf("manager=%s", k), TestHelperFindPoliciesForSubject(k, s))
		}
	})

	t.Run("type=migrate 6 to 7", func(t *testing.T) {
		for k, s := range map[string]ManagerMigrator{
			"postgres": migrators["postgres"],
			"mysql":    migrators["mysql"],
		} {
			t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {

				// This create part is only necessary to populate the data store with some values. If you
				// migrate you won't need this
				for _, c := range TestManagerPolicies {
					t.Run(fmt.Sprintf("create=%s", k), func(t *testing.T) {
						require.NoError(t, s.Create(c))
					})
				}

				require.NoError(t, s.Migrate())

				for _, c := range TestManagerPolicies {
					t.Run(fmt.Sprintf("fetch=%s", k), func(t *testing.T) {
						get, err := s.GetManager().Get(c.GetID())
						require.NoError(t, err)

						AssertPolicyEqual(t, c, get)

						require.NoError(t, s.GetManager().Delete(c.GetID()))
					})
				}
			})
		}
	})
}
