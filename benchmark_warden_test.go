package ladon_test

import (
	"fmt"
	"github.com/ory-am/common/integration"
	"github.com/ory-am/ladon"
	"github.com/pborman/uuid"
	"log"
	"testing"
)

func benchmarkLadon(i int, b *testing.B, warden *ladon.Ladon) {
	for _, pol := range generatePolicies(i) {
		warden.Manager.Create(pol)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		warden.IsAllowed(&ladon.Request{
			Subject:  "foo",
			Action:   "bar",
			Resource: "baz",
		})
	}
}

func BenchmarkLadon(b *testing.B) {
	mysql := integration.ConnectToMySQL()
	pg := integration.ConnectToPostgres("ladon_bench")

	for _, num := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("store=memory/policies=%d", num), func(b *testing.B) {
			matcher := ladon.NewRegexpMatcher(4096)
			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: &ladon.MemoryManager{
					Policies: map[string]ladon.Policy{},
					Matcher:  matcher,
				},
				Matcher: matcher,
			})
		})

		b.Run(fmt.Sprintf("store=mysql/policies=%d", num), func(b *testing.B) {
			s := ladon.NewSQLManager(mysql, nil)
			if err := s.CreateSchemas(); err != nil {
				log.Fatalf("Could not create mysql schema: %v", err)
			}

			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: s,
				Matcher: ladon.NewRegexpMatcher(4096),
			})
		})

		b.Run(fmt.Sprintf("store=postgres/policies=%d", num), func(b *testing.B) {
			s := ladon.NewSQLManager(pg, nil)
			if err := s.CreateSchemas(); err != nil {
				log.Fatalf("Could not create mysql schema: %v", err)
			}

			benchmarkLadon(num, b, &ladon.Ladon{
				Manager: s,
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
			Subjects:  []string{"foo<.*>bar<.*>", id + "<[^sdf]+>"},
			Actions:   []string{"foobar", "foobar", "foobar", "foobar", "foobar"},
			Resources: []string{"foobar", id},
			Effect:    ladon.AllowAccess,
		}
	}
	return policies
}
