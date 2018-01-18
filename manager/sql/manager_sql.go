// Copyright © 2017 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// SQLManager is a postgres implementation for Manager to store policies persistently.
type SQLManager struct {
	db       *sqlx.DB
	database string
}

// NewSQLManager initializes a new SQLManager for given db instance.
func NewSQLManager(db *sqlx.DB, schema []string) *SQLManager {
	database := db.DriverName()
	switch database {
	case "pgx", "pq":
		database = "postgres"
	}

	return &SQLManager{
		db:       db,
		database: database,
	}
}

// CreateSchemas creates ladon_policy tables
func (s *SQLManager) CreateSchemas(schema, table string) (int, error) {
	if _, ok := Migrations[s.database]; !ok {
		return 0, errors.Errorf("Database %s is not supported", s.database)
	}

	source := Migrations[s.database].Migrations

	migrate.SetSchema(schema)
	migrate.SetTable(table)
	n, err := migrate.Exec(s.db.DB, s.database, source, migrate.Up)
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
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
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
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
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

	if _, ok := Migrations[s.database]; !ok {
		return errors.Errorf("Database %s is not supported", s.database)
	}

	if _, err = tx.Exec(s.db.Rebind(Migrations[s.database].QueryInsertPolicy), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
		return errors.WithStack(err)
	}

	type relation struct {
		p []string
		t string
	}
	var relations = []relation{
		{p: policy.GetActions(), t: "action"},
		{p: policy.GetResources(), t: "resource"},
		{p: policy.GetSubjects(), t: "subject"},
	}

	for _, rel := range relations {
		var query string
		var queryRel string

		switch rel.t {
		case "action":
			query = Migrations[s.database].QueryInsertPolicyActions
			queryRel = Migrations[s.database].QueryInsertPolicyActionsRel
		case "resource":
			query = Migrations[s.database].QueryInsertPolicyResources
			queryRel = Migrations[s.database].QueryInsertPolicyResourcesRel
		case "subject":
			query = Migrations[s.database].QueryInsertPolicySubjects
			queryRel = Migrations[s.database].QueryInsertPolicySubjectsRel
		}

		for _, template := range rel.p {
			h := sha256.New()
			h.Write([]byte(template))
			id := fmt.Sprintf("%x", h.Sum(nil))

			compiled, err := compiler.CompileRegex(template, policy.GetStartDelimiter(), policy.GetEndDelimiter())
			if err != nil {
				return errors.WithStack(err)
			}

			if _, err := tx.Exec(s.db.Rebind(query), id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1); err != nil {
				return errors.WithStack(err)
			}
			if _, err := tx.Exec(s.db.Rebind(queryRel), policy.GetID(), id); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (s *SQLManager) FindRequestCandidates(r *Request) (Policies, error) {
	query := Migrations[s.database].QueryRequestCandidates

	rows, err := s.db.Query(s.db.Rebind(query), r.Subject, r.Subject)
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

WHERE p.id=?`

var getAllQuery = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	(SELECT * from ladon_policy ORDER BY id LIMIT ? OFFSET ?) as p

LEFT JOIN ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN ladon_action as action ON ra.action = action.id
LEFT JOIN ladon_resource as resource ON rr.resource = resource.id`

// GetAll returns all policies
func (s *SQLManager) GetAll(limit, offset int64) (Policies, error) {
	query := s.db.Rebind(getAllQuery)

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// Get retrieves a policy.
func (s *SQLManager) Get(id string) (Policy, error) {
	query := s.db.Rebind(getQuery)

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
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
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
