package ladon_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/ory-am/common/pkg"
	"github.com/ory-am/ladon"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"gopkg.in/ory-am/dockertest.v2"
)

var managerPolicies = []*ladon.DefaultPolicy{
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
			"owner": &ladon.EqualsSubjectCondition{},
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
			"ip": &ladon.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &ladon.EqualsSubjectCondition{},
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
			"ip": &ladon.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &ladon.EqualsSubjectCondition{},
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
			"ip": &ladon.CIDRCondition{
				CIDR: "1234",
			},
			"owner": &ladon.EqualsSubjectCondition{},
		},
	},
}

var managers = map[string]ladon.Manager{}

var containers = []dockertest.ContainerID{}

func TestMain(m *testing.M) {
	connectPG()
	connectRDB()
	connectMEM()

	retCode := m.Run()

	for _, c := range containers {
		c.KillRemove()
	}

	os.Exit(retCode)
}

func connectMEM() {
	managers["memory"] = ladon.NewMemoryManager()
}

func connectPG() {
	var db *sql.DB
	c, err := dockertest.ConnectToPostgreSQL(15, time.Second, func(url string) bool {
		var err error
		db, err = sql.Open("postgres", url)
		if err != nil {
			return false
		}
		return db.Ping() == nil
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
	s := ladon.NewPostgresManager(db)

	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	managers["postgres"] = s
}

var rethinkManager *ladon.RethinkManager

func connectRDB() {
	var err error
	var session *r.Session

	c, err := dockertest.ConnectToRethinkDB(20, time.Second, func(url string) bool {
		if session, err = r.Connect(r.ConnectOpts{Address: url, Database: "hydra"}); err != nil {
			return false
		} else if _, err = r.DBCreate("hydra").RunWrite(session); err != nil {
			log.Printf("Database exists: %s", err)
			return false
		} else if _, err = r.TableCreate("hydra_policies").RunWrite(session); err != nil {
			log.Printf("Could not create table: %s", err)
			return false
		}

		rethinkManager = &ladon.RethinkManager{
			Session:  session,
			Table:    r.Table("hydra_policies"),
			Policies: make(map[string]ladon.Policy),
		}

		if err := rethinkManager.Watch(context.Background()); err != nil {
			log.Printf("Could not watch: %s", err)
			return false
		}
		return true
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
	managers["rethink"] = rethinkManager
}

func TestColdStartRethinkManager(t *testing.T) {
	id := uuid.New()
	err := rethinkManager.Create(&ladon.DefaultPolicy{
		ID:          id,
		Description: "description",
		Subjects:    []string{"user", "anonymous"},
		Effect:      ladon.AllowAccess,
		Resources:   []string{"article", "user"},
		Actions:     []string{"create", "update"},
		Conditions:  ladon.Conditions{},
	})
	assert.Nil(t, err)
	time.Sleep(500 * time.Millisecond)
	_, err = rethinkManager.Get(id)
	assert.Nil(t, err)

	rethinkManager.Policies = make(map[string]ladon.Policy)
	_, err = rethinkManager.Get(id)
	assert.NotNil(t, err)

	err = rethinkManager.ColdStart()
	assert.Nil(t, err)

	_, err = rethinkManager.Get(id)
	assert.Nil(t, err)

	rethinkManager.Policies = make(map[string]ladon.Policy)
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

			get, err := s.Get(c.GetID())
			assert.Nil(t, err, "%s: %s", k, err)
			pkg.AssertObjectKeysEqual(t, c, get, "Description", "Subjects", "Resources", "Effect", "Actions")
			assert.Equal(t, len(c.Conditions), len(get.GetConditions()), "%s", k)
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

		// Test user without policy
		policies, err = s.FindPoliciesForSubject("foobar")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(policies), k)
	}
}
