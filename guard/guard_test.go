package Guard

import (
	"fmt"
	"github.com/ory-am/ladon/policy"
	"github.com/stretchr/testify/assert"
	"testing"
)

var policies = []policy.Policy{
	&policy.DefaultPolicy{"1", "description", []string{"user"}, policy.AllowAccess, []string{"articles:[0-9]+"}, []string{"create", "update", "foo|bar"}},
	&policy.DefaultPolicy{"2", "description", []string{"user"}, policy.DenyAccess, []string{"articles:[0-9]+"}, []string{"create", "update"}},
	&policy.DefaultPolicy{"3", "description", []string{}, policy.AllowAccess, []string{"articles:[0-9]+"}, []string{"view"}},
	&policy.DefaultPolicy{"4", "description", []string{"user|anonymous"}, policy.AllowAccess, []string{"articles:[0-9]+"}, []string{"view"}},
}

type test struct {
	resource   string
	permission string
	subject    string
	policies   []policy.Policy
	expect     bool
}

var cases = []test{
	// No policy given
	test{"articles:123", "", "user", []policy.Policy{}, false},

	test{"articles:123", "create", "user", []policy.Policy{policies[0]}, true},
	test{"articles:123", "update", "user", []policy.Policy{policies[0]}, true},
	test{"articles:123", "foo", "user", []policy.Policy{policies[0]}, true},
	test{"articles:123", "bar", "user", []policy.Policy{policies[0]}, true},

	// Permission not given
	test{"articles:123", "delete", "user", []policy.Policy{policies[0]}, false},
	test{"articles:123", "create", "anonymous", []policy.Policy{policies[0]}, false},

	// Force denial
	test{"articles:123", "create", "user", []policy.Policy{policies[0], policies[1]}, false},
	test{"articles:123", "update", "user", []policy.Policy{policies[0], policies[1]}, false},
	test{"articles:123", "create", "user", []policy.Policy{policies[1], policies[0]}, false},

	// Regexp does not match
	test{"articles:", "create", "user", []policy.Policy{policies[0]}, false},
	test{"articles", "create", "user", []policy.Policy{policies[0]}, false},
	test{"articles:abc", "create", "user", []policy.Policy{policies[0]}, false},
	test{"articles:123abc", "create", "user", []policy.Policy{policies[0]}, false},
	test{"articles:123", "[foo|bar]", "user", []policy.Policy{policies[0]}, false},

	test{"articles:123", "view", "anonymous", []policy.Policy{policies[2]}, true},
	test{"articles:123", "view", "user", []policy.Policy{policies[2]}, true},

	test{"articles:123", "view", "anonymous", []policy.Policy{policies[3]}, true},
	test{"articles:123", "view", "user", []policy.Policy{policies[3]}, true},
	test{"articles:123", "view", "peter", []policy.Policy{policies[3]}, false},
}

func TestIsGranted(t *testing.T) {
	g := &Guard{}
	for k, c := range cases {
		result, err := g.IsGranted(c.resource, c.permission, c.subject, c.policies)
		assert.Nil(t, err)
		assert.Equal(t, result, c.expect, fmt.Sprintf("Failed test case %d: %v", k, c))
	}
}
