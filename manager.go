package ladon

// Manager is responsible for managing and persisting policies.
type Manager interface {

	// Create persists the policy.
	Create(policy Policy) error

	// Update updates an existing policy.
	Update(policy Policy) error

	// Get retrieves a policy.
	Get(id string) (Policy, error)

	// Delete removes a policy.
	Delete(id string) error

	// GetAll retrieves all policies.
	GetAll(limit, offset int64) (Policies, error)

	// FindRequestCandidates returns candidates that could match the request object. It either returns
	// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
	// the error.
	FindRequestCandidates(r *Request) (Policies, error)
}
