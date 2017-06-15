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
