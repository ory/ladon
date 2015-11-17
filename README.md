# Ladon Authorization

[![Build Status](https://travis-ci.org/ory-am/ladon.svg?branch=master)](https://travis-ci.org/ory-am/ladon)
[![Coverage Status](https://coveralls.io/repos/ory-am/ladon/badge.svg?branch=master&service=github)](https://coveralls.io/github/ory-am/ladon?branch=master)

![Ladon](https://upload.wikimedia.org/wikipedia/commons/5/5c/Reggio_calabria_museo_nazionale_mosaico_da_kaulon.jpg)

[Ladon](https://en.wikipedia.org/wiki/Ladon_%28mythology%29) is the serpent dragon protecting your resources.
A policy based authorization library written in [Go](https://golang.org). Ships with PostgreSQL for persistence.

## What is this?

Ladon is a authorization library. But you can also call it a policy administration and policy decision point. First, let's look at the policy layout:

```
{
    // This should be a unique ID. This ID is required for database retrieval.
    id: "68819e5a-738b-41ec-b03c-b58a1b19d043",

    // A human readable description. Not required
    description: "something humanly readable",

    // Which identity does this policy affect?
    // As you can see here, you can use regular expressions inside < >.
    subjects: ["max", "peter", "<zac|ken>"],

    // Should the policy allow or deny access?
    effect: "allow",

    // Which resources this policy affects.
    // Again, you can put regular expressions in inside < >.
    resources: ["urn:something:resource_a", "urn:something:resource_b", "urn:something:foo:<.+>"],

    // Which permissions this policy affects. Supports RegExp
    // Again, you can put regular expressions in inside < >.
    permission: ["<create|delete>", "get"],

    // Under which conditions this policy is active.
    conditions: [
        // Currently, only an exemplary SubjectIsOwner condition is available.
        {
            "op": "SubjectIsOwner"
        }
    ]
}
```

Easy, right? To create such a policy do:

```go
import github.com/ory-am/ladon/policy

var pol := &policy.DefaultPolicy{
    ID: "68819e5a-738b-41ec-b03c-b58a1b19d043",
    Description: "something humanly readable",
    // ...
    Conditions: []policy.Condition{
        &policy.DefaultCondition{
            Operator: "SubjectIsOwner",
        },
    },
}
```

Let's see this in action! We're assuming that the passed policy has the same values as the policy layout.

```go
import github.com/ory-am/ladon/guard

var guardian := new(guard.DefaultGuard)
granted, err := guardian.IsGranted("urn:something:resource_a", "delete", "ken", []policy.Policy{pol}, nil)
// if err != nil ...
log.Print(granted) // output: false
```

Why is the output false? If we hadn't defined the SubjectIsOwner condition, isGranted would return true. But since we did,
and we did not pass a context, the policy was not accountable for this set of properties and thus the return value is false.
Let's try it again:

```go
import github.com/ory-am/ladon/guard
import github.com/ory-am/ladon/guard/operator

var guardian := new(guard.DefaultGuard)
var ctx := operator.Context{
    Owner: "ken"
}
granted, err := guardian.IsGranted("urn:something:resource_a", "delete", "ken", []policy.Policy{pol}, nil)
// if err != nil ...
log.Print(granted) // output: true
```

Please be aware that all checks are *case sensitive* because subject values could be case sensitive IDs.
