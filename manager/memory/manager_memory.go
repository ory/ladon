/*
 * Copyright © 2016-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright 	2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

package memory

import (
	"sync"

	"github.com/pkg/errors"

	. "github.com/ory/ladon"
	"github.com/ory/pagination"
	"sort"
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
	keys := make([]string, len(m.Policies))
	i := 0
	m.RLock()
	for key := range m.Policies {
		keys[i] = key
		i++
	}

	start, end := pagination.Index(int(limit), int(offset), len(m.Policies))
	sort.Strings(keys)
	ps := make(Policies, len(keys[start:end]))
	i = 0
	for _, key := range keys[start:end] {
		ps[i] = m.Policies[key]
		i++
	}
	m.RUnlock()
	return ps, nil
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

func (m *MemoryManager) findAllPolicies() (Policies, error) {
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

// FindRequestCandidates returns candidates that could match the request object. It either returns
// a set that exactly matches the request, or a superset of it. If an error occurs, it returns nil and
// the error.
func (m *MemoryManager) FindRequestCandidates(r *Request) (Policies, error) {
	return m.findAllPolicies()
}

// FindPoliciesForSubject returns policies that could match the subject. It either returns
// a set of policies that applies to the subject, or a superset of it.
// If an error occurs, it returns nil and the error.
func (m *MemoryManager) FindPoliciesForSubject(subject string) (Policies, error) {
	return m.findAllPolicies()
}

// FindPoliciesForResource returns policies that could match the resource. It either returns
// a set of policies that apply to the resource, or a superset of it.
// If an error occurs, it returns nil and the error.
func (m *MemoryManager) FindPoliciesForResource(resource string) (Policies, error) {
	return m.findAllPolicies()
}
