package ladon

// Manager is responsible for managing and persisting policies.
type Manager interface {

	// Create persists the policy.
	Create(policy Policy) error

	// Get retrieves a policy.
	Get(id string) (Policy, error)

	// Delete removes a policy.
	Delete(id string) error

	// Finds all policies associated with the subject.
	FindPoliciesForSubject(subject string) (Policies, error)
}
