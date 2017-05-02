package ladon_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ory/ladon"
	"github.com/ory/ladon/manager/memory"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

func benchmarkLadon(i int, b *testing.B, warden *ladon.Ladon) {
	//var concurrency = 30
	//var sem = make(chan bool, concurrency)
	//
	//for _, pol := range generatePolicies(i) {
	//	sem <- true
	//	go func(pol ladon.Policy) {
	//		defer func() { <-sem }()
	//		if err := warden.Manager.Create(pol); err != nil {
	//			b.Logf("Got error from warden.Manager.Create: %s", err)
	//		}
	//	}(pol)
	//}
	//
	//for i := 0; i < cap(sem); i++ {
	//	sem <- true
	//}

	for _, pol := range generatePolicies(i) {
		if err := warden.Manager.Create(pol); err != nil {
			b.Logf("Got error from warden.Manager.Create: %s", err)
		}
	}

	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		if err = warden.IsAllowed(&ladon.Request{
			Subject:  "5",
			Action:   "bar",
			Resource: "baz",
		}); errors.Cause(err) == ladon.ErrRequestDenied || errors.Cause(err) == ladon.ErrRequestForcefullyDenied || err == nil {
		} else {
			b.Logf("Got error from warden: %s", err)
		}
	}
}

func BenchmarkLadon(b *testing.B) {
	for _, num := range []int{10, 100, 1000, 10000, 100000, 1000000} {
		b.Run(fmt.Sprintf("store=memory/policies=%d", num), func(b *testing.B) {
			matcher := ladon.NewRegexpMatcher(4096)
			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: memory.NewMemoryManager(),
				Matcher: matcher,
			})
		})

		b.Run(fmt.Sprintf("store=mysql/policies=%d", num), func(b *testing.B) {
			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: managers["mysql"],
				Matcher: ladon.NewRegexpMatcher(4096),
			})
		})

		b.Run(fmt.Sprintf("store=postgres/policies=%d", num), func(b *testing.B) {
			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: managers["postgres"],
				Matcher: ladon.NewRegexpMatcher(4096),
			})
		})
	}
}

func generatePolicies(n int) map[string]ladon.Policy {
	policies := map[string]ladon.Policy{}
	for i := 0; i <= n; i++ {
		id := uuid.New()
		policies[id] = &ladon.DefaultPolicy{
			ID:        id,
			Subjects:  []string{"foobar", "some-resource" + fmt.Sprintf("%d", i%100), strconv.Itoa(i)},
			Actions:   []string{"foobar", "foobar", "foobar", "foobar", "foobar"},
			Resources: []string{"foobar", id},
			Effect:    ladon.AllowAccess,
		}
	}
	return policies
}
