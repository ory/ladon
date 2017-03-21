package memory

import (
	"context"
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
	sync.RWMutex
	policies map[string]policy.Policy
}

type key uint16

const bootstrapKey key = 1

// BootstrapData allows for initializing the in-memory manager with some
// policies.
func BootstrapData(pols []policy.Policy) manager.Option {
	return func(o *manager.Options) {
		o.Metadata = context.WithValue(o.Metadata, bootstrapKey, pols)
	}
}

func getPolicies(o *manager.Options) map[string]policy.Policy {
	policyList, ok := o.Metadata.Value(bootstrapKey).([]policy.Policy)
	if !ok {
		return nil
	}
	policies := make(map[string]policy.Policy, len(policyList))
	for _, pol := range policyList {
		policies[pol.GetID()] = pol
	}
	return policies
}

// NewManager constructs and initializes new MemoryManager with no policies
func NewManager(opts ...manager.Option) (manager.Manager, error) {
	var o manager.Options
	for _, opt := range opts {
		opt(&o)
	}
	return &MemoryManager{policies: getPolicies(&o)}, nil
}

// Create a new policy in MemoryManager
func (m *MemoryManager) Create(policy policy.Policy) error {
	m.Lock()
	defer m.Unlock()
	if _, found := m.policies[policy.GetID()]; found {
		return errors.New("Policy exists")
	}

	m.policies[policy.GetID()] = policy
	return nil
}

// Get retrieves a policy.
func (m *MemoryManager) Get(id string) (policy.Policy, error) {
	m.RLock()
	defer m.RUnlock()
	p, ok := m.policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *MemoryManager) Delete(id string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.policies, id)
	return nil
}

// FindPoliciesForSubject finds all policies associated with the subject.
func (m *MemoryManager) FindPoliciesForSubject(subject string) (policy.Policies, error) {
	m.RLock()
	defer m.RUnlock()
	var ps policy.Policies
	for _, p := range m.policies {
		if ok, err := policy.Match(p, p.GetSubjects(), subject); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}
