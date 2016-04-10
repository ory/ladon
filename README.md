# Ladon

[![Build Status](https://travis-ci.org/ory-am/ladon.svg?branch=master)](https://travis-ci.org/ory-am/ladon)
[![Coverage Status](https://coveralls.io/repos/ory-am/ladon/badge.svg?branch=master&service=github)](https://coveralls.io/github/ory-am/ladon?branch=master)

![Ladon](https://upload.wikimedia.org/wikipedia/commons/5/5c/Reggio_calabria_museo_nazionale_mosaico_da_kaulon.jpg)

[Ladon](https://en.wikipedia.org/wiki/Ladon_%28mythology%29) is the serpent dragon protecting your resources.
A policy based authorization library written in [Go](https://golang.org). Ships with PostgreSQL and RethinkDB storage interfaces.

Utilizes ory-am/dockertest V2 for tests. Please refer to [ory-am/dockertest](https://github.com/ory-am/dockertest) for more information on how to setup testing environment.

Please be aware that ladon does not have a stable release just yet. Once it does, it will be available through gopkg.in.

## What is this and how does it work?

Ladon is an access control library. You might also call it a policy administration and policy decision point. Ladon
answers the question:

> **Who** is **able** to do **what** on **something** in a **context**

* **Who** An arbitrary unique subject name, for example "ken" or "printer-service.mydomain.com".
* **Able**: The effect which is always "allow" or "deny".
* **What**: An arbitrary action name, for example "delete", "create" or "scoped:action:something".
* **Something**: An arbitrary unique resource name, for example "something", "resources:articles:1234" or some uniform
  resource name like "urn:isbn:3827370191".
* **Context**: The current context which may environment information like the IP Address, request date or the resource owner name amongst other things.

## Usage

### Policies

Policies are an essential part of Ladon. Policies are documents which define who is allowed to perform an action on a resource.
Policies must implement the `ladon.Policy` interface. A standard implementation of this interface is `ladon.DefaultPolicy`:

```go
import github.com/ory-am/ladon

var pol := &ladon.DefaultPolicy{
    // A required unique identifier. Used primarily for database retrieval.
    ID: "68819e5a-738b-41ec-b03c-b58a1b19d043",

    // A optional human readable description.
    Description: "something humanly readable",

    // A subject can be an user or a service. It is the "who" in "who is allowed to do what on something".
    // As you can see here, you can use regular expressions inside < >.
    Subjects: []string{"max", "peter", "<zac|ken>"},

    // Which resources this policy affects.
    // Again, you can put regular expressions in inside < >.
    Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},

    // Which actions this policy affects. Supports RegExp
    // Again, you can put regular expressions in inside < >.
    Actions: []string{"<create|delete>", "get"},

    // Under which conditions this policy is "active".
    Conditions: Conditions{
        // In this example, the policy is only "active" when the requested subject is the owner of the resource as well.
        ConditionSubjectIsOwner{},

        // Additionally, the policy will only match if the requests remote ip address matches 127.0.0.1
        CIDRCondition{
            CIDR: "127.0.0.1/32",
        },
    }
}
```

## Policy management

Ladon comes with `ladon.Manager`, a policy management interface which is implemented using RethinkDB and PostgreSQL.
Storing policies.

**A word on Condition creators**
Unmarshalling lists with multiple types is not trivial in Go. Ladon comes with creators (factories) for the different conditions.
The manager receives a list of allowed condition creators who assist him in finding and creating the right condition objects.

### In memory

```go
import "github.com/ory-am/ladon"
import "github.com/ory-am/ladon/memory"


func main() {
    warden := ladon.Ladon{
        Manager: memory.New()
    }

    // ...
}
```

### Using a backend

You will notice that all persistent implementations require an additional argument when setting up. This argument
is called `allowedConditionCreators` and must contain a list of allowed condition creators (or "factories"). Because it is
not trivial to unmarshal lists of various types (required by `ladon.Conditions`), we wrote some helpers to do that for you.

You can always pass `ladon.DefaultConditionCreators` which contains a list of all available condition creators.

#### PostgreSQL

```go
import "github.com/ory-am/ladon"
import "github.com/ory-am/ladon/postgres"
import "database/sql"
import _ "github.com/lib/pq"

func main() {
    db, err = sql.Open("postgres", "postgres://foo:bar@localhost/ladon")
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	manager = postgres.New(db, ladon.DefaultConditionCreators)
	if err = s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create schemas: %v", err)
	}

    warden := ladon.Ladon{
        Manager: manager,
    }

    // ...
}
```

### Warden

Now that we have defined our policies, we can use the warden to check if a request is valid.
`ladon.Ladon`, which is the default implementation for the `ladon.Warden` interface defines `ladon.Ladon.IsAllowed()` which
will return `nil` if the access request can be granted and an error otherwise.

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "peter",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
    }); err != nil {
        log.Fatal("Access denied")
    }

    // ...
}
```

### Examples

Let's assume that we are using the policy from above for the following requests.

#### Subject mismatch

This request will fail, because the subject "attacker" does not match `[]string{"max", "peter", "<zac|ken>"}` and since
no other policy is given, the request will be denied.

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "attacker",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
    }); err != nil { // this will be true
        log.Fatal("Access denied")
    }

    // ...
}
```

#### Owner mismatch

Although the subject "ken" matches `[]string{"max", "peter", "<zac|ken>"}` the request will fail because
ken is not the owner of `myrn:some.domain.com:resource:123` (peter is).

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "ken",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
        Context: ladon.Context{
            Owner: "peter"
        }
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

#### IP address mismatch

Although the subject "peter" matches `[]string{"max", "peter", "<zac|ken>"}` the request will fail because
the "IPMatchesCondition" is not full filled.

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "peter",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
        Context: ladon.Context{
            Owner: "peter"
        }
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

#### All good!

This request will be allowed because all requirements are met.

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "peter",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
        Context: ladon.Context{
            Owner: "peter",
            ClientIP: "127.0.0.1",
        }
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

#### Full example

```go
import "github.com/ory-am/ladon"

var pol := &ladon.DefaultPolicy{
    // A required unique identifier. Used primarily for database retrieval.
    ID: "68819e5a-738b-41ec-b03c-b58a1b19d043",

    // A optional human readable description.
    Description: "something humanly readable",

    // A subject can be an user or a service. It is the "who" in "who is allowed to do what on something".
    // As you can see here, you can use regular expressions inside < >.
    Subjects: []string{"max", "peter", "<zac|ken>"},

    // Which resources this policy affects.
    // Again, you can put regular expressions in inside < >.
    Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},

    // Which actions this policy affects. Supports RegExp
    // Again, you can put regular expressions in inside < >.
    Actions: []string{"<create|delete>", "get"},

    // Under which conditions this policy is "active".
    Conditions: Conditions{
        // In this example, the policy is only "active" when the requested subject is the owner of the resource as well.
        ConditionSubjectIsOwner{},

        // Additionally, the policy will only match if the requests remote ip address matches 127.0.0.1
        CIDRCondition{
            CIDR: "127.0.0.1/32",
        },
    }
}


func main() {
    // ...
    warden := ladon.Ladon{
        Manager: memory.New()
    }
    warden.Manager.Create(pol)

    if err := warden.IsAllowed(&ladon.Request{
        Subject: "peter",
        Action: "delete",
        Resource: "myrn:some.domain.com:resource:123",
        Context: ladon.Context{
            Owner: "peter",
            ClientIP: "127.0.0.1",
        }
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

## Good to know

* All checks are *case sensitive* because subject values could be case sensitive IDs.
* If `ladon.Ladon` is not able to match a policy with the request, it will default to denying the request and return an error.

Ladon does not use reflection for matching conditions to their appropriate structs due to security reasons.

### Useful commands

**Create mocks**
```sh
mockgen -package internal -destination internal/manager.go github.com/ory-am/ladon Manager
```
