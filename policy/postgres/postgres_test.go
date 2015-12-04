package postgres

import (
	"github.com/ory-am/common/pkg"
	"github.com/ory-am/dockertest"
	"github.com/ory-am/ladon/policy"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
	"time"
)

var s *Store

var conditions = []policy.DefaultCondition{
	{Operator: "foo", Extra: map[string]interface{}{"bar": "baz"}},
	{Operator: "bar", Extra: map[string]interface{}{"foo": "baz"}},
}

var cases = []*policy.DefaultPolicy{
	{uuid.New(), "description", []string{"user", "anonymous"}, policy.AllowAccess, []string{"article", "user"}, []string{"create", "update"}, conditions},
	{uuid.New(), "description", []string{}, policy.AllowAccess, []string{"<article|user>"}, []string{"view"}, nil},
	{uuid.New(), "description", []string{"<peter|max>"}, policy.DenyAccess, []string{"article", "user"}, []string{"view"}, conditions},
	{uuid.New(), "description", []string{"<user|max|anonymous>", "peter"}, policy.DenyAccess, []string{".*"}, []string{"disable"}, conditions},
	{uuid.New(), "description", []string{"<.*>"}, policy.AllowAccess, []string{"<article|user>"}, []string{"view"}, conditions},
	{uuid.New(), "description", []string{"<us[er]+>"}, policy.AllowAccess, []string{"<article|user>"}, []string{"view"}, conditions},
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

func TestCreateErrors(t *testing.T) {
	assert.NotNil(t, &policy.DefaultPolicy{ID: "invalid-format"})
}

func TestGetErrors(t *testing.T) {
	_, err := s.Get(uuid.New())
	assert.Equal(t, pkg.ErrNotFound, err)
	_, err = s.Get("asdf")
	assert.NotNil(t, err)
}

func TestCreateGetDelete(t *testing.T) {
	for _, c := range cases {
		err := s.Create(c)
		assert.Nil(t, err)

		get, err := s.Get(c.GetID())
		assert.Nil(t, err)
		pkg.AssertObjectKeysEqual(t, c, get, "Description", "Subjects", "Resources", "Effect", "Permissions")
		assert.Equal(t, len(c.Conditions), len(get.GetConditions()))
		//assert.True(t, reflect.DeepEqual(c, get), "%v does not equal %v", c, get)
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
