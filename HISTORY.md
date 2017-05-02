# History of breaking changes

## 0.6.0

Version 0.6.0 includes some larger BC breaks. This version focuses on various
performance boosts for both in-memory and SQL adapters, removes some technical debt
and restructures the repository.

### New location

The location of this library changed from `github.com/ory-am/ladon` to `github.com/ory/ladon`.

### Deprecating Redis and RethinkDB

Redis and RethinkDB are no longer maintained by ORY and were moved to
[ory-am/ladon-community](https://github.com/ory-am/ladon-community). The adapters had various
bugs and performance issues which is why they were removed from the official repository.

### New packages

The SQLManager and MemoryManager moved to their own packages in `ladon/manager/sql` and `ladon/manager/memory`.
This change was made to avoid pulling dependencies that are not required by the user.

### IMPORTANT: SQL Changes

The SQLManager was rewritten completely. Now, the database is 3NF (normalized) and includes
various improvements over the previous, naive adapter. The greatest challenge is matching
regular expressions within SQL databases, which causes significant overhead.

While there is an auto-migration for the schema, the data **is not automatically transferred to
the new schema**.

However, we provided a migration helper. For usage, check out
[xxx_manager_sql_migrator_test.go](xxx_manager_sql_migrator_test.go) or this short example:

```go
var db = getSqlDatabaseFromSomewhere()
s := NewSQLManager(db, nil)

if err := s.CreateSchemas(); err != nil {
    log.Fatalf("Could not create mysql schema: %v", err)
}

migrator := &SQLManagerMigrateFromMajor0Minor6ToMajor0Minor7{
    DB:db,
    SQLManager:s,
}

err := migrator.Migrate()
```

Please run this migrator **only once and make back ups before you run it**.
