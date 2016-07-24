package ladon

import (
	"sync"

	"github.com/go-errors/errors"
)

// Manager is a in-memory implementation of Manager.
type MemoryManager struct {
	Policies map[string]Policy
	sync.RWMutex
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		Policies: map[string]Policy{},
	}
}

func (m *MemoryManager) Create(policy Policy) error {
	m.Lock()
	defer m.Unlock()
	if _, found := m.Policies[policy.GetID()]; found {
		return errors.New("Policy exists")
	}

	m.Policies[policy.GetID()] = policy
	return nil
}

// Get retrieves a policy.
func (m *MemoryManager) Get(id string) (Policy, error) {
	m.RLock()
	defer m.RUnlock()
	p, ok := m.Policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *MemoryManager) Delete(id string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.Policies, id)
	return nil
}

// Finds all policies associated with the subject.
func (m *MemoryManager) FindPoliciesForSubject(subject string) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := Policies{}
	for _, p := range m.Policies {
		if ok, err := Match(p, p.GetSubjects(), subject); err != nil {
			return Policies{}, err
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}
