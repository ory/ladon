package ladon

import (
	"github.com/pkg/errors"
)

// Ladon is an implementation of Warden.
type Ladon struct {
	Manager Manager
	Matcher matcher
}

func (l *Ladon) matcher() matcher {
	if l.Matcher == nil {
		l.Matcher = DefaultMatcher
	}
	return l.Matcher
}

// IsAllowed returns nil if subject s has permission p on resource r with context c or an error otherwise.
func (l *Ladon) IsAllowed(r *Request) (err error) {
	policies, err := l.Manager.FindRequestCandidates(r)
	if err != nil {
		return err
	}

	// Although the manager is responsible of matching the policies, it might decide to just scan for
	// subjects, it might return all policies, or it might have a different pattern matching than Golang.
	// Thus, we need to make sure that we actually matched the right policies.
	return l.doPoliciesAllow(r, policies)
}

func (l *Ladon) doPoliciesAllow(r *Request, policies []Policy) (err error) {
	var allowed = false

	// Iterate through all policies
	for _, p := range policies {
		// Does the action match with one of the policies?
		// This is the first check because usually actions are a superset of get|update|delete|set
		// and thus match faster.
		if pm, err := l.matcher().Matches(p, p.GetActions(), r.Action); err != nil {
			return errors.WithStack(err)
		} else if !pm {
			// no, continue to next policy
			continue
		}

		// Does the subject match with one of the policies?
		// There are usually less subjects than resources which is why this is checked
		// before checking for resources.
		if sm, err := l.matcher().Matches(p, p.GetSubjects(), r.Subject); err != nil {
			return err
		} else if !sm {
			// no, continue to next policy
			continue
		}

		// Does the resource match with one of the policies?
		if rm, err := l.matcher().Matches(p, p.GetResources(), r.Resource); err != nil {
			return errors.WithStack(err)
		} else if !rm {
			// no, continue to next policy
			continue
		}

		// Are the policies conditions met?
		// This is checked first because it usually has a small complexity.
		if !l.passesConditions(p, r) {
			// no, continue to next policy
			continue
		}

		// Is the policies effect deny? If yes, this overrides all allow policies -> access denied.
		if !p.AllowAccess() {
			return errors.WithStack(ErrRequestForcefullyDenied)
		}
		allowed = true
	}

	if !allowed {
		return errors.WithStack(ErrRequestDenied)
	}

	return nil
}

func (l *Ladon) passesConditions(p Policy, r *Request) bool {
	for key, condition := range p.GetConditions() {
		if pass := condition.Fulfills(r.Context[key], r); !pass {
			return false
		}
	}
	return true
}
