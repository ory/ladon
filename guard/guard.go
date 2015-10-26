package guard

import (
	"fmt"
	"github.com/ory-am/ladon/policy"
	"regexp"
)

type Guard struct{}

func (g *Guard) IsGranted(resource, permission, subject string, policies []policy.Policy) (bool, error) {
	allow := false

	// Iterate through all policies
	for _, p := range policies {
		// Does the resource match with one of the policies?
		if rm, err := matches(p.GetResources(), resource); err != nil {
			return false, err
		} else if !rm {
			continue
		}

		// Does the action match with one of the policies?
		if pm, err := matches(p.GetPermissions(), permission); err != nil {
			return false, err
		} else if !pm {
			continue
		}

		// Does the subject match with one of the policies?
		if sm, err := matches(p.GetSubjects(), subject); err != nil {
			return false, err
		} else if !sm && len(p.GetSubjects()) > 0 {
			// If no match exists, but the subjects are scoped, this policy is irrelevant
			continue
		}

		// Does the policy enforce a deny policy? If yes, this overrides all allow policies -> access denied.
		if !p.HasAccess() {
			return false, nil
		}
		allow = true
	}
	return allow, nil
}

func matches(haystack []string, needle string) (bool, error) {
	for _, h := range haystack {
		matches, err := regexp.MatchString(fmt.Sprintf("^%s$", h), needle)
		if err != nil {
			return false, err
		} else if matches {
			return true, nil
		}
	}
	return false, nil
}
