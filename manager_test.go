package ladon

import (
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/ory-am/common/pkg"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ory-am/common/integration"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

var managerPolicies = []*DefaultPolicy{
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"user", "anonymous"},
		Effect:      AllowAccess,
		Resources:   []string{"article", "user"},
		Actions:     []string{"create", "update"},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{},
		Effect:      AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<peter|max>"},
		Effect:      DenyAccess,
		Resources:   []string{"article", "user"},
		Actions:     []string{"view"},
		Conditions: Conditions{
			"owner": &EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<user|max|anonymous>", "peter"},
		Effect:      DenyAccess,
		Resources:   []string{".*"},
		Actions:     []string{"disable"},
		Conditions: Conditions{
			"ip": &CIDRCondition{
				CIDR: "1234",
			},
			"owner": &EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<.*>"},
		Effect:      AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: Conditions{
			"ip": &CIDRCondition{
				CIDR: "1234",
			},
			"owner": &EqualsSubjectCondition{},
		},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"<us[er]+>"},
		Effect:      AllowAccess,
		Resources:   []string{"<article|user>"},
		Actions:     []string{"view"},
		Conditions: Conditions{
			"ip": &CIDRCondition{
				CIDR: "1234",
			},
			"owner": &EqualsSubjectCondition{},
		},
	},
}

var managers = map[string]Manager{}
var rethinkManager *RethinkManager

func TestMain(m *testing.M) {
	connectPG()
	connectRDB()
	connectMEM()
	connectPG()
	connectMySQL()
	connectRedis()

	s := m.Run()
	integration.KillAll()
	os.Exit(s)
}

func connectMEM() {
	managers["memory"] = NewMemoryManager()
}

func connectPG() {
	var db = integration.ConnectToPostgres("ladon")
	s := NewSQLManager(db, nil)
	if err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
}

func connectMySQL() {
	var db = integration.ConnectToMySQL()
	s := NewSQLManager(db, nil)
	if err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
}

func connectRDB() {
	var session = integration.ConnectToRethinkDB("ladon", "policies")
	rethinkManager = NewRethinkManager(session, "")

	rethinkManager.Watch(context.Background())
	time.Sleep(time.Second)
	managers["rethink"] = rethinkManager
}

func connectRedis() {
	var db = integration.ConnectToRedis()
	managers["redis"] = NewRedisManager(db, "")
}

func TestColdStart(t *testing.T) {
	assert.Nil(t, rethinkManager.Create(&DefaultPolicy{ID: "foo", Description: "description foo"}))
	assert.Nil(t, rethinkManager.Create(&DefaultPolicy{ID: "bar", Description: "description bar"}))

	time.Sleep(time.Second / 2)
	rethinkManager.Policies = make(map[string]Policy)
	assert.Nil(t, rethinkManager.ColdStart())

	c1, err := rethinkManager.Get("foo")
	assert.Nil(t, err)
	c2, err := rethinkManager.Get("bar")
	assert.Nil(t, err)

	assert.NotEqual(t, c1, c2)
	assert.Equal(t, "description foo", c1.GetDescription())
	assert.Equal(t, "description bar", c2.GetDescription())
	rethinkManager.Policies = make(map[string]Policy)
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
