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

var cases = []*policy.DefaultPolicy{
	&policy.DefaultPolicy{"1", "description", []string{"user", "anonymous"}, policy.AllowAccess, []string{"article", "user"}, []string{"create", "update"}},
	&policy.DefaultPolicy{"2", "description", []string{}, policy.AllowAccess, []string{"article|user"}, []string{"view"}},
	&policy.DefaultPolicy{"3", "description", []string{"peter|max"}, policy.DenyAccess, []string{"article", "user"}, []string{"view"}},
	&policy.DefaultPolicy{"4", "description", []string{"user|max|anonymous", "peter"}, policy.DenyAccess, []string{".*"}, []string{"disable"}},
	&policy.DefaultPolicy{"5", "description", []string{".*"}, policy.AllowAccess, []string{"article|user"}, []string{"view"}},
	&policy.DefaultPolicy{"6", "description", []string{"us[er]+"}, policy.AllowAccess, []string{"article|user"}, []string{"view"}},
}

func TestMain(m *testing.M) {
	c, db, err := dockertest.OpenPostgreSQLContainerConnection(15, time.Millisecond*500)
	if err != nil {
		log.Fatalf("Could not set up PostgreSQL container: %v", err)
	}
	defer c.KillRemove()

	s = &Store{db}
	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	os.Exit(m.Run())
}

func TestCreateGetDelete(t *testing.T) {
	for _, c := range cases {
		create, err := s.Create(c.GetID(), c.GetDescription(), c.GetEffect(), c.GetSubjects(), c.GetPermissions(), c.GetResources())
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(c, create), "%v does not equal %v", c, create)

		get, err := s.Get(c.GetID())
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(c, create), "%v does not equal %v", c, get)
	}

	for _, c := range cases {
		assert.Nil(t, s.Delete(c.GetID()))
		_, err := s.Get(c.GetID())
		assert.NotNil(t, err)
	}
}

func TestFindPoliciesForSubject(t *testing.T) {
	for _, c := range cases {
		_, err := s.Create(c.GetID(), c.GetDescription(), c.GetEffect(), c.GetSubjects(), c.GetPermissions(), c.GetResources())
		require.Nil(t, err)
	}

	policies, err := s.FindPoliciesForSubject("user")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(policies))

	policies, err = s.FindPoliciesForSubject("foobar")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(policies))
}
