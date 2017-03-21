package manager

import "github.com/ory/ladon/policy"

//go:generate mockgen -package access_test -destination ../access/manager_mock_test.go github.com/ory/ladon/manager Manager

// Manager is responsible for managing and persisting policies.
type Manager interface {

	// Create persists the policy.
	Create(policy policy.Policy) error

	// Get retrieves a policy.
	Get(id string) (policy.Policy, error)

	// Delete removes a policy.
	Delete(id string) error

	// Finds all policies associated with the subject.
	FindPoliciesForSubject(subject string) (policy.Policies, error)
}
