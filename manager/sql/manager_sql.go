package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	. "github.com/ory-am/ladon"

	"github.com/jmoiron/sqlx"
	"github.com/ory-am/common/compiler"
	"github.com/ory-am/common/pkg"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

var sqlDown = map[string][]string{
	"1": []string{
		"DROP TABLE ladon_policy",
		"DROP TABLE ladon_policy_subject",
		"DROP TABLE ladon_policy_permission",
		"DROP TABLE ladon_policy_resource",
	},
}

var sqlUp = map[string][]string{
	"1": []string{`CREATE TABLE IF NOT EXISTS ladon_policy (
	id           varchar(255) NOT NULL PRIMARY KEY,
	description  text NOT NULL,
	effect       text NOT NULL CHECK (effect='allow' OR effect='deny'),
	conditions 	 text NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_subject (
compiled text NOT NULL,
template varchar(1023) NOT NULL,
policy   varchar(255) NOT NULL,
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_permission (
compiled text NOT NULL,
template varchar(1023) NOT NULL,
policy   varchar(255) NOT NULL,
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_resource (
compiled text NOT NULL,
template varchar(1023) NOT NULL,
policy   varchar(255) NOT NULL,
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE
)`},
}

var migrations = map[string]*migrate.MemoryMigrationSource{
	"postgres": &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{Id: "1", Up: sqlUp["1"], Down: sqlDown["1"]},
			{
				Id: "2",
				Up: []string{
					"CREATE INDEX ladon_policy_subject_compiled_idx ON ladon_policy_subject (compiled text_pattern_ops)",
					"CREATE INDEX ladon_policy_permission_compiled_idx ON ladon_policy_permission (compiled text_pattern_ops)",
					"CREATE INDEX ladon_policy_resource_compiled_idx ON ladon_policy_resource (compiled text_pattern_ops)",
				},
				Down: []string{
					"DROP INDEX ladon_policy_subject_compiled_idx",
					"DROP INDEX ladon_policy_permission_compiled_idx",
					"DROP INDEX ladon_policy_resource_compiled_idx",
				},
			},
		},
	},
	"mysql": &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{Id: "1", Up: sqlUp["1"], Down: sqlDown["1"]},
			{
				Id: "2",
				Up: []string{
					"CREATE FULLTEXT INDEX ladon_policy_subject_compiled_idx ON ladon_policy_subject (compiled)",
					"CREATE FULLTEXT INDEX ladon_policy_permission_compiled_idx ON ladon_policy_permission (compiled)",
					"CREATE FULLTEXT INDEX ladon_policy_resource_compiled_idx ON ladon_policy_resource (compiled)",
				},
				Down: []string{
					"DROP INDEX ladon_policy_subject_compiled_idx",
					"DROP INDEX ladon_policy_permission_compiled_idx",
					"DROP INDEX ladon_policy_resource_compiled_idx",
				},
			},
		},
	},
}

// SQLManager is a postgres implementation for Manager to store policies persistently.
type SQLManager struct {
	db     *sqlx.DB
	schema []string
}

// NewSQLManager initializes a new SQLManager for given db instance.
func NewSQLManager(db *sqlx.DB, schema []string) *SQLManager {
	return &SQLManager{
		db:     db,
		schema: schema,
	}
}

// CreateSchemas creates ladon_policy tables
func (s *SQLManager) CreateSchemas() error {
	var source *migrate.MemoryMigrationSource
	switch s.db.DriverName() {
	case "postgres", "pgx":
		source = migrations["postgres"]
	case "mysql":
		source = migrations["mysql"]
	default:
		return errors.Errorf("Database driver %s is not supported", s.db.DriverName())
	}

	n, err := migrate.Exec(s.db.DB, s.db.DriverName(), source, migrate.Up)
	if err != nil {
		return errors.Wrapf(err, "Could not migrate sql schema, applied %d migrations", n)
	}
	return nil
}

// Create inserts a new policy
func (s *SQLManager) Create(policy Policy) (err error) {
	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if tx, err := s.db.Begin(); err != nil {
		return errors.WithStack(err)
	} else if _, err = tx.Exec(s.db.Rebind("INSERT INTO ladon_policy (id, description, effect, conditions) VALUES (?, ?, ?, ?)"), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	} else if err = createLinkSQL(s.db, tx, "ladon_policy_subject", policy, policy.GetSubjects()); err != nil {
		return err
	} else if err = createLinkSQL(s.db, tx, "ladon_policy_permission", policy, policy.GetActions()); err != nil {
		return err
	} else if err = createLinkSQL(s.db, tx, "ladon_policy_resource", policy, policy.GetResources()); err != nil {
		return err
	} else if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLManager) FindRequestCandidates(r *Request) (Policies, error) {
	var policies = map[string]*DefaultPolicy{}

	var query string = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template, resource.template, permission.template
FROM
	ladon_policy as p

INNER JOIN
	ladon_policy_subject as lps
ON
	lps.policy = p.id

INNER JOIN
	ladon_policy_subject as subject
ON
	subject.policy = p.id
INNER JOIN
	ladon_policy_resource as resource
ON
	resource.policy = p.id
INNER JOIN
	ladon_policy_permission as permission
ON
	permission.policy = p.id

WHERE`
	switch s.db.DriverName() {
	case "postgres", "pgx":
		query = query + `
	$1 ~ ('^' || lps.compiled || '$')`
		break
	case "mysql":
		query = query + `
	? REGEXP BINARY CONCAT('^', lps.compiled, '$')`
		break
	default:
		return nil, errors.Errorf("Database driver %s is not supported", s.db.DriverName())
	}

	rows, err := s.db.Query(query, r.Subject)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var p DefaultPolicy
		var conditions []byte
		var resource, subject, action string

		if err = rows.Scan(&p.ID, &p.Effect, &conditions, &p.Description, &subject, &resource, &action); err != nil {
			return nil, errors.WithStack(err)
		}

		p.Conditions = Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return nil, errors.WithStack(err)
		}

		if c, ok := policies[p.ID]; ok {
			policies[p.ID].Actions = append(c.Actions, action)
			policies[p.ID].Subjects = append(c.Subjects, subject)
			policies[p.ID].Resources = append(c.Resources, resource)
		} else {
			p.Actions = []string{action}
			p.Resources = []string{resource}
			p.Subjects = []string{subject}
			policies[p.ID] = &p
		}
	}


	var result = make(Policies, len(policies))
	var count int
	for _, v := range policies {
		v.Actions = uniq(v.Actions)
		v.Resources = uniq(v.Resources)
		v.Subjects = uniq(v.Subjects)
		result[count] = v
		count++
	}

	return result, nil
}

// Get retrieves a policy.
func (s *SQLManager) Get(id string) (Policy, error) {
	var p DefaultPolicy
	var conditions []byte

	if err := s.db.QueryRow(s.db.Rebind("SELECT id, description, effect, conditions FROM ladon_policy WHERE id=?"), id).Scan(&p.ID, &p.Description, &p.Effect, &conditions); err == sql.ErrNoRows {
		return nil, pkg.ErrNotFound
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	p.Conditions = Conditions{}
	if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
		return nil, errors.WithStack(err)
	}

	subjects, err := getLinkedSQL(s.db, "ladon_policy_subject", id)
	if err != nil {
		return nil, err
	}
	permissions, err := getLinkedSQL(s.db, "ladon_policy_permission", id)
	if err != nil {
		return nil, err
	}
	resources, err := getLinkedSQL(s.db, "ladon_policy_resource", id)
	if err != nil {
		return nil, err
	}

	p.Actions = permissions
	p.Subjects = subjects
	p.Resources = resources
	return &p, nil
}

// Delete removes a policy.
func (s *SQLManager) Delete(id string) error {
	_, err := s.db.Exec(s.db.Rebind("DELETE FROM ladon_policy WHERE id=?"), id)
	return errors.WithStack(err)
}

func getLinkedSQL(db *sqlx.DB, table, policy string) ([]string, error) {
	urns := []string{}
	rows, err := db.Query(db.Rebind(fmt.Sprintf("SELECT template FROM %s WHERE policy=?", table)), policy)
	if err == sql.ErrNoRows {
		return nil, errors.Wrap(pkg.ErrNotFound, "")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	defer rows.Close()
	for rows.Next() {
		var urn string
		if err = rows.Scan(&urn); err != nil {
			return []string{}, errors.WithStack(err)
		}
		urns = append(urns, urn)
	}
	return urns, nil
}

func createLinkSQL(db *sqlx.DB, tx *sql.Tx, table string, p Policy, templates []string) error {
	for _, template := range templates {
		reg, err := compiler.CompileRegex(template, p.GetStartDelimiter(), p.GetEndDelimiter())

		// Execute SQL statement
		query := db.Rebind(fmt.Sprintf("INSERT INTO %s (policy, template, compiled) VALUES (?, ?, ?)", table))
		if _, err = tx.Exec(query, p.GetID(), template, reg.String()); err != nil {
			if rb := tx.Rollback(); rb != nil {
				return errors.WithStack(rb)
			}
			return errors.WithStack(err)
		}
	}
	return nil
}

func uniq(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
