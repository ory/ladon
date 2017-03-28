package ladon

import (
	"sync"

	"github.com/pkg/errors"
)

// MemoryManager is an in-memory (non-persistent) implementation of Manager.
type MemoryManager struct {
	Policies map[string]Policy
	sync.RWMutex
	Matcher matcher
}

// NewMemoryManager constructs and initializes new MemoryManager with no policies
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		Policies: map[string]Policy{},
	}
}

// Create a new pollicy to MemoryManager
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

// GetAll retrieves a all policy.
func (m *MemoryManager) GetAll() (Policies, error) {
	return nil, errors.New("Not implemented")
}

// Delete removes a policy.
func (m *MemoryManager) Delete(id string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.Policies, id)
	return nil
}

func (m *MemoryManager) matcher() matcher {
	if m.Matcher == nil {
		m.Matcher = DefaultMatcher
	}
	return m.Matcher
}

// FindPoliciesForSubject finds all policies associated with the subject.
func (m *MemoryManager) FindPoliciesForSubject(subject string) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := Policies{}
	for _, p := range m.Policies {
		if ok, err := m.matcher().Matches(p, p.GetSubjects(), subject); err != nil {
			return Policies{}, err
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}
