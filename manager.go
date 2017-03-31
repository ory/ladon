package ladon

// Manager is responsible for managing and persisting policies.
type Manager interface {

	// Create persists the policy.
	Create(policy Policy) error

	// Get retrieves a policy.
	Get(id string) (Policy, error)

	// Delete removes a policy.
	Delete(id string) error

	// Matches a request to policies. If no matching is supported by the manager, it should return all policies.
	MatchRequest(r *Request) (Policies, error)
}
