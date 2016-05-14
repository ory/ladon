package ladon_test

import (
	"testing"

	"github.com/ory-am/ladon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pols = []ladon.Policy{
	&ladon.DefaultPolicy{
		ID:          "68819e5a-738b-41ec-b03c-b58a1b19d043",
		Description: "something humanly readable",
		Subjects:    []string{"max", "peter", "<zac|ken>"},
		Resources:   []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},
		Actions:     []string{"<create|delete>", "get"},
		Effect:      ladon.AllowAccess,
		Conditions: ladon.Conditions{
			"owner": &ladon.EqualsSubjectCondition{},
			"clientIP": &ladon.CIDRCondition{
				CIDR: "127.0.0.1/32",
			},
		},
	},
	&ladon.DefaultPolicy{
		ID:        "38819e5a-738b-41ec-b03c-b58a1b19d041",
		Subjects:  []string{"max"},
		Actions:   []string{"update"},
		Resources: []string{"<.*>"},
		Effect:    ladon.AllowAccess,
	},
	&ladon.DefaultPolicy{
		ID:        "38919e5a-738b-41ec-b03c-b58a1b19d041",
		Subjects:  []string{"max"},
		Actions:   []string{"broadcast"},
		Resources: []string{"<.*>"},
		Effect:    ladon.DenyAccess,
	},
}

func TestLadon(t *testing.T) {
	warden := &ladon.Ladon{
		Manager: memory.NewMemoryManager(),
	}
	for _, pol := range pols {
		require.Nil(t, warden.Manager.Create(pol))
	}

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

				Context: ladon.Context{
					"owner":    "peter",
					"clientIP": "0.0.0.0",
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

				Context: ladon.Context{
					"owner":    "zac",
					"clientIP": "127.0.0.1",
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

				Context: ladon.Context{
					"owner":    "peter",
					"clientIP": "127.0.0.1",
				},
			},
			expectErr: false,
		},
		{
			d: "should pass because max is allowed to update all resources",
			r: &ladon.Request{
				Subject:  "max",
				Action:   "update",
				Resource: "myrn:some.domain.com:resource:123",
			},
			expectErr: false,
		},
		{
			d: "should pass because max is allowed to update all resource, even if none is given",
			r: &ladon.Request{
				Subject:  "max",
				Action:   "update",
				Resource: "",
			},
			expectErr: false,
		},
		{
			d: "should be rejected",
			r: &ladon.Request{
				Subject:  "max",
				Action:   "broadcast",
				Resource: "myrn:some.domain.com:resource:123",
			},
			expectErr: true,
		},
		{
			d: "should be rejected",
			r: &ladon.Request{
				Subject: "max",
				Action:  "broadcast",
			},
			expectErr: true,
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
	warden := &ladon.Ladon{Manager: ladon.NewMemoryManager()}
	assert.NotNil(t, warden.IsAllowed(&ladon.Request{}))
}
