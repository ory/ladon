package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pols = []Policy{
	&DefaultPolicy{
		ID:          "68819e5a-738b-41ec-b03c-b58a1b19d043",
		Description: "something humanly readable",
		Subjects:    []string{"max", "peter", "<zac|ken>"},
		Resources:   []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},
		Actions:     []string{"<create|delete>", "get"},
		Effect:      AllowAccess,
		Conditions: Conditions{
			"owner": &EqualsSubjectCondition{},
			"clientIP": &CIDRCondition{
				CIDR: "127.0.0.1/32",
			},
		},
	},
	&DefaultPolicy{
		ID:        "38819e5a-738b-41ec-b03c-b58a1b19d041",
		Subjects:  []string{"max"},
		Actions:   []string{"update"},
		Resources: []string{"<.*>"},
		Effect:    AllowAccess,
	},
	&DefaultPolicy{
		ID:        "38919e5a-738b-41ec-b03c-b58a1b19d041",
		Subjects:  []string{"max"},
		Actions:   []string{"broadcast"},
		Resources: []string{"<.*>"},
		Effect:    DenyAccess,
	},
}

func TestLadon(t *testing.T) {
	warden := &Ladon{
		Manager: NewMemoryManager(),
	}
	for _, pol := range pols {
		require.Nil(t, warden.Manager.Create(pol))
	}

	for k, c := range []struct {
		d         string
		r         *Request
		expectErr bool
	}{
		{
			d: "should fail because client ip mismatch",
			r: &Request{
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
			d: "should fail because subject is not owner",
			r: &Request{
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
			d: "should pass because policy is satisfied",
			r: &Request{
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
			d: "should pass because max is allowed to update all resources",
			r: &Request{
				Subject:  "max",
				Action:   "update",
				Resource: "myrn:some.domain.com:resource:123",
			},
			expectErr: false,
		},
		{
			d: "should pass because max is allowed to update all resource, even if none is given",
			r: &Request{
				Subject:  "max",
				Action:   "update",
				Resource: "",
			},
			expectErr: false,
		},
		{
			d: "should be rejected",
			r: &Request{
				Subject:  "max",
				Action:   "broadcast",
				Resource: "myrn:some.domain.com:resource:123",
			},
			expectErr: true,
		},
		{
			d: "should be rejected",
			r: &Request{
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
	warden := &Ladon{Manager: NewMemoryManager()}
	assert.NotNil(t, warden.IsAllowed(&Request{}))
}
