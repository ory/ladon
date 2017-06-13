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
