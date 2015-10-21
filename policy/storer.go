package policy

const AllowAccess = "allow"
const DenyAccess = "deny"

// Storer is responsible for managing the policy storage backend.
type Storer interface {

	// Create persists a new policy in the storage backend.
	//  store.Create("123", "description", policy.AllowAccess, []{"peter", "max"}, []{"create|delete"}, []{"articles:.*", "posts:1234", "posts:12345"}
	//  store.Create("124", "description", policy.DenyAccess, []{"max"}, []{"update"}, []{"posts:*"}
	//
	// Parameters subjects, permissions and resources can be regular expressions like "create|delete". They will always have a ^ pre- and $ appended
	// to enforce proper matching. So "create|delete" becomes "^create|delete$".
	Create(id, description string, effect string, subjects, permissions, resources []string) (Policy, error)

	// Get retrieves a policy from the storage backend.
	Get(id string) (Policy, error)

	// Delete removes a policy from the storage backend.
	Delete(id string) error

	// Finds all policies associated with subject.
	FindPoliciesForSubject(subject string) ([]Policy, error)
}
