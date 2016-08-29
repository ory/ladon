# ![Ladon](logo.png)

[![Build Status](https://travis-ci.org/ory-am/ladon.svg?branch=master)](https://travis-ci.org/ory-am/ladon)
[![Coverage Status](https://coveralls.io/repos/ory-am/ladon/badge.svg?branch=master&service=github)](https://coveralls.io/github/ory-am/ladon?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ory-am/ladon)](https://goreportcard.com/report/github.com/ory-am/ladon)

[Ladon](https://en.wikipedia.org/wiki/Ladon_%28mythology%29) is the serpent dragon protecting your resources.
A policy based authorization library written in [Go](https://golang.org). Ships with PostgreSQL and RethinkDB storage and utilizes ory-am/dockertest V2 for tests. Please refer to [ory-am/dockertest](https://github.com/ory-am/dockertest) for more information on how to setup testing environment.

Be aware that ladon is only a library. If you are looking for runnable server, check out **[Hydra](https://github.com/ory-am/hydra)**.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Installation](#installation)
- [What is this and how does it work?](#what-is-this-and-how-does-it-work)
  - [Ladon vs ACL](#ladon-vs-acl)
  - [Ladon vs RBAC](#ladon-vs-rbac)
- [How could Ladon work in my environment?](#how-could-ladon-work-in-my-environment)
  - [Access request without context](#access-request-without-context)
  - [Access request with resource and context](#access-request-with-resource-and-context)
- [Usage](#usage)
  - [Policies](#policies)
  - [Policy management](#policy-management)
    - [In memory](#in-memory)
    - [Using a backend](#using-a-backend)
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

We recommend to use [Glide](https://github.com/Masterminds/glide) or [Godep](https://github.com/tools/godep), because
there might be breaking changes in the future.

## What is this and how does it work?

Ladon is an access control library. You might also call it a policy administration and policy decision point. Ladon
answers the question:

> **Who** is **able** to do **what** on **something** with some **context**

* **Who** An arbitrary unique subject name, for example "ken" or "printer-service.mydomain.com".
* **Able**: The effect which is always "allow" or "deny".
* **What**: An arbitrary action name, for example "delete", "create" or "scoped:action:something".
* **Something**: An arbitrary unique resource name, for example "something", "resources:articles:1234" or some uniform
    resource name like "urn:isbn:3827370191".
* **Context**: The current context which may environment information like the IP Address,
    request date, the resource owner name, the department ken is working in and anything you like.

### Ladon vs ACL

> An access control list (ACL), with respect to a computer file system, is a list of permissions attached to an object.
  An ACL specifies which users or system processes are granted access to objects, as well as what operations are allowed on given objects.
  Each entry in a typical ACL specifies a subject and an operation. For instance, if a file object has an ACL that contains
  (Alice: read,write; Bob: read), this would give Alice permission to read and write the file and Bob to only read it.
  \- *[Source](https://en.wikipedia.org/wiki/Access_control_list)*

Compare this with Ladon and you get:

* **Who**: The ACL subject (Alice, Bob).
* **What**: The Operation or permission.
* **Something**: The object.

ACL however is a white list (Alice is granted permission read on object foo). Ladon however can be used to blacklist as well:
Alice is disallowed permission read on object foo.

Without tweaking, ACL does not support departments, ip addresses, request dates and other environmental information. Ladon does.

### Ladon vs RBAC

> In computer systems security, role-based access control (RBAC) is an approach to restricting system access to authorized users.
  RBAC is sometimes referred to as role-based security. Within an organization, roles are created for various job functions.
  The permissions to perform certain operations are assigned to specific roles. Members or staff (or other system users)
  are assigned particular roles, and through those role assignments acquire the computer permissions to perform particular
  computer-system functions. Since users are not assigned permissions directly, but only acquire them through their role (or roles),
  management of individual user rights becomes a matter of simply assigning appropriate roles to the user's account;
  this simplifies common operations, such as adding a user, or changing a user's department.
  \- *[Source](https://en.wikipedia.org/wiki/Role-based_access_control)*

Compare this with Ladon and you get:

* **Who**: The role
* **What**: The Operation or permission.

Again, RBAC is a white list. RBAC does not know objects (*something*) neither does RBAC know contexts.

## How could Ladon work in my environment?

Ladon does not come with a HTTP handler. We believe that it is your job to decide
if you want to use Protobuf, RESTful, HTTP, AMPQ, or some other protocol. It's up to you to write handlers!

The following examples will give you a better understanding of what you can do with Ladon.

### Access request without context

A valid access request and policy requires at least the affected subject, action and effect:

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://ladon.myorg.com/policies" <<EOF
      {
          "description": "One policy to rule them all.",
          "subjects": ["users:peter", "users:ken", "groups:admins"],
          "actions" : ["delete"],
          "resources": [
            "<.*>"
          ],
          "effect": "allow"
      }
  EOF
```

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://ladon.myorg.com/warden" <<EOF
      {
          "subject": "users:peter",
          "action" : "delete"
      }
  EOF

{
    "allowed": true
}
```

**Note:** Because *resources* matches everything (`.*`), it is not required to pass a resource name to the warden.

### Access request with resource and context

The next example uses resources and (context) conditions to further refine access control requests.

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://ladon.myorg.com/policies" <<EOF
      {
          "description": "One policy to rule them all.",
          "subjects": ["users:peter", "users:ken", "groups:admins"],
          "actions" : ["delete"],
          "effect": "allow",
          "resources": [
            "resource:articles<.*>"
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

```
> curl \
      -X POST \
      -H "Content-Type: application/json" \
      -d@- \
      "https://ladon.myorg.com/warden" <<EOF
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

Ladon comes with `ladon.Manager`, a policy management interface which is implemented using RethinkDB and PostgreSQL.
Storing policies.

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
        Manager: ladon.NewPostgresManager(db),
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
mockgen -package internal -destination internal/manager.go github.com/ory-am/ladon Manager
```
