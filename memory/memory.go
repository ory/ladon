package memory

import (
	"github.com/go-errors/errors"
	"github.com/ory-am/ladon"
)

type Manager struct {
	Policies map[string]ladon.Policy
}

func New() *Manager {
	return &Manager{
		Policies: map[string]ladon.Policy{},
	}
}

func (m *Manager) Create(policy ladon.Policy) error {
	m.Policies[policy.GetID()] = policy
	return nil
}

// Get retrieves a policy.
func (m *Manager) Get(id string) (ladon.Policy, error) {
	p, ok := m.Policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *Manager) Delete(id string) error {
	delete(m.Policies, id)
	return nil
}

// Finds all policies associated with the subject.
func (m *Manager) FindPoliciesForSubject(subject string) (ladon.Policies, error) {
	ps := ladon.Policies{}
	for _, p := range m.Policies {
		if ok, err := ladon.Match(p, p.GetSubjects(), subject); err != nil {
			return ladon.Policies{}, err
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}
