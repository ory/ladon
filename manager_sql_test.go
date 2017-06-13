package ladon_test

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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
