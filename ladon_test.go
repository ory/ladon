package ladon_test

import (
	"fmt"
	"testing"

	. "github.com/ory/ladon"
	. "github.com/ory/ladon/manager/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A bunch of exemplary policies
var pols = []Policy{
	&DefaultPolicy{
		ID: "1",
		Description: `This policy allows max, peter, zac and ken to create, delete and get the listed resources,
			but only if the client ip matches and the request states that they are the owner of those resources as well.`,
		Subjects:  []string{"max", "peter", "<zac|ken>"},
		Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},
		Actions:   []string{"<create|delete>", "get"},
		Effect:    AllowAccess,
		Conditions: Conditions{
			"owner": &EqualsSubjectCondition{},
			"clientIP": &CIDRCondition{
				CIDR: "127.0.0.1/32",
			},
		},
	},
	&DefaultPolicy{
		ID:          "2",
		Description: "This policy allows max to update any resource",
		Subjects:    []string{"max"},
		Actions:     []string{"update"},
		Resources:   []string{"<.*>"},
		Effect:      AllowAccess,
	},
	&DefaultPolicy{
		ID:          "3",
		Description: "This policy denies max to broadcast any of the resources",
		Subjects:    []string{"max"},
		Actions:     []string{"broadcast"},
		Resources:   []string{"<.*>"},
		Effect:      DenyAccess,
	},
}

// Some test cases
var cases = []struct {
	description   string
	accessRequest *Request
	expectErr     bool
}{
	{
		description: "should fail because no policy is matching as field clientIP does not satisfy the CIDR condition of policy 1.",
		accessRequest: &Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: Context{
				"owner":    "peter",
				"clientIP": "0.0.0.0",
			},
		},
		expectErr: true,
	},
	{
		description: "should fail because no policy is matching as the owner of the resource 123 is zac, not peter!",
		accessRequest: &Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: Context{
				"owner":    "zac",
				"clientIP": "127.0.0.1",
			},
		},
		expectErr: true,
	},
	{
		description: "should pass because policy 1 is matching and has effect allow.",
		accessRequest: &Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: Context{
				"owner":    "peter",
				"clientIP": "127.0.0.1",
			},
		},
		expectErr: false,
	},
	{
		description: "should pass because max is allowed to update all resources.",
		accessRequest: &Request{
			Subject:  "max",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:123",
		},
		expectErr: false,
	},
	{
		description: "should pass because max is allowed to update all resource, even if none is given.",
		accessRequest: &Request{
			Subject:  "max",
			Action:   "update",
			Resource: "",
		},
		expectErr: false,
	},
	{
		description: "should fail because max is not allowed to broadcast any resource.",
		accessRequest: &Request{
			Subject:  "max",
			Action:   "broadcast",
			Resource: "myrn:some.domain.com:resource:123",
		},
		expectErr: true,
	},
	{
		description: "should fail because max is not allowed to broadcast any resource, even empty ones!",
		accessRequest: &Request{
			Subject: "max",
			Action:  "broadcast",
		},
		expectErr: true,
	},
}

func TestLadon(t *testing.T) {
	// Instantiate ladon with the default in-memory store.
	warden := &Ladon{Manager: NewMemoryManager()}

	// Add the policies defined above to the memory manager.
	for _, pol := range pols {
		require.Nil(t, warden.Manager.Create(pol))
	}

	for k, c := range cases {
		t.Run(fmt.Sprintf("case=%d-%s", k, c.description), func(t *testing.T) {

			// This is where we ask the warden if the access requests should be granted
			err := warden.IsAllowed(c.accessRequest)

			assert.Equal(t, c.expectErr, err != nil)
		})
	}
}

func TestLadonEmpty(t *testing.T) {
	// If no policy was given, the warden must return an error!
	warden := &Ladon{Manager: NewMemoryManager()}
	assert.NotNil(t, warden.IsAllowed(&Request{}))
}
