package rethinkdb

import (
	"time"

	rdb "github.com/dancannon/gorethink"

	"log"
	"os"
	"testing"

	"github.com/ory-am/common/pkg"
	"github.com/ory-am/ladon"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ory-am/dockertest.v2"
)

var session *rdb.Session
var s *Manager

var cases = []*ladon.DefaultPolicy{
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"user", "anonymous"},
		Effect:      ladon.AllowAccess,
		Resources:   []string{"article", "user"},
		Actions:     []string{"create", "update"},
		Conditions:  ladon.Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{},
		Effect:      ladon.AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions:  ladon.Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<peter|max>"},
		Effect:      ladon.DenyAccess,
		Resources:   []string{"article", "user"},
		Actions:     []string{"view"},
		Conditions: ladon.Conditions{
			&ladon.CIDRCondition{
				CIDR: "1234",
			},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<user|max|anonymous>", "peter"},
		Effect:      ladon.DenyAccess,
		Resources:   []string{".*"},
		Actions:     []string{"disable"},
		Conditions: ladon.Conditions{
			&ladon.CIDRCondition{
				CIDR: "1234",
			},
			&ladon.SubjectIsNotOwnerCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<.*>"},
		Effect:      ladon.AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: ladon.Conditions{
			&ladon.CIDRCondition{
				CIDR: "1234",
			},
			&ladon.SubjectIsNotOwnerCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<us[er]+>"},
		Effect:      ladon.AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: ladon.Conditions{
			&ladon.CIDRCondition{
				CIDR: "1234",
			},
			&ladon.SubjectIsNotOwnerCondition{},
		},
	},
}

func TestMain(m *testing.M) {
	if testing.Short() {
		os.Exit(0)
	}

	c, err := dockertest.ConnectToRethinkDB(20, time.Second, func(url string) bool {
		rdbSession, err := rdb.Connect(rdb.ConnectOpts{
			Address:  url,
			Database: "hydra"})
		if err != nil {
			return false
		}

		_, err = rdb.DBCreate("hydra").RunWrite(rdbSession)
		if err != nil {
			return false
		}

		s = New(rdbSession)

		if err := s.CreateTables(); err != nil {
			return false
		}

		session = rdbSession

		return true
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	retCode := m.Run()

	// force teardown
	tearDown(session, c)

	os.Exit(retCode)
}

func tearDown(session *rdb.Session, c dockertest.ContainerID) {
	defer session.Close()
	c.KillRemove()
}

func TestCreateErrors(t *testing.T) {
	assert.NotNil(t, &ladon.DefaultPolicy{
		ID: "invalid-format",
	})
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
		pkg.AssertObjectKeysEqual(t, c, get, "Description", "Subjects", "Resources", "Effect", "Actions")
		assert.Equal(t, len(c.Conditions), len(get.GetConditions()))
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

	policies, err = s.FindPoliciesForSubject("peter")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(policies))

	// Test case-sensitive matching
	policies, err = s.FindPoliciesForSubject("User")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(policies))

	// Test user without policy
	policies, err = s.FindPoliciesForSubject("foobar")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(policies))
}
