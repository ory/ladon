# ![Ladon](logo.png)

[![Build Status](https://travis-ci.org/ory-am/ladon.svg?branch=master)](https://travis-ci.org/ory-am/ladon)
[![Coverage Status](https://coveralls.io/repos/ory-am/ladon/badge.svg?branch=master&service=github)](https://coveralls.io/github/ory-am/ladon?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ory-am/ladon)](https://goreportcard.com/report/github.com/ory-am/ladon)

[Ladon](https://en.wikipedia.org/wiki/Ladon_%28mythology%29) is the serpent dragon protecting your resources.

Ladon is a library written in [Go](https://golang.org) for access control policies, similar to [Role Based Access Control](https://en.wikipedia.org/wiki/Role-based_access_control)
or [Access Control Lists](https://en.wikipedia.org/wiki/Access_control_list). 
In contrast to [ACL](https://en.wikipedia.org/wiki/Access_control_list) and [RBAC](https://en.wikipedia.org/wiki/Role-based_access_control)
you get fine-grained access control with the ability to answer questions in complex environments such as multi-tenant or distributed applications
and large organizations. Ladon is inspired by [AWS IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html).

Ladon ships with storage adapters for SQL (officially supported: MySQL, PostgreSQL) and RethinkDB (community supported).

**[Hydra](https://github.com/ory-am/hydra)**, an OAuth2 and OpenID Connect implementation uses Ladon for access control.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Installation](#installation)
- [Concepts](#concepts)
- [Usage](#usage)
  - [Policies](#policies)
  - [Policy management](#policy-management)
    - [In memory](#in-memory)
    - [Using a backend](#using-a-backend)
      - [MySQL](#mysql)
      - [PostgreSQL](#postgresql)
  - [Warden](#warden)
  - [Conditions](#conditions)
  - [Examples](#examples)
    - [Subject mismatch](#subject-mismatch)
    - [Owner mismatch](#owner-mismatch)
    - [IP address mismatch](#ip-address-mismatch)
    - [Working example](#working-example)
    - [Full code for working example](#full-code-for-working-example)
- [Good to know](#good-to-know)
- [Useful commands](#useful-commands)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

Ladon utilizes ory-am/dockertest for tests.
Please refer to [ory-am/dockertest](https://github.com/ory-am/dockertest) for more information of how to setup testing environment.

## Installation

```
go get github.com/ory-am/ladon
```

We recommend to use [Glide](https://github.com/Masterminds/glide) for dependency management. Ladon uses [semantic
versioning](http://semver.org/) and versions beginning with zero (`0.1.2`) might introduce backwards compatibility
breaks with [each minor version](http://semver.org/#how-should-i-deal-with-revisions-in-the-0yz-initial-development-phase).

## Concepts

Ladon is an access control library that answers the question:

> **Who** is **able** to do **what** on **something** given some **context**

* **Who** An arbitrary unique subject name, for example "ken" or "printer-service.mydomain.com".
* **Able**: The effect which can be either "allow" or "deny".
* **What**: An arbitrary action name, for example "delete", "create" or "scoped:action:something".
* **Something**: An arbitrary unique resource name, for example "something", "resources.articles.1234" or some uniform
    resource name like "urn:isbn:3827370191".
* **Context**: The current context containing information about the environment such as the IP Address,
    request date, the resource owner name, the department ken is working in or any other information you want to pass along.
    (optional)

To decide what the answer is, Ladon uses policy documents which can be represented as JSON

```json
{
  "description": "One policy to rule them all.",
  "subjects": ["users:<[peter|ken]>", "users:maria", "groups:admins"],
  "actions" : ["delete", "<[create|update]>"],
  "effect": "allow",
  "resources": [
    "resources:articles:<.*>",
    "resources:printer"
  ],
  "conditions": {
    "remoteIP": {
        "type": "CIDRCondition",
        "options": {
            "cidr": "192.168.0.1/16"
        }
    }
  }
}
```

and can answer access requests that look like:

```json
{
  "subject": "users:peter",
  "action" : "delete",
  "resource": "resource:articles:ladon-introduction",
  "context": {
    "remoteIP": "192.168.0.5"
  }
}
```

However, Ladon does not come with a HTTP or server implementation. It does not restrict JSON either. We believe that it is your job to decide
if you want to use Protobuf, RESTful, HTTP, AMPQ, or some other protocol. It's up to you to write server!

The following example should give you an idea what a RESTful flow *could* look like. Initially we create a policy by
POSTing it to an artificial HTTP endpoint:

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://my-ladon-implementation.localhost/policies" <<EOF
        {
          "description": "One policy to rule them all.",
          "subjects": ["users:<[peter|ken]>", "users:maria", "groups:admins"],
          "actions" : ["delete", "<[create|update]>"],
          "effect": "allow",
          "resources": [
            "resources:articles:<.*>",
            "resources:printer"
          ],
          "conditions": {
            "remoteIP": {
                "type": "CIDRCondition",
                "options": {
                    "cidr": "192.168.0.1/16"
                }
            }
          }
        }
  EOF
```

Then we test if "peter" (ip: "192.168.0.5") is allowed to "delete" the "ladon-introduction" article:

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://my-ladon-implementation.localhost/ladon" <<EOF
        {
          "subject": "users:peter",
          "action" : "delete",
          "resource": "resource:articles:ladon-introduction",
          "context": {
            "remoteIP": "192.168.0.5"
          }
        }
  EOF

{
    "allowed": true
}
```

## Usage

### Policies

Policies are an essential part of Ladon. Policies are documents which define who is allowed to perform an action on a resource.
Policies must implement the `ladon.Policy` interface. A standard implementation of this interface is `ladon.DefaultPolicy`:

```go
import "github.com/ory-am/ladon"

var pol = &ladon.DefaultPolicy{
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

	// Should access be allowed or denied?
	// Note: If multiple policies match an access request, ladon.DenyAccess will always override ladon.AllowAccess
	// and thus deny access.
	Effect: ladon.AllowAccess,

	// Under which conditions this policy is "active".
	Conditions: ladon.Conditions{
		// In this example, the policy is only "active" when the requested subject is the owner of the resource as well.
		"resourceOwner": &ladon.EqualsSubjectCondition{},

		// Additionally, the policy will only match if the requests remote ip address matches address range 127.0.0.1/32
		"remoteIPAddress": &ladon.CIDRCondition{
			CIDR: "127.0.0.1/32",
		},
	},
}

```

### Policy management

Ladon comes with `ladon.Manager`, a policy management interface which ships with implementations for RethinkDB, PostgreSQL and MySQL.

**A word on Condition creators**
Unmarshalling lists with multiple types is not trivial in Go. Ladon comes with creators (factories) for the different conditions.
The manager receives a list of allowed condition creators who assist him in finding and creating the right condition objects.

#### In memory

```go
import (
	"github.com/ory-am/ladon"
)


func main() {
	warden := &ladon.Ladon{
		Manager: ladon.NewMemoryManager(),
	}
	err := warden.Manager.Create(pol)

    // ...
}
```

#### Using a backend

You will notice that all persistent implementations require an additional argument when setting up. This argument
is called `allowedConditionCreators` and must contain a list of allowed condition creators (or "factories"). Because it is
not trivial to unmarshal lists of various types (required by `ladon.Conditions`), we wrote some helpers to do that for you.

You can always pass `ladon.DefaultConditionCreators` which contains a list of all available condition creators.

##### MySQL

```go
import "github.com/ory-am/ladon"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func main() {
    db, err = sql.Open("mysql", "user:pass@tcp(127.0.0.1:3306)"")
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

    warden := ladon.Ladon{
        Manager: ladon.NewSQLManager(db),
    }

    // ...
}
```

##### PostgreSQL

```go
import "github.com/ory-am/ladon"
import "database/sql"
import _ "github.com/lib/pq"

func main() {
    db, err = sql.Open("postgres", "postgres://foo:bar@localhost/ladon")
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

    warden := ladon.Ladon{
        Manager: ladon.NewSQLManager(db),
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

### Conditions

There are a couple of conditions available:

* [CIDR Condition](condition_cidr.go): Matches CIDR IP Ranges.
* [String Equal Condition](condition_string_equal.go): Matches two strings.
* [Subject Condition](condition_subject_equal.go): Matches when the condition field is equal to the subject field.

You can add custom conditions by using `ladon.ConditionFactories` (for more information see condition.go):

```go
import "github.com/ory-am/ladon"

func main() {
    // ...

    ladon.ConditionFactories[new(CustomCondition).GetName()] = func() Condition {
        return new(CustomCondition)
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
        Context: &ladon.Context{
            "resourceOwner": "peter",
        },
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
        Context: &ladon.Context{
            "resourceOwner": "peter",
        },
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

#### Working example

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
            "resourceOwner": "peter",
            "remoteIPAddress": "127.0.0.1",
        },
    }); err != nil {
        log.Print("Access denied")
    }

    // ...
}
```

#### Full code for working example

To view the example's full code, click [here](ladon_test.go). To run it, call `go test -run=TestLadon .`

## Good to know

* All checks are *case sensitive* because subject values could be case sensitive IDs.
* If `ladon.Ladon` is not able to match a policy with the request, it will default to denying the request and return an error.

Ladon does not use reflection for matching conditions to their appropriate structs due to security reasons.

## Useful commands

**Create mocks**
```sh
mockgen -package ladon_test -destination manager_mock_test.go github.com/ory-am/ladon Manager
```
