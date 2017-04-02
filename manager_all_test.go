package ladon_test

import (
	. "github.com/ory-am/ladon"
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
	"golang.org/x/net/context"
	r "github.com/GoRethink/gorethink"
	. "github.com/ory-am/ladon/manager/rethink"
	. "github.com/ory-am/ladon/manager/redis"
	. "github.com/ory-am/ladon/manager/sql"
	. "github.com/ory-am/ladon/manager/memory"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
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
	var wg sync.WaitGroup
	wg.Add(5)
	 connectMySQL(&wg)
	 connectPG(&wg)
	 connectRDB(&wg)
	 connectRedis(&wg)
	 connectMEM(&wg)
	//wg.Wait()

	s := m.Run()
	//integration.KillAll()
	os.Exit(s)
}

func connectMEM(wg *sync.WaitGroup) {
	defer wg.Done()
	managers["memory"] = NewMemoryManager()
}

func connectPG(wg *sync.WaitGroup) {
	defer wg.Done()
	var db = integration.ConnectToPostgres("ladon")
	s := NewSQLManager(db, nil)
	if err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
}

func connectMySQL(wg *sync.WaitGroup) {
	defer wg.Done()
	var db = integration.ConnectToMySQL()
	s := NewSQLManager(db, nil)
	if err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
}

func connectRDB(wg *sync.WaitGroup) {
	defer wg.Done()
	var session = integration.ConnectToRethinkDB("ladon", "policies")
	rethinkManager = &RethinkManager{
		Session:  session,
		Table:    r.Table("policies"),
		Policies: make(map[string]Policy),
	}

	rethinkManager.Watch(context.Background())
	time.Sleep(time.Second)
	managers["rethink"] = rethinkManager
}

func connectRedis(wg *sync.WaitGroup) {
	defer wg.Done()
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
		t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {
			for i, c := range managerPolicies {
				t.Run(fmt.Sprintf("case=%d/id=%s", i, c.GetID()), func(t *testing.T) {
					require.Nil(t, s.Create(c))
					time.Sleep(time.Millisecond * 100)

					get, err := s.Get(c.GetID())
					require.Nil(t, err)

					assertPolicyEqual(t, get, c)

					require.Nil(t, s.Delete(c.GetID()), k)
					_, err = s.Get(c.GetID())
					assert.NotNil(t, err)
				})
			}
		})
	}
}
