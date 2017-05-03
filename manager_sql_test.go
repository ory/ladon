package ladon_test

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/ory/ladon"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			Conditions: ladon.Conditions{
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
			Conditions: ladon.Conditions{
				"foo": &ladon.StringEqualCondition{
					Equals: "foo",
				},
			},
		},
		{
			ID:          uuid.New(),
			Description: "description",
			Subjects:    []string{},
			Effect:      ladon.AllowAccess,
			Resources:   []string{"master", "user", "article"},
			Actions:     []string{"create", "update", "delete"},
			Conditions: ladon.Conditions{
				"foo": &ladon.StringEqualCondition{
					Equals: "foo",
				},
			},
		},
		{
			ID:          uuid.New(),
			Description: "description",
			Effect:      ladon.AllowAccess,
			Resources:   []string{"master", "user", "article"},
			Actions:     []string{"create", "update", "delete"},
			Conditions: ladon.Conditions{
				"foo": &ladon.StringEqualCondition{
					Equals: "foo",
				},
			},
		},
	}

	for k, s := range map[string]ladon.Manager{
		"postgres": managers["postgres"],
		"mysql":    managers["mysql"],
	} {
		t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {
			for _, c := range policies {
				t.Run(fmt.Sprintf("create=%s", k), func(t *testing.T) {
					require.NoError(t, s.Create(c))
				})
			}

			res, err := s.FindRequestCandidates(&ladon.Request{
				Subject:  "sqlmatch",
				Resource: "article",
				Action:   "create",
			})
			require.NoError(t, err)
			require.Len(t, res, 2)

			if policies[0].ID == res[0].GetID() {
				assertPolicyEqual(t, policies[0], res[0])
				assertPolicyEqual(t, policies[1], res[1])
			} else {
				assertPolicyEqual(t, policies[0], res[1])
				assertPolicyEqual(t, policies[1], res[0])
			}

			res, err = s.FindRequestCandidates(&ladon.Request{
				Subject:  "sqlamatch",
				Resource: "article",
				Action:   "create",
			})

			require.NoError(t, err)
			require.Len(t, res, 1)
			assertPolicyEqual(t, policies[0], res[0])
		})
	}
}

func assertPolicyEqual(t *testing.T, expected, got ladon.Policy) {
	assert.Equal(t, expected.GetID(), got.GetID())
	assert.Equal(t, expected.GetDescription(), got.GetDescription())
	assert.Equal(t, expected.GetEffect(), got.GetEffect())

	// This won't work in the memory manager
	//assert.NotNil(t, got.GetActions())
	//assert.NotNil(t, got.GetResources())
	//assert.NotNil(t, got.GetSubjects())

	assert.NoError(t, testEq(expected.GetActions(), got.GetActions()))
	assert.NoError(t, testEq(expected.GetResources(), got.GetResources()))
	assert.NoError(t, testEq(expected.GetSubjects(), got.GetSubjects()))
	assert.EqualValues(t, expected.GetConditions(), got.GetConditions())
}

func testEq(a, b []string) error {

	// We don't care about nil types
	//if a == nil && b == nil {
	//	return true
	//}
	//
	//if a == nil || b == nil {
	//	return false
	//}

	if len(a) != len(b) {
		return errors.Errorf("Length not equal: %v (%d) != %v (%d)", a, len(a), b, len(b))
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
			return errors.Errorf("No match found: %s from %v in %v", i, a, b)
		}
	}

	return nil
}
