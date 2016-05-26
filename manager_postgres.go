package ladon

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-errors/errors"
	"github.com/ory-am/common/compiler"
	"github.com/ory-am/common/pkg"
)

var schemas = []string{
	`CREATE TABLE IF NOT EXISTS ladon_policy (
		id           text NOT NULL PRIMARY KEY,
		description  text DEFAULT '',
		created_at   timestamp DEFAULT NOW(),
		previous	 text NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
		effect       text NOT NULL CHECK (effect='allow' OR effect='deny'),
		conditions 	 json DEFAULT '[]'
	)`,
	`CREATE TABLE IF NOT EXISTS ladon_policy_subject (
    	compiled text NOT NULL,
    	template text NOT NULL,
    	policy   text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (template, policy)
	)`,
	`CREATE TABLE IF NOT EXISTS ladon_policy_permission (
    	compiled text NOT NULL,
    	template text NOT NULL,
    	policy   text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (template, policy)
	)`,
	`CREATE TABLE IF NOT EXISTS ladon_policy_resource (
    	compiled text NOT NULL,
    	template text NOT NULL,
    	policy   text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (template, policy)
	)`,
}

// Manager is a postgres implementation of Manager.
type PostgresManager struct {
	db *sql.DB
}

func NewPostgresManager(db *sql.DB) *PostgresManager {
	return &PostgresManager{db}
}

func (s *PostgresManager) CreateSchemas() error {
	for _, query := range schemas {
		if _, err := s.db.Exec(query); err != nil {
			log.Printf("Error creating schema %s", query)
			return err
		}
	}
	return nil
}

func (s *PostgresManager) Create(policy Policy) (err error) {
	conditions := []byte("[]")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return err
		}
	}

	if tx, err := s.db.Begin(); err != nil {
		return err
	} else if _, err = tx.Exec("INSERT INTO ladon_policy (id, description, effect, conditions) VALUES ($1, $2, $3, $4)", policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
		return err
	} else if err = createLinkPG(tx, "ladon_policy_subject", policy, policy.GetSubjects()); err != nil {
		return err
	} else if err = createLinkPG(tx, "ladon_policy_permission", policy, policy.GetActions()); err != nil {
		return err
	} else if err = createLinkPG(tx, "ladon_policy_resource", policy, policy.GetResources()); err != nil {
		return err
	} else if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (s *PostgresManager) Get(id string) (Policy, error) {
	var p DefaultPolicy
	var conditions []byte
	if err := s.db.QueryRow("SELECT id, description, effect, conditions FROM ladon_policy WHERE id=$1", id).Scan(&p.ID, &p.Description, &p.Effect, &conditions); err == sql.ErrNoRows {
		return nil, pkg.ErrNotFound
	} else if err != nil {
		return nil, errors.New(err)
	}

	p.Conditions = Conditions{}
	if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
		return nil, errors.New(err)
	}

	subjects, err := getLinkedPG(s.db, "ladon_policy_subject", id)
	if err != nil {
		return nil, err
	}
	permissions, err := getLinkedPG(s.db, "ladon_policy_permission", id)
	if err != nil {
		return nil, err
	}
	resources, err := getLinkedPG(s.db, "ladon_policy_resource", id)
	if err != nil {
		return nil, err
	}

	p.Actions = permissions
	p.Subjects = subjects
	p.Resources = resources
	return &p, nil
}

func (s *PostgresManager) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM ladon_policy WHERE id=$1", id)
	return err
}

func (s *PostgresManager) FindPoliciesForSubject(subject string) (policies Policies, err error) {
	find := func(query string, args ...interface{}) (ids []string, err error) {
		rows, err := s.db.Query(query, args...)
		if err == sql.ErrNoRows {
			return nil, pkg.ErrNotFound
		} else if err != nil {
			return nil, errors.New(err)
		}
		defer rows.Close()
		for rows.Next() {
			var urn string
			if err = rows.Scan(&urn); err != nil {
				return nil, errors.New(err)
			}
			ids = append(ids, urn)
		}
		return ids, nil
	}

	subjects, err := find("SELECT policy FROM ladon_policy_subject WHERE $1 ~ ('^' || compiled || '$')", subject)
	if err != nil {
		return policies, err
	}

	for _, id := range subjects {
		p, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func getLinkedPG(db *sql.DB, table, policy string) ([]string, error) {
	urns := []string{}
	rows, err := db.Query(fmt.Sprintf("SELECT template FROM %s WHERE policy=$1", table), policy)
	if err == sql.ErrNoRows {
		return nil, pkg.ErrNotFound
	} else if err != nil {
		return nil, errors.New(err)
	}

	defer rows.Close()
	for rows.Next() {
		var urn string
		if err = rows.Scan(&urn); err != nil {
			return []string{}, errors.New(err)
		}
		urns = append(urns, urn)
	}
	return urns, nil
}

func createLinkPG(tx *sql.Tx, table string, p Policy, templates []string) error {
	for _, template := range templates {
		reg, err := compiler.CompileRegex(template, p.GetStartDelimiter(), p.GetEndDelimiter())

		// Execute SQL statement
		query := fmt.Sprintf("INSERT INTO %s (policy, template, compiled) VALUES ($1, $2, $3)", table)
		if _, err = tx.Exec(query, p.GetID(), template, reg.String()); err != nil {
			if rb := tx.Rollback(); rb != nil {
				return rb
			}
			return err
		}
	}
	return nil
}
