package manager_test

import (
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/common/integration"
	"github.com/ory/common/pkg"
	"github.com/ory/ladon/access"
	"github.com/ory/ladon/manager"
	"github.com/ory/ladon/policy"
)

var managerPolicies = []*policy.DefaultPolicy{
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"user", "anonymous"},
		Effect:      access.Allow,
		Resources:   []string{"article", "user"},
		Actions:     []string{"create", "update"},
		Conditions:  access.Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{},
		Effect:      access.Allow,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions:  access.Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<peter|max>"},
		Effect:      access.Deny,
		Resources:   []string{"article", "user"},
		Actions:     []string{"view"},
		Conditions: access.Conditions{
			"owner": &access.EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<user|max|anonymous>", "peter"},
		Effect:      access.Deny,
		Resources:   []string{".*"},
		Actions:     []string{"disable"},
		Conditions: access.Conditions{
			"ip": &access.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &access.EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<.*>"},
		Effect:      access.Allow,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: access.Conditions{
			"ip": &access.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &access.EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<us[er]+>"},
		Effect:      access.Allow,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: access.Conditions{
			"ip": &access.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &access.EqualsSubjectCondition{},
		},
	},
}

var managers = make(map[string]manager.Manager)

func init() {
	managers["postgres"], _ = manager.New("sql",
		manager.Connection(integration.ConnectToPostgres("ladon")))
	managers["mysql"], _ = manager.New("sql",
		manager.Connection(integration.ConnectToMySQL()))
	managers["rethink"], _ = manager.New("rethink",
		manager.Connection(integration.ConnectToRethinkDB("ladon", "policies")))
	managers["redis"], _ = manager.New("redis",
		manager.Connection(integration.ConnectToRedis()))
}

func TestMain(m *testing.M) {
	s := m.Run()
	integration.KillAll()
	os.Exit(s)
}

func TestGetErrors(t *testing.T) {
	for k, s := range managers {
		_, err := s.Get(uuid.New())
		assert.EqualError(t, err, pkg.ErrNotFound.Error(), k)

		_, err = s.Get("asdf")
		assert.NotNil(t, err)
	}
}

func TestCreateGetDelete(t *testing.T) {
	for k, s := range managers {
		for _, c := range managerPolicies {
			err := s.Create(c)
			assert.Nil(t, err, "%s: %s", k, err)
			time.Sleep(time.Millisecond * 100)

			get, err := s.Get(c.GetID())
			assert.Nil(t, err, "%s: %s", k, err)
			pkg.AssertObjectKeysEqual(t, c, get, "Description", "Subjects", "Resources", "Effect", "Actions")
			assert.EqualValues(t, c.Conditions, get.GetConditions(), "%s", k)
		}

		for _, c := range managerPolicies {
			assert.Nil(t, s.Delete(c.GetID()), k)
			_, err := s.Get(c.GetID())
			assert.NotNil(t, err, "%s: %s", k, err)
		}
	}
}

func TestFindPoliciesForSubject(t *testing.T) {
	for k, s := range managers {
		for _, c := range managerPolicies {
			require.Nil(t, s.Create(c), k)
		}

		policies, err := s.FindPoliciesForSubject("user")
		assert.Nil(t, err)
		assert.Equal(t, 4, len(policies), k)

		policies, err = s.FindPoliciesForSubject("peter")
		assert.Nil(t, err)
		assert.Equal(t, 3, len(policies), k)

		// Test case-sensitive matching
		policies, err = s.FindPoliciesForSubject("User")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(policies), k)

		// Test case-sensitive matching
		policies, err = s.FindPoliciesForSubject("taKwq")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(policies), k)

		// Test user without policy
		policies, err = s.FindPoliciesForSubject("foobar")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(policies), k)
	}
}
