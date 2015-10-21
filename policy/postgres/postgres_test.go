package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
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

var db *sql.DB
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
	c, ip, port, err := dockertest.SetupPostgreSQLContainer(time.Second * 5)
	if err != nil {
		log.Fatalf("Could not set up PostgreSQL container: %v", err)
	}
	defer c.KillRemove()

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable", dockertest.PostgresUsername, dockertest.PostgresPassword, ip, port)
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("Could not set up PostgreSQL container: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

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
