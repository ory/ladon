package ladon

import (
	"github.com/pkg/errors"

	"github.com/ory/ladon/access"
	"github.com/ory/ladon/manager"
	"github.com/ory/ladon/policy"
)

// Ladon is an implementation of Warden.
type Ladon struct {
	Manager manager.Manager
}

// IsAllowed returns nil if subject s has permission p on resource r with context c or an error otherwise.
func (g *Ladon) IsAllowed(r *access.Request) (err error) {
	policies, err := g.Manager.FindPoliciesForSubject(r.Subject)
	if err != nil {
		return err
	}

	return g.doPoliciesAllow(r, policies)
}

func (g *Ladon) doPoliciesAllow(r *access.Request, policies []policy.Policy) (err error) {
	var allowed = false

	// Iterate through all policies
	for _, p := range policies {
		// Does the action match with one of the policies?
		if pm, err := policy.Match(p, p.GetActions(), r.Action); err != nil {
			return errors.WithStack(err)
		} else if !pm {
			// no, continue to next policy
			continue
		}

		// Does the subject match with one of the policies?
		if sm, err := policy.Match(p, p.GetSubjects(), r.Subject); err != nil {
			return err
		} else if !sm {
			// no, continue to next policy
			continue
		}

		// Does the resource match with one of the policies?
		if rm, err := policy.Match(p, p.GetResources(), r.Resource); err != nil {
			return errors.WithStack(err)
		} else if !rm {
			// no, continue to next policy
			continue
		}

		// Are the policies conditions met?
		if !g.passesConditions(p, r) {
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

func (g *Ladon) passesConditions(p policy.Policy, r *access.Request) bool {
	for key, condition := range p.GetConditions() {
		if pass := condition.Fulfills(r.Context[key], r); !pass {
			return false
		}
	}
	return true
}
