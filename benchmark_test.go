package ladon_test

import (
	"fmt"
	"github.com/ory-am/ladon"
	"github.com/pborman/uuid"
	"testing"
)

func benchmarkLadon(i int, b *testing.B) {
	var warden *ladon.Ladon

	warden = &ladon.Ladon{Manager: ladon.NewMemoryManager()}
	for _, pol := range generatePolicies(i) {
		warden.Manager.Create(pol)
	}

	for n := 0; n < b.N; n++ {
		warden.IsAllowed(&ladon.Request{
			Subject:  "foo",
			Action:   "bar",
			Resource: "baz",
		})
	}
}

func BenchmarkLadon(b *testing.B) {
	for _, num := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("policies=%d", num), func(b *testing.B) {
			benchmarkLadon(num, b)
		})
	}
}

func generatePolicies(n int) map[string]ladon.Policy {
	policies := map[string]ladon.Policy{}
	for i := 0; i <= n; i++ {
		id := uuid.New()
		policies[id] = &ladon.DefaultPolicy{
			ID:        id,
			Subjects:  []string{"foo<.*>bar<.*>", id + "<[^sdf]+>"},
			Actions:   []string{"foobar", "foobar", "foobar", "foobar", "foobar"},
			Resources: []string{"foobar", id},
			Effect:    ladon.AllowAccess,
		}
	}
	return policies
}
