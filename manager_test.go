package ladon_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory-am/common/pkg"
	"github.com/ory-am/dockertest"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/ory-am/ladon"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	r "gopkg.in/dancannon/gorethink.v2"
	"gopkg.in/redis.v5"
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
var containers = []dockertest.ContainerID{}
var rethinkManager *RethinkManager

func TestMain(m *testing.M) {
	defer func() {
		for _, c := range containers {
			c.KillRemove()
		}
	}()

	connectPG()
	connectRDB()
	connectMEM()
	connectPG()
	connectMySQL()
	connectRedis()

	retCode := m.Run()
	os.Exit(retCode)
}

func connectMEM() {
	managers["memory"] = NewMemoryManager()
}

func connectPG() {
	var db *sqlx.DB
	c, err := dockertest.ConnectToPostgreSQL(15, time.Second, func(url string) bool {
		var err error
		db, err = sqlx.Open("postgres", url)
		if err != nil {
			log.Printf("Got error in postgres connector: %s", err)
			return false
		}
		return db.Ping() == nil
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
	s := NewSQLManager(db, nil)

	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
}

func connectMySQL() {
	var db *sqlx.DB
	c, err := dockertest.ConnectToMySQL(15, time.Second, func(url string) bool {
		var err error
		db, err = sqlx.Open("mysql", url)
		if err != nil {
			log.Printf("Got error in mysql connector: %s", err)
			return false
		}
		return db.Ping() == nil
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
	s := NewSQLManager(db, nil)

	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
}

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

		rethinkManager = &RethinkManager{
			Session:  session,
			Table:    r.Table("hydra_policies"),
			Policies: make(map[string]Policy),
		}

		rethinkManager.Watch(context.Background())
		time.Sleep(time.Second)
		return true
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
	managers["rethink"] = rethinkManager
}

func connectRedis() {
	var db *redis.Client
	c, err := dockertest.ConnectToRedis(15, time.Second, func(url string) bool {
		db = redis.NewClient(&redis.Options{
			Addr: url,
		})

		return db.Ping().Err() == nil
	})

	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	containers = append(containers, c)
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
