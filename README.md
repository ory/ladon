# Ladon Authorization

[![Build Status](https://travis-ci.org/ory-am/ladon.svg?branch=master)](https://travis-ci.org/ory-am/ladon)
[![Coverage Status](https://coveralls.io/repos/ory-am/ladon/badge.svg?branch=master&service=github)](https://coveralls.io/github/ory-am/ladon?branch=master)

![Ladon](https://upload.wikimedia.org/wikipedia/commons/5/5c/Reggio_calabria_museo_nazionale_mosaico_da_kaulon.jpg)

[Ladon](https://en.wikipedia.org/wiki/Ladon_%28mythology%29) is the serpent dragon protecting your resources.
A policy based authorization library written in [Go](https://golang.org). Ships with PostgreSQL for persistence.

### Policies

```
{
    id: "68819e5a-738b-41ec-b03c-b58a1b19d043",
    description: "something humanly readable",

    // Who this policy affects. Supports RegExp
    subjects: ["max", "peter", "zac|ken"],

    // Can be allow or deny
    effect: "allow",

    // Which resources this policy affects. Supports RegExp
    resources: ["urn:something:resource_a", "urn:something:resource_b", "urn:something:foo:.+"],

    // Which permissions this policy affects. Supports RegExp
    permission: ["create|delete", "get"],

    // Under which conditions this policy is active.
    conditions: [
        {
            "op": "SubjectIsOwner"
        }
    ]
}
```
