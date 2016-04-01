package ladon_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/ory-am/ladon"
)

var guardConditions = map[string][]DefaultCondition{
	"sio":     {DefaultCondition{Operator: "SubjectIsOwner"}},
	"sino":    {DefaultCondition{Operator: "SubjectIsNotOwner"}},
	"invalid": {DefaultCondition{Operator: "invalid"}},
}

var guardPolicies = []Policy{
	&DefaultPolicy{"1", "description", []string{"zac"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"create", "update", "<foo|bar>"}, nil},
	&DefaultPolicy{"2", "description", []string{"zac"}, DenyAccess, []string{"articles:<[0-9]+>"}, []string{"create", "update"}, nil},
	&DefaultPolicy{"3", "description", []string{}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"view"}, nil},
	&DefaultPolicy{"4", "description", []string{"<zac|anonymous>"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"view"}, nil},
	&DefaultPolicy{"5", "subject is owner condition", []string{"zac"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"create"}, guardConditions["sio"]},
	&DefaultPolicy{"6", "subject is not owner condition", []string{"zac"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"create"}, guardConditions["sino"]},
	&DefaultPolicy{"7", "invalid condition", []string{"zac"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"update"}, guardConditions["invalid"]},
	&DefaultPolicy{"8", "invalid resource regex", []string{"zac"}, AllowAccess, []string{"<[0-9+>"}, []string{"update"}, nil},
	&DefaultPolicy{"9", "invalid subject regex", []string{"<[0-9+>"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"update"}, nil},
	&DefaultPolicy{"10", "invalid permission regex", []string{"zac"}, AllowAccess, []string{"articles:<[0-9]+>"}, []string{"<[0-9+>"}, nil},
}

var guardContexts = map[string]*Context{
	"zac":   {Owner: "zac"},
	"peter": {Owner: "peter"},
}

type guardTest struct {
	resource   string
	permission string
	subject    string
	policies   []Policy
	ctx        *Context
	expect     bool
}

var guardCases = []guardTest{
	// 0-5
	{"articles:123", "", "zac", []Policy{}, nil, false},
	{"articles:123", "create", "zac", []Policy{guardPolicies[0]}, nil, true},
	{"articles:123", "update", "zac", []Policy{guardPolicies[0]}, nil, true},
	{"articles:123", "foo", "zac", []Policy{guardPolicies[0]}, nil, true},
	{"articles:123", "bar", "zac", []Policy{guardPolicies[0]}, nil, true},
	{"articles:123", "delete", "zac", []Policy{guardPolicies[0]}, nil, false},

	// 6-10
	{"articles:123", "create", "anonymous", []Policy{guardPolicies[0]}, nil, false},
	{"articles:123", "create", "zac", []Policy{guardPolicies[0], guardPolicies[1]}, nil, false},
	{"articles:123", "update", "zac", []Policy{guardPolicies[0], guardPolicies[1]}, nil, false},
	{"articles:123", "create", "zac", []Policy{guardPolicies[1], guardPolicies[0]}, nil, false},
	{"articles:", "create", "zac", []Policy{guardPolicies[0]}, nil, false},

	// 11-15
	{"articles", "create", "zac", []Policy{guardPolicies[0]}, nil, false},
	{"articles:abc", "create", "zac", []Policy{guardPolicies[0]}, nil, false},
	{"articles:123abc", "create", "zac", []Policy{guardPolicies[0]}, nil, false},
	{"articles:123", "[foo|bar]", "zac", []Policy{guardPolicies[0]}, nil, false},
	{"articles:123", "view", "anonymous", []Policy{guardPolicies[2]}, nil, true},

	// 16-20
	{"articles:123", "view", "zac", []Policy{guardPolicies[2]}, nil, true},
	{"articles:123", "view", "anonymous", []Policy{guardPolicies[3]}, nil, true},
	{"articles:123", "view", "zac", []Policy{guardPolicies[3]}, nil, true},
	{"articles:123", "view", "peter", []Policy{guardPolicies[3]}, nil, false},
	{"articles:20", "create", "zac", []Policy{guardPolicies[4]}, guardContexts["zac"], true},

	// 21-25
	{"articles:21", "create", "zac", []Policy{guardPolicies[4]}, guardContexts["peter"], false},
	{"articles:22", "update", "zac", []Policy{guardPolicies[4]}, guardContexts["zac"], false},
	{"articles:23", "update", "zac", []Policy{guardPolicies[4]}, guardContexts["peter"], false},
	{"articles:24", "create", "zac", []Policy{guardPolicies[5]}, guardContexts["zac"], false},
	{"articles:25", "create", "zac", []Policy{guardPolicies[5]}, guardContexts["peter"], true},

	{"articles:26", "update", "zac", []Policy{guardPolicies[5]}, guardContexts["zac"], false},
	{"articles:27", "update", "zac", []Policy{guardPolicies[5]}, guardContexts["peter"], false},
	{"articles:28", "update", "zac", []Policy{guardPolicies[6]}, guardContexts["zac"], false},
	{"articles:29", "update", "zac", []Policy{guardPolicies[7]}, nil, false},
	{"articles:30", "update", "zac", []Policy{guardPolicies[8]}, nil, false},

	{"articles:31", "update", "zac", []Policy{guardPolicies[9]}, nil, false},
}

func TestIsGranted(t *testing.T) {
	g := &Guard{}
	for k, c := range guardCases {
		result, _ := g.IsGranted(c.resource, c.permission, c.subject, c.policies, c.ctx)
		assert.Equal(t, c.expect, result, fmt.Sprintf("Failed test case %d: %v", k, c))
	}
}

func BenchmarkIsGranted(b *testing.B) {
	b.ReportAllocs()
	g := &Guard{}
	for i := 0; i < b.N; i++ {
		_, _ = g.IsGranted("articles:30", "update", "zac", guardPolicies, guardContexts["zac"])
	}
}
