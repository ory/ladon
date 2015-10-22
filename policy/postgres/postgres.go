package postgres

import (
	"database/sql"
	"fmt"
	"github.com/ory-am/ladon/policy"
	"log"
)

var schemas = []string{
	`CREATE TABLE ladon_policy (
		id           text NOT NULL PRIMARY KEY,
		description  text,
		effect       text NOT NULL CHECK (effect='allow' OR effect='deny')
	)`,
	`CREATE TABLE ladon_policy_subject (
    	urn text NOT NULL,
    	policy text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (urn, policy)
	)`,
	`CREATE TABLE ladon_policy_permission (
    	urn text NOT NULL,
    	policy text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (urn, policy)
	)`,
	`CREATE TABLE ladon_policy_resource (
    	urn text NOT NULL,
    	policy text NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    	PRIMARY KEY (urn, policy)
	)`,
}

type Store struct {
	db *sql.DB
}

func (s *Store) CreateSchemas() error {
	for _, sql := range schemas {
		if _, err := s.db.Exec(sql); err != nil {
			log.Printf("Error creating schema %s", sql)
			return err
		}
	}
	return nil
}

func (s *Store) Create(id, description string, effect string, subjects, permissions, resources []string) (policy.Policy, error) {
	if tx, err := s.db.Begin(); err != nil {
		return nil, err
	} else if _, err = tx.Exec("INSERT INTO ladon_policy (id, description, effect) VALUES ($1, $2, $3)", id, description, effect); err != nil {
		return nil, err
	} else if err = createLink(tx, "ladon_policy_subject", id, subjects); err != nil {
		return nil, err
	} else if err = createLink(tx, "ladon_policy_permission", id, permissions); err != nil {
		return nil, err
	} else if err = createLink(tx, "ladon_policy_resource", id, resources); err != nil {
		return nil, err
	} else if err = tx.Commit(); err != nil {
		if rb := tx.Rollback(); rb != nil {
			return nil, rb
		}
		return nil, err
	}

	return &policy.DefaultPolicy{
		ID:          id,
		Description: description,
		Subjects:    subjects,
		Effect:      effect,
		Resources:   resources,
		Permissions: permissions,
	}, nil
}

func (s *Store) Get(id string) (policy.Policy, error) {
	var p policy.DefaultPolicy
	if err := s.db.QueryRow("SELECT id, description, effect FROM ladon_policy WHERE id=$1", id).Scan(&p.ID, &p.Description, &p.Effect); err != nil {
		return nil, err
	}

	subjects, err := getLinked(s.db, "ladon_policy_subject", id)
	if err != nil {
		return nil, err
	}
	permissions, err := getLinked(s.db, "ladon_policy_permission", id)
	if err != nil {
		return nil, err
	}
	resources, err := getLinked(s.db, "ladon_policy_resource", id)
	if err != nil {
		return nil, err
	}

	p.Permissions = permissions
	p.Subjects = subjects
	p.Resources = resources
	return &p, nil
}

func (s *Store) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM ladon_policy WHERE id=$1", id)
	return err
}

func (s *Store) FindPoliciesForSubject(subject string) (policies []policy.Policy, err error) {
	find := func(sql string, args ...interface{}) (ids []string, err error) {
		rows, err := s.db.Query(sql, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var urn string
			if err = rows.Scan(&urn); err != nil {
				return nil, err
			}
			ids = append(ids, urn)
		}
		return ids, nil
	}

	subjects, err := find("SELECT policy FROM ladon_policy_subject WHERE $1 ~* ('^' || urn || '$')", subject)
	if err != nil {
		return policies, err
	}
	globals, err := find("SELECT id FROM ladon_policy p LEFT JOIN ladon_policy_subject ps ON p.id = ps.policy WHERE ps.policy IS NULL")
	if err != nil {
		return policies, err
	}

	ids := append(subjects, globals...)
	for _, id := range ids {
		p, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func getLinked(db *sql.DB, table, policy string) ([]string, error) {
	urns := []string{}
	rows, err := db.Query(fmt.Sprintf("SELECT urn FROM %s WHERE policy=$1", table), policy)
	if err != nil {
		return urns, err
	}
	defer rows.Close()
	for rows.Next() {
		var urn string
		if err = rows.Scan(&urn); err != nil {
			return []string{}, err
		}
		urns = append(urns, urn)
	}
	return urns, nil
}

func createLink(tx *sql.Tx, table, policy string, urns []string) error {
	for _, urn := range urns {
		// Execute SQL statement
		query := fmt.Sprintf("INSERT INTO %s (policy, urn) VALUES ($1, $2)", table)
		_, err := tx.Exec(query, policy, urn)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				return rb
			}
			return err
		}
	}
	return nil
}
