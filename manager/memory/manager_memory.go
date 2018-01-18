// Copyright © 2017 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"sync"

	. "github.com/ory/ladon"
	"github.com/ory/pagination"
	"github.com/pkg/errors"
)

// MemoryManager is an in-memory (non-persistent) implementation of Manager.
type MemoryManager struct {
	Policies map[string]Policy
	sync.RWMutex
}

// NewMemoryManager constructs and initializes new MemoryManager with no policies.
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		Policies: map[string]Policy{},
	}
}

// Update updates an existing policy.
func (m *MemoryManager) Update(policy Policy) error {
	m.Lock()
	defer m.Unlock()
	m.Policies[policy.GetID()] = policy
	return nil
}

// GetAll returns all policies.
func (m *MemoryManager) GetAll(limit, offset int64) (Policies, error) {
	ps := make(Policies, len(m.Policies))
	i := 0

	for _, p := range m.Policies {
		ps[i] = p
		i++
	}

	start, end := pagination.Index(int(limit), int(offset), len(ps))
	return ps[start:end], nil
}

// Create a new pollicy to MemoryManager.
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

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *MemoryManager) FindRequestCandidates(r *Request) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := make(Policies, len(m.Policies))
	var count int
	for _, p := range m.Policies {
		ps[count] = p
		count++
	}
	return ps, nil
}
