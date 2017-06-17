package sql

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	. "github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

var sqlDown = map[string][]string{
	"1": {
		"DROP TABLE ladon_policy",
		"DROP TABLE ladon_policy_subject",
		"DROP TABLE ladon_policy_permission",
		"DROP TABLE ladon_policy_resource",
	},
}

var sqlUp = map[string][]string{
	"1": {`CREATE TABLE IF NOT EXISTS ladon_policy (
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
	"2": {`CREATE TABLE IF NOT EXISTS ladon_subject (
id          varchar(64) NOT NULL PRIMARY KEY,
has_regex   bool NOT NULL,
compiled 	varchar(511) NOT NULL UNIQUE,
template 	varchar(511) NOT NULL UNIQUE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_action (
id       varchar(64) NOT NULL PRIMARY KEY,
has_regex   bool NOT NULL,
compiled varchar(511) NOT NULL UNIQUE,
template varchar(511) NOT NULL UNIQUE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_resource (
id       varchar(64) NOT NULL PRIMARY KEY,
has_regex   bool NOT NULL,
compiled varchar(511) NOT NULL UNIQUE,
template varchar(511) NOT NULL UNIQUE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_subject_rel (
policy varchar(255) NOT NULL,
subject varchar(64) NOT NULL,
PRIMARY KEY (policy, subject),
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
FOREIGN KEY (subject) REFERENCES ladon_subject(id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_action_rel (
policy varchar(255) NOT NULL,
action varchar(64) NOT NULL,
PRIMARY KEY (policy, action),
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
FOREIGN KEY (action) REFERENCES ladon_action(id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS ladon_policy_resource_rel (
policy varchar(255) NOT NULL,
resource varchar(64) NOT NULL,
PRIMARY KEY (policy, resource),
FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
FOREIGN KEY (resource) REFERENCES ladon_resource(id) ON DELETE CASCADE
)`,
	},
}

var migrations = map[string]*migrate.MemoryMigrationSource{
	"postgres": {
		Migrations: []*migrate.Migration{
			{Id: "1", Up: sqlUp["1"], Down: sqlDown["1"]},
			{
				Id: "2",
				Up: sqlUp["2"],
			},
			{Id: "3",
				Up: []string{
					"CREATE INDEX ladon_subject_compiled_idx ON ladon_subject (compiled text_pattern_ops)",
					"CREATE INDEX ladon_permission_compiled_idx ON ladon_action (compiled text_pattern_ops)",
					"CREATE INDEX ladon_resource_compiled_idx ON ladon_resource (compiled text_pattern_ops)",
				},
				Down: []string{
					"DROP INDEX ladon_subject_compiled_idx",
					"DROP INDEX ladon_permission_compiled_idx",
					"DROP INDEX ladon_resource_compiled_idx",
				},
			},
		},
	},
	"mysql": {
		Migrations: []*migrate.Migration{
			{Id: "1", Up: sqlUp["1"], Down: sqlDown["1"]},
			{
				Id:   "2",
				Up:   sqlUp["2"],
				Down: []string{},
			},
			{
				Id: "3",
				Up: []string{
					"CREATE FULLTEXT INDEX ladon_subject_compiled_idx ON ladon_subject (compiled)",
					"CREATE FULLTEXT INDEX ladon_action_compiled_idx ON ladon_action (compiled)",
					"CREATE FULLTEXT INDEX ladon_resource_compiled_idx ON ladon_resource (compiled)",
				},
				Down: []string{
					"DROP INDEX ladon_subject_compiled_idx",
					"DROP INDEX ladon_permission_compiled_idx",
					"DROP INDEX ladon_resource_compiled_idx",
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
func (s *SQLManager) CreateSchemas(schema, table string) (int, error) {
	var source *migrate.MemoryMigrationSource
	switch s.db.DriverName() {
	case "postgres", "pgx":
		source = migrations["postgres"]
	case "mysql":
		source = migrations["mysql"]
	default:
		return 0, errors.Errorf("Database driver %s is not supported", s.db.DriverName())
	}

	migrate.SetSchema(schema)
	migrate.SetTable(table)
	n, err := migrate.Exec(s.db.DB, s.db.DriverName(), source, migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema, applied %d migrations", n)
	}
	return n, nil
}

// Update updates an existing policy.
func (s *SQLManager) Update(policy Policy) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(policy.GetID(), tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := s.create(policy, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	return nil
}

// Create inserts a new policy
func (s *SQLManager) Create(policy Policy) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLManager) create(policy Policy, tx *sqlx.Tx) (err error) {
	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	switch s.db.DriverName() {
	case "postgres", "pgx":
		if _, err = tx.Exec(s.db.Rebind("INSERT INTO ladon_policy (id, description, effect, conditions) SELECT ?, ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM ladon_policy WHERE id = ?)"), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions, policy.GetID()); err != nil {
			if err := tx.Rollback(); err != nil {
				return errors.WithStack(err)
			}
			return errors.WithStack(err)
		}
	case "mysql":
		if _, err = tx.Exec(s.db.Rebind("INSERT IGNORE INTO ladon_policy (id, description, effect, conditions) VALUES (?, ?, ?, ?)"), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
			if err := tx.Rollback(); err != nil {
				return errors.WithStack(err)
			}
			return errors.WithStack(err)
		}
	default:
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.Errorf("Database driver %s is not supported", s.db.DriverName())
	}

	type relation struct {
		p []string
		t string
	}
	var relations = []relation{{p: policy.GetActions(), t: "action"}, {p: policy.GetResources(), t: "resource"}, {p: policy.GetSubjects(), t: "subject"}}

	for _, v := range relations {
		for _, template := range v.p {
			h := sha256.New()
			h.Write([]byte(template))
			id := fmt.Sprintf("%x", h.Sum(nil))

			compiled, err := compiler.CompileRegex(template, policy.GetStartDelimiter(), policy.GetEndDelimiter())
			if err != nil {
				if err := tx.Rollback(); err != nil {
					return errors.WithStack(err)
				}
				return errors.WithStack(err)
			}

			switch s.db.DriverName() {
			case "postgres", "pgx":
				if _, err := tx.Exec(s.db.Rebind(fmt.Sprintf("INSERT INTO ladon_%s (id, template, compiled, has_regex) SELECT ?, ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM ladon_%[1]s WHERE id = ?)", v.t)), id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) > -1, id); err != nil {
					if err := tx.Rollback(); err != nil {
						return errors.WithStack(err)
					}
					return errors.WithStack(err)
				}

				if _, err := tx.Exec(s.db.Rebind(fmt.Sprintf("INSERT INTO ladon_policy_%s_rel (policy, %[1]s) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM ladon_policy_%[1]s_rel WHERE policy = ? AND %[1]s = ?)", v.t)), policy.GetID(), id, policy.GetID(), id); err != nil {
					if err := tx.Rollback(); err != nil {
						return errors.WithStack(err)
					}
					return errors.WithStack(err)
				}
				break

			case "mysql":
				if _, err := tx.Exec(s.db.Rebind(fmt.Sprintf("INSERT IGNORE INTO ladon_%s (id, template, compiled, has_regex) VALUES (?, ?, ?, ?)", v.t)), id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) > -1); err != nil {
					if err := tx.Rollback(); err != nil {
						return errors.WithStack(err)
					}
					return errors.WithStack(err)
				}

				if _, err := tx.Exec(s.db.Rebind(fmt.Sprintf("INSERT IGNORE INTO ladon_policy_%s_rel (policy, %s) VALUES (?, ?)", v.t, v.t)), policy.GetID(), id); err != nil {
					if err := tx.Rollback(); err != nil {
						return errors.WithStack(err)
					}
					return errors.WithStack(err)
				}
				break
			default:
				if err := tx.Rollback(); err != nil {
					return errors.WithStack(err)
				}
				return errors.Errorf("Database driver %s is not supported", s.db.DriverName())
			}
		}
	}

	return nil
}

func (s *SQLManager) FindRequestCandidates(r *Request) (Policies, error) {
	var query string = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	ladon_policy as p

INNER JOIN ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN ladon_policy_resource_rel as rr ON rr.policy = p.id

INNER JOIN ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN ladon_action as action ON ra.action = action.id
LEFT JOIN ladon_resource as resource ON rr.resource = resource.id

WHERE`
	switch s.db.DriverName() {
	case "postgres", "pgx":
		query = query + `
( subject.has_regex IS NOT TRUE AND subject.template = $1 )
OR
( subject.has_regex IS TRUE AND $2 ~ subject.compiled )`
		break
	case "mysql":
		query = query + `
( subject.has_regex = 0 AND subject.template = ? )
OR
( subject.has_regex = 1 AND ? REGEXP BINARY subject.compiled )`
		break
	default:
		return nil, errors.Errorf("Database driver %s is not supported", s.db.DriverName())
	}

	rows, err := s.db.Query(query, r.Subject, r.Subject)
	if err == sql.ErrNoRows {
		return nil, NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func scanRows(rows *sql.Rows) (Policies, error) {
	var policies = map[string]*DefaultPolicy{}

	for rows.Next() {
		var p DefaultPolicy
		var conditions []byte
		var resource, subject, action sql.NullString
		p.Actions = []string{}
		p.Subjects = []string{}
		p.Resources = []string{}

		if err := rows.Scan(&p.ID, &p.Effect, &conditions, &p.Description, &subject, &resource, &action); err == sql.ErrNoRows {
			return nil, NewErrResourceNotFound(err)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		p.Conditions = Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return nil, errors.WithStack(err)
		}

		if c, ok := policies[p.ID]; ok {
			if action.Valid {
				policies[p.ID].Actions = append(c.Actions, action.String)
			}

			if subject.Valid {
				policies[p.ID].Subjects = append(c.Subjects, subject.String)
			}

			if resource.Valid {
				policies[p.ID].Resources = append(c.Resources, resource.String)
			}
		} else {
			if action.Valid {
				p.Actions = []string{action.String}
			}

			if subject.Valid {
				p.Subjects = []string{subject.String}
			}

			if resource.Valid {
				p.Resources = []string{resource.String}
			}

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

var getQuery = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	ladon_policy as p

LEFT JOIN ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN ladon_action as action ON ra.action = action.id
LEFT JOIN ladon_resource as resource ON rr.resource = resource.id

`

// GetAll returns all policies
func (s *SQLManager) GetAll(limit, offset int64) (Policies, error) {
	query := s.db.Rebind(getQuery + "LIMIT ? OFFSET ?")

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// Get retrieves a policy.
func (s *SQLManager) Get(id string) (Policy, error) {
	query := s.db.Rebind(getQuery + "WHERE p.id=?")

	rows, err := s.db.Query(query, id)
	if err == sql.ErrNoRows {
		return nil, NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	policies, err := scanRows(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if len(policies) == 0 {
		return nil, NewErrResourceNotFound(sql.ErrNoRows)
	}

	return policies[0], nil
}

// Delete removes a policy.
func (s *SQLManager) Delete(id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(id, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	return nil
}

// Delete removes a policy.
func (s *SQLManager) delete(id string, tx *sqlx.Tx) error {
	_, err := tx.Exec(s.db.Rebind("DELETE FROM ladon_policy WHERE id=?"), id)
	return errors.WithStack(err)
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
