package ladon_test

import (
	"testing"

	"github.com/ory-am/ladon"
	"github.com/ory-am/ladon/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		&ladon.SubjectIsOwnerCondition{},

		// Additionally, the policy will only match if the requests remote ip address matches 127.0.0.1
		&ladon.CIDRCondition{
			CIDR: "127.0.0.1/32",
		},
	},
}

func TestLadon(t *testing.T) {
	warden := &ladon.Ladon{
		Manager: memory.New(),
	}
	require.Nil(t, warden.Manager.Create(pol))

	for k, c := range []struct {
		d         string
		r         *ladon.Request
		expectErr bool
	}{
		{
			d: "should fail because client ip mismatch",
			r: &ladon.Request{
				Subject:  "peter",
				Action:   "delete",
				Resource: "myrn:some.domain.com:resource:123",

				Context: &ladon.Context{
					Owner:    "peter",
					ClientIP: "0.0.0.0",
				},
			},
			expectErr: true,
		},
		{
			d: "should fail because subject is not owner",
			r: &ladon.Request{
				Subject:  "peter",
				Action:   "delete",
				Resource: "myrn:some.domain.com:resource:123",

				Context: &ladon.Context{
					Owner:    "zac",
					ClientIP: "127.0.0.1",
				},
			},
			expectErr: true,
		},
		{
			d: "should pass because policy is satisfied",
			r: &ladon.Request{
				Subject:  "peter",
				Action:   "delete",
				Resource: "myrn:some.domain.com:resource:123",

				Context: &ladon.Context{
					Owner:    "peter",
					ClientIP: "127.0.0.1",
				},
			},
			expectErr: false,
		},
	} {
		t.Logf("Joining (%d) %s", k, c.d)
		err := warden.IsAllowed(c.r)
		assert.Equal(t, c.expectErr, err != nil, "Failed (%d) %s", k, c.d)
		if err != nil {
			t.Logf("Error (%d) %s: %s", err.Error(), k, c.d)
		}
	}
}

func TestLadonEmpty(t *testing.T) {
	warden := &ladon.Ladon{Manager: memory.New()}
	assert.NotNil(t, warden.IsAllowed(&ladon.Request{}))
}
