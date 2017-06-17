package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/ory/ladon"
	. "github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
)

type SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7 struct {
	DB         *sqlx.DB
	SQLManager *SQLManager
}

func (s *SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7) GetManager() ladon.Manager {
	return s.SQLManager
}

// Get retrieves a policy.
func (s *SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7) Migrate() error {
	rows, err := s.DB.Query(s.DB.Rebind("SELECT id, description, effect, conditions FROM ladon_policy"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer rows.Close()

	var pols = Policies{}
	for rows.Next() {
		var p DefaultPolicy
		var conditions []byte

		if err := rows.Scan(&p.ID, &p.Description, &p.Effect, &conditions); err != nil {
			return errors.WithStack(err)
		}

		p.Conditions = Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return errors.WithStack(err)
		}

		subjects, err := getLinkedSQL(s.DB, "ladon_policy_subject", p.GetID())
		if err != nil {
			return errors.WithStack(err)
		}
		permissions, err := getLinkedSQL(s.DB, "ladon_policy_permission", p.GetID())
		if err != nil {
			return errors.WithStack(err)
		}
		resources, err := getLinkedSQL(s.DB, "ladon_policy_resource", p.GetID())
		if err != nil {
			return errors.WithStack(err)
		}

		log.Printf("[DEBUG] Found policy %s", p.GetID())

		p.Actions = permissions
		p.Subjects = subjects
		p.Resources = resources
		pols = append(pols, &p)
	}

	log.Printf("[DEBUG] Found %d policies, migrating", len(pols))

	for _, p := range pols {
		log.Printf("[DEBUG] Inserting policy %s", p.GetID())
		if err := s.SQLManager.Create(p); err != nil {
			log.Printf("[DEBUG] Unable to insert policy %d: %s", p.GetID(), err)
			return errors.WithStack(err)
		}
	}

	log.Printf("[DEBUG] Migrated %d policies successfully", len(pols))

	return nil
}

func getLinkedSQL(db *sqlx.DB, table, policy string) ([]string, error) {
	urns := []string{}
	rows, err := db.Query(db.Rebind(fmt.Sprintf("SELECT template FROM %s WHERE policy=?", table)), policy)
	if err == sql.ErrNoRows {
		return nil, errors.Wrap(ladon.ErrNotFound, "")
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

// Create inserts a new policy
func (s *SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7) Create(policy Policy) (err error) {
	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if tx, err := s.DB.Begin(); err != nil {
		return errors.WithStack(err)
	} else if _, err = tx.Exec(s.DB.Rebind("INSERT INTO ladon_policy (id, description, effect, conditions) VALUES (?, ?, ?, ?)"), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	} else if err = createLinkSQL(s.DB, tx, "ladon_policy_subject", policy, policy.GetSubjects()); err != nil {
		return err
	} else if err = createLinkSQL(s.DB, tx, "ladon_policy_permission", policy, policy.GetActions()); err != nil {
		return err
	} else if err = createLinkSQL(s.DB, tx, "ladon_policy_resource", policy, policy.GetResources()); err != nil {
		return err
	} else if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	return nil
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
