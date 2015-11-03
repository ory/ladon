package postgres

import (
	"github.com/ory-am/dockertest"
	"github.com/ory-am/ladon/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

var s *Store

var conditions = []policy.Condition{
	&policy.DefaultCondition{Operator: "foo", Extra: map[string]interface{}{"bar": "baz"}},
}

var cases = []*policy.DefaultPolicy{
	{"1", "description", []string{"user", "anonymous"}, policy.AllowAccess, []string{"article", "user"}, []string{"create", "update"}, conditions},
	{"2", "description", []string{}, policy.AllowAccess, []string{"article|user"}, []string{"view"}, nil},
	{"3", "description", []string{"peter|max"}, policy.DenyAccess, []string{"article", "user"}, []string{"view"}, conditions},
	{"4", "description", []string{"user|max|anonymous", "peter"}, policy.DenyAccess, []string{".*"}, []string{"disable"}, conditions},
	{"5", "description", []string{".*"}, policy.AllowAccess, []string{"article|user"}, []string{"view"}, conditions},
	{"6", "description", []string{"us[er]+"}, policy.AllowAccess, []string{"article|user"}, []string{"view"}, conditions},
}

func TestMain(m *testing.M) {
	c, db, err := dockertest.OpenPostgreSQLContainerConnection(15, time.Millisecond*500)
	if err != nil {
		log.Fatalf("Could not set up PostgreSQL container: %v", err)
	}
	defer c.KillRemove()

	s = New(db)
	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	os.Exit(m.Run())
}

func TestCreateGetDelete(t *testing.T) {
	for _, c := range cases {
		err := s.Create(c)
		assert.Nil(t, err)

		get, err := s.Get(c.GetID())
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(c, get), "%v does not equal %v", c, get)
	}

	for _, c := range cases {
		assert.Nil(t, s.Delete(c.GetID()))
		_, err := s.Get(c.GetID())
		assert.NotNil(t, err)
	}
}

func TestFindPoliciesForSubject(t *testing.T) {
	for _, c := range cases {
		require.Nil(t, s.Create(c))
	}

	policies, err := s.FindPoliciesForSubject("user")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(policies))

	policies, err = s.FindPoliciesForSubject("foobar")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(policies))
}
