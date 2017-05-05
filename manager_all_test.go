package ladon_test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/ory-am/common/integration"
	. "github.com/ory/ladon"
	. "github.com/ory/ladon/manager/memory"
	. "github.com/ory/ladon/manager/sql"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Subjects:    []string{},
		Effect:      AllowAccess,
		Resources:   []string{},
		Actions:     []string{"view"},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{},
		Effect:      AllowAccess,
		Resources:   []string{},
		Actions:     []string{},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{},
		Effect:      AllowAccess,
		Resources:   []string{"foo"},
		Actions:     []string{},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"foo"},
		Effect:      AllowAccess,
		Resources:   []string{"foo"},
		Actions:     []string{},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Subjects:    []string{"foo"},
		Effect:      AllowAccess,
		Resources:   []string{},
		Actions:     []string{},
		Conditions:  Conditions{},
	},
	{
		ID:          uuid.New(),
		Description: "description",
		Effect:      AllowAccess,
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
var migrators = map[string]ManagerMigrator{}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(3)
	connectMySQL(&wg)
	connectPG(&wg)
	connectMEM(&wg)
	wg.Wait()

	s := m.Run()
	integration.KillAll()
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
	if _, err := s.CreateSchemas("",""); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
	migrators["postgres"] = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
		DB:         db,
		SQLManager: s,
	}
}

func connectMySQL(wg *sync.WaitGroup) {
	defer wg.Done()
	var db = integration.ConnectToMySQL()
	s := NewSQLManager(db, nil)
	if _, err := s.CreateSchemas("",""); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
	migrators["mysql"] = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
		DB:         db,
		SQLManager: s,
	}
}

func TestGetErrors(t *testing.T) {
	for _, s := range managers {
		_, err := s.Get(uuid.New())
		assert.Error(t, err)

		_, err = s.Get("asdf")
		assert.Error(t, err)
	}
}

func TestCreateGetDelete(t *testing.T) {
	for k, s := range managers {
		t.Run(fmt.Sprintf("manager=%s", k), func(t *testing.T) {
			for i, c := range managerPolicies {
				t.Run(fmt.Sprintf("case=%d/id=%s/type=create", i, c.GetID()), func(t *testing.T) {
					_, err := s.Get(c.GetID())
					require.Error(t, err)
					require.NoError(t, s.Create(c))
				})

				t.Run(fmt.Sprintf("case=%d/id=%s/type=query", i, c.GetID()), func(t *testing.T) {
					get, err := s.Get(c.GetID())
					require.NoError(t, err)

					assertPolicyEqual(t, c, get)
				})
			}

			t.Run("type=query-all", func(t *testing.T) {
				pols, err := s.GetAll(100, 0)
				require.NoError(t, err)
				assert.Len(t, pols, len(managerPolicies))

				found := map[string]int{}
				for _, got := range pols {
					for _, expect := range managerPolicies {
						if got.GetID() == expect.GetID() {
							found[got.GetID()]++
						}
					}
				}

				for _, f := range found {
					assert.Equal(t, 1, f)
				}
			})

			for i, c := range managerPolicies {
				t.Run(fmt.Sprintf("case=%d/id=%s/type=delete", i, c.GetID()), func(t *testing.T) {
					assert.NoError(t, s.Delete(c.ID))

					_, err := s.Get(c.GetID())
					assert.Error(t, err)
				})
			}
		})
	}
}
