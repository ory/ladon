package ladon_test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	. "github.com/ory/ladon"
	"github.com/ory/ladon/integration"
	. "github.com/ory/ladon/manager/memory"
	. "github.com/ory/ladon/manager/sql"
)

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
	if _, err := s.CreateSchemas("", ""); err != nil {
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
	if _, err := s.CreateSchemas("", ""); err != nil {
		log.Fatalf("Could not create mysql schema: %v", err)
	}

	managers["mysql"] = s
	migrators["mysql"] = &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
		DB:         db,
		SQLManager: s,
	}
}

func TestGetErrors(t *testing.T) {
	for k, s := range managers {
		t.Run("manager="+k, TestHelperGetErrors(s))
	}
}

func TestCreateGetDelete(t *testing.T) {
	for k, s := range managers {
		t.Run(fmt.Sprintf("manager=%s", k), TestHelperCreateGetDelete(s))
	}
}
