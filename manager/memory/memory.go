package memory

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/ladon/manager"
	"github.com/ory/ladon/policy"
)

func init() {
	manager.DefaultManagers["memory"] = NewManager
}

// MemoryManager is an in-memory (non-persistent) implementation of Manager.
type MemoryManager struct {
	Policies map[string]policy.Policy
	sync.RWMutex
}

// NewManager constructs and initializes new MemoryManager with no policies
func NewManager(opts ...manager.Option) (manager.Manager, error) {
	var o manager.Options
	for _, opt := range opts {
		opt(&o)
	}

	policies := make(map[string]policy.Policy, len(o.Policies))
	for _, policy := range o.Policies {
		policies[policy.GetID()] = policy
	}
	return &MemoryManager{Policies: policies}, nil
}

// Create a new pollicy to MemoryManager
func (m *MemoryManager) Create(policy policy.Policy) error {
	m.Lock()
	defer m.Unlock()
	if _, found := m.Policies[policy.GetID()]; found {
		return errors.New("Policy exists")
	}

	m.Policies[policy.GetID()] = policy
	return nil
}

// Get retrieves a policy.
func (m *MemoryManager) Get(id string) (policy.Policy, error) {
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

// FindPoliciesForSubject finds all policies associated with the subject.
func (m *MemoryManager) FindPoliciesForSubject(subject string) (policy.Policies, error) {
	m.RLock()
	defer m.RUnlock()
	var ps policy.Policies
	for _, p := range m.Policies {
		if ok, err := policy.Match(p, p.GetSubjects(), subject); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}
