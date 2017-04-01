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
	"reflect"
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
			Resources:   []string{"article", "user", "master"},
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
			Resources:   []string{"article", "user", "master"},
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
			if (reflect.DeepEqual(policies[0], res[0])) {
				assert.Equal(t, policies[0], res[0])
				assert.Equal(t, policies[1], res[1])
			} else {
				assert.Equal(t, policies[0], res[1])
				assert.Equal(t, policies[1], res[0])
			}

			res, err = s.FindRequestCandidates(&ladon.Request{
				Subject: "sqlamatch",
				Resource: "article",
				Action: "create",
			})
			require.Nil(t, err)
			require.Len(t, res, 1)
			assert.Equal(t, policies[0], res[0])
		})
	}
}
