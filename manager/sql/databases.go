package sql

import migrate "github.com/rubenv/sql-migrate"

type Statements struct {
	Migrations                    *migrate.MemoryMigrationSource
	QueryInsertPolicy             string
	QueryInsertPolicyActions      string
	QueryInsertPolicyActionsRel   string
	QueryInsertPolicyResources    string
	QueryInsertPolicyResourcesRel string
	QueryInsertPolicySubjects     string
	QueryInsertPolicySubjectsRel  string
	QueryRequestCandidates        string
}

var sharedMigrations = []*migrate.Migration{
	{
		Id: "1",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS ladon_policy (
				id           varchar(255) NOT NULL PRIMARY KEY,
				description  text NOT NULL,
				effect       text NOT NULL CHECK (effect='allow' OR effect='deny'),
				conditions	 text NOT NULL
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
			)`,
		},
		Down: []string{
			"DROP TABLE ladon_policy",
			"DROP TABLE ladon_policy_subject",
			"DROP TABLE ladon_policy_permission",
			"DROP TABLE ladon_policy_resource",
		},
	},
	{
		Id: "2",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS ladon_subject (
				id          varchar(64) NOT NULL PRIMARY KEY,
				has_regex   bool NOT NULL,
				compiled    varchar(511) NOT NULL UNIQUE,
				template    varchar(511) NOT NULL UNIQUE
			)`,
			`CREATE TABLE IF NOT EXISTS ladon_action (
				id          varchar(64) NOT NULL PRIMARY KEY,
				has_regex   bool NOT NULL,
				compiled    varchar(511) NOT NULL UNIQUE,
				template    varchar(511) NOT NULL UNIQUE
			)`,
			`CREATE TABLE IF NOT EXISTS ladon_resource (
				id          varchar(64) NOT NULL PRIMARY KEY,
				has_regex   bool NOT NULL,
				compiled    varchar(511) NOT NULL UNIQUE,
				template    varchar(511) NOT NULL UNIQUE
			)`,
			`CREATE TABLE IF NOT EXISTS ladon_policy_subject_rel (
				policy   varchar(255) NOT NULL,
				subject  varchar(64) NOT NULL,
				PRIMARY KEY (policy, subject),
				FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
				FOREIGN KEY (subject) REFERENCES ladon_subject(id) ON DELETE CASCADE
			)`,
			`CREATE TABLE IF NOT EXISTS ladon_policy_action_rel (
				policy  varchar(255) NOT NULL,
				action  varchar(64) NOT NULL,
				PRIMARY KEY (policy, action),
				FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
				FOREIGN KEY (action) REFERENCES ladon_action(id) ON DELETE CASCADE
			)`,
			`CREATE TABLE IF NOT EXISTS ladon_policy_resource_rel (
				policy    varchar(255) NOT NULL,
				resource  varchar(64) NOT NULL,
				PRIMARY KEY (policy, resource),
				FOREIGN KEY (policy) REFERENCES ladon_policy(id) ON DELETE CASCADE,
				FOREIGN KEY (resource) REFERENCES ladon_resource(id) ON DELETE CASCADE
			)`,
		},
		Down: []string{},
	},
}

var Migrations = map[string]Statements{
	"postgres": {
		Migrations: &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				sharedMigrations[0],
				sharedMigrations[1],
				{
					Id: "3",
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
		QueryInsertPolicy:             `INSERT INTO ladon_policy(id, description, effect, conditions) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM ladon_policy WHERE id = $1)`,
		QueryInsertPolicyActions:      `INSERT INTO ladon_action (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM ladon_action WHERE id = $1)`,
		QueryInsertPolicyActionsRel:   `INSERT INTO ladon_policy_action_rel (policy, action) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM ladon_policy_action_rel WHERE policy = $1 AND action = $2)`,
		QueryInsertPolicyResources:    `INSERT INTO ladon_resource (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM ladon_resource WHERE id = $1)`,
		QueryInsertPolicyResourcesRel: `INSERT INTO ladon_policy_resource_rel (policy, resource) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM ladon_policy_resource_rel WHERE policy = $1 AND resource = $2)`,
		QueryInsertPolicySubjects:     `INSERT INTO ladon_subject (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM ladon_subject WHERE id = $1)`,
		QueryInsertPolicySubjectsRel:  `INSERT INTO ladon_policy_subject_rel (policy, subject) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM ladon_policy_subject_rel WHERE policy = $1 AND subject = $2)`,
		QueryRequestCandidates: `
		SELECT
			p.id,
			p.effect,
			p.conditions,
			p.description,
			subject.template AS subject,
			resource.template AS resource,
			action.template AS action
		FROM
			ladon_policy AS p

			INNER JOIN ladon_policy_subject_rel AS rs ON rs.policy = p.id
			LEFT JOIN ladon_policy_action_rel AS ra ON ra.policy = p.id
			LEFT JOIN ladon_policy_resource_rel AS rr ON rr.policy = p.id

			INNER JOIN ladon_subject AS subject ON rs.subject = subject.id
			LEFT JOIN ladon_action AS action ON ra.action = action.id
			LEFT JOIN ladon_resource AS resource ON rr.resource = resource.id
		WHERE
			(subject.has_regex IS NOT TRUE AND subject.template = $1)
			OR
			(subject.has_regex IS TRUE AND $2 ~ subject.compiled)`,
	},
	"mysql": {
		Migrations: &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				sharedMigrations[0],
				sharedMigrations[1],
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
		QueryInsertPolicy:             `INSERT IGNORE INTO ladon_policy (id, description, effect, conditions) VALUES(?,?,?,?)`,
		QueryInsertPolicyActions:      `INSERT IGNORE INTO ladon_action (id, template, compiled, has_regex) VALUES(?,?,?,?)`,
		QueryInsertPolicyActionsRel:   `INSERT IGNORE INTO ladon_policy_action_rel (policy, action) VALUES(?,?)`,
		QueryInsertPolicyResources:    `INSERT IGNORE INTO ladon_resource (id, template, compiled, has_regex) VALUES(?,?,?,?)`,
		QueryInsertPolicyResourcesRel: `INSERT IGNORE INTO ladon_policy_resource_rel (policy, resource) VALUES(?,?)`,
		QueryInsertPolicySubjects:     `INSERT IGNORE INTO ladon_subject (id, template, compiled, has_regex) VALUES(?,?,?,?)`,
		QueryInsertPolicySubjectsRel:  `INSERT IGNORE INTO ladon_policy_subject_rel (policy, subject) VALUES(?,?)`,
		QueryRequestCandidates: `
		SELECT
			p.id,
			p.effect,
			p.conditions,
			p.description,
			subject.template AS subject,
			resource.template AS resource,
			action.template AS action
		FROM
			ladon_policy AS p

			INNER JOIN ladon_policy_subject_rel AS rs ON rs.policy = p.id
			LEFT JOIN ladon_policy_action_rel AS ra ON ra.policy = p.id
			LEFT JOIN ladon_policy_resource_rel AS rr ON rr.policy = p.id

			INNER JOIN ladon_subject AS subject ON rs.subject = subject.id
			LEFT JOIN ladon_action AS action ON ra.action = action.id
			LEFT JOIN ladon_resource AS resource ON rr.resource = resource.id
		WHERE
			(subject.has_regex = 0 AND subject.template = ?)
			OR
			(subject.has_regex = 1 AND ? REGEXP BINARY subject.compiled)`,
	},
}
