package ladon_test

import (
	"testing"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/ory-am/ladon"
	"github.com/pborman/uuid"
	"fmt"
)

// This test is skipped because the method was deprecated
//
func TestFindPoliciesForSubject(t *testing.T) {
	policies := []*ladon.DefaultPolicy{
		{
			ID:          uuid.New(),
			Description: "description",
			Subjects:    []string{"sql<.*>match"},
			Effect:      ladon.AllowAccess,
			Resources:   []string{"master", "user", "article"},
			Actions:     []string{"create", "update", "delete"},
			Conditions:  ladon.Conditions{
				"foo": &ladon.StringEqualCondition{
					Equals: "foo",
				},
			},
		},
		{
			ID:          uuid.New(),
			Description: "description",
			Subjects:    []string{"sqlmatch"},
			Effect:      ladon.AllowAccess,
			Resources:   []string{"master", "user", "article"},
			Actions:     []string{"create", "update", "delete"},
			Conditions:  ladon.Conditions{
				"foo": &ladon.StringEqualCondition{
					Equals: "foo",
				},
			},
		},
	}

	for k, s := range map[string]ladon.Manager{
		"postgres": managers["postgres"],
		"mysql": managers["mysql"],
	} {
		t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {
			for _, c := range policies {
				require.Nil(t, s.Create(c))
			}

			res, err := s.FindRequestCandidates(&ladon.Request{
				Subject: "sqlmatch",
				Resource: "article",
				Action: "create",
			})
			require.Nil(t, err)
			require.Len(t, res, 2)
			if policies[0].ID == res[0].GetID() {
				assertEqual(t, policies[0], res[0])
				assertEqual(t, policies[1], res[1])
			} else {
				assertEqual(t, policies[0], res[1])
				assertEqual(t, policies[1], res[0])
			}

			res, err = s.FindRequestCandidates(&ladon.Request{
				Subject: "sqlamatch",
				Resource: "article",
				Action: "create",
			})
			require.Nil(t, err)
			require.Len(t, res, 1)
			assertEqual(t, policies[0], res[0])
		})
	}
}

func assertEqual(t *testing.T, a, b  ladon.Policy) {
	assert.Equal(t, a.GetID(), b.GetID())
	assert.Equal(t, a.GetDescription(), b.GetDescription())
	assert.Equal(t, a.GetEffect(), b.GetEffect())
	assert.True(t, testEq(a.GetActions(), b.GetActions()))
	assert.True(t, testEq(a.GetResources(), b.GetResources()))
	assert.True(t, testEq(a.GetSubjects(), b.GetSubjects()))
	assert.EqualValues(t, a.GetConditions(), b.GetConditions())
}
func testEq(a, b []string) bool {

	if a == nil && b == nil {
		return true;
	}

	if a == nil || b == nil {
		return false;
	}

	if len(a) != len(b) {
		return false
	}

	var found bool
	for i := range a {
		found = false

		for y := range b {
			if a[i] == b[y] {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
