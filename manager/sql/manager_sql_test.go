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

package sql_test

import (
	"flag"
	"fmt"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/ladon"
	"github.com/ory/ladon/manager/sql"
	"github.com/ory/sqlcon/dockertest"
	"github.com/stretchr/testify/require"
	"sync"
)

type testDB struct {
	m  *sql.SQLManager
	db *sqlx.DB
}

type testQuery struct {
	checkRegexState string
}

var queries = map[string]testQuery{
	"postgres": {
		checkRegexState: "SELECT COUNT(m.id) FROM %v m JOIN %v r ON m.id = r.%v WHERE r.policy = $1 AND m.has_regex != $2",
	},
	"mysql": {
		checkRegexState: "SELECT COUNT(m.id) FROM %v m JOIN %v r ON m.id = r.%v WHERE r.policy = ? AND m.has_regex != ?",
	},
}
var policyRelations = []string{"action", "resource", "subject"}
var managers = map[string]testDB{}
var managersMutex = &sync.Mutex{}

func TestMain(m *testing.M) {
	runner := dockertest.Register()

	flag.Parse()
	if !testing.Short() {
		dockertest.Parallel([]func(){
			func() {
				db, err := dockertest.ConnectToTestPostgreSQL()
				if err != nil {
					log.Fatalf("Could not connect to PostgresSQL database: %v", err)
					return
				}
				s := sql.NewSQLManager(db, nil)
				s.CreateSchemas("", "")
				managersMutex.Lock()
				managers["postgres"] = testDB{m: s, db: db}
				managersMutex.Unlock()
			},
			func() {
				db, err := dockertest.ConnectToTestMySQL()
				if err != nil {
					log.Fatalf("Could not connect to MySQL database: %v", err)
					return
				}
				s := sql.NewSQLManager(db, nil)
				s.CreateSchemas("", "")
				managersMutex.Lock()
				managers["mysql"] = testDB{m: s, db: db}
				managersMutex.Unlock()
			},
		})
	}

	runner.Exit(m.Run())
}

func TestCreateExplicitPolicy(t *testing.T) {
	for k, v := range managers {
		id := "test-explicit"
		p := &ladon.DefaultPolicy{
			Actions:     []string{"view", "edit"},
			Description: "An explicit (non-regex) test policy",
			Effect:      ladon.AllowAccess,
			ID:          id,
			Resources:   []string{"blog:post:1"},
			Subjects:    []string{"otto"},
		}
		err := v.m.Create(p)
		require.NoError(t, err)
		expectRegexStateForAllRelations(t, k, v, id, false)
	}
}

func TestCreateRegexPolicy(t *testing.T) {
	for k, v := range managers {
		id := "test-regex"
		p := &ladon.DefaultPolicy{
			Actions:     []string{"<(view|edit)>"},
			Description: "An regex test policy",
			Effect:      ladon.AllowAccess,
			ID:          id,
			Resources:   []string{"blog:post:<.*>"},
			Subjects:    []string{"<(otto|hugo)>"},
		}
		err := v.m.Create(p)
		require.NoError(t, err)
		expectRegexStateForAllRelations(t, k, v, id, true)
	}
}

func expectRegexStateForAllRelations(t *testing.T, dbType string, tdb testDB, policyID string, regexState bool) {
	for _, v := range policyRelations {
		expectRegexState(t, dbType, tdb, policyID, v, regexState)
	}
}

func expectRegexState(t *testing.T, dbType string, tdb testDB, policyID string, rel string, regexState bool) {
	tbl := fmt.Sprintf("ladon_%v", rel)
	relTbl := fmt.Sprintf("ladon_policy_%v_rel", rel)
	query := fmt.Sprintf(queries[dbType].checkRegexState, tbl, relTbl, rel)
	var count int
	row := tdb.db.QueryRow(query, policyID, regexState)
	require.NoError(t, row.Scan(&count))
	require.Equal(t, 0, count, "Expected has_regex to be %v but got %v entries with a different state", regexState, count)
}
