package ladon

import (
	"encoding/json"
	"sync"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	r "gopkg.in/gorethink/gorethink.v3"
)

// stupid hack
type rdbSchema struct {
	ID          string          `json:"id" gorethink:"id"`
	Description string          `json:"description" gorethink:"description"`
	Subjects    []string        `json:"subjects" gorethink:"subjects"`
	Effect      string          `json:"effect" gorethink:"effect"`
	Resources   []string        `json:"resources" gorethink:"resources"`
	Actions     []string        `json:"actions" gorethink:"actions"`
	Conditions  json.RawMessage `json:"conditions" gorethink:"conditions"`
}

func rdbToPolicy(s *rdbSchema) (*DefaultPolicy, error) {
	if s == nil {
		return nil, nil
	}

	ret := &DefaultPolicy{
		ID:          s.ID,
		Description: s.Description,
		Subjects:    s.Subjects,
		Effect:      s.Effect,
		Resources:   s.Resources,
		Actions:     s.Actions,
		Conditions:  Conditions{},
	}

	if err := ret.Conditions.UnmarshalJSON(s.Conditions); err != nil {
		return nil, errors.WithStack(err)
	}

	return ret, nil

}

func rdbFromPolicy(p Policy) (*rdbSchema, error) {
	cs, err := p.GetConditions().MarshalJSON()
	if err != nil {
		return nil, err
	}
	return &rdbSchema{
		ID:          p.GetID(),
		Description: p.GetDescription(),
		Subjects:    p.GetSubjects(),
		Effect:      p.GetEffect(),
		Resources:   p.GetResources(),
		Actions:     p.GetActions(),
		Conditions:  cs,
	}, err
}

// NewRethinkManager initializes a new RethinkManager for given session, table name defaults
// to "policies".
func NewRethinkManager(session *r.Session, table string) *RethinkManager {
	if table == "" {
		table = "policies"
	}

	policies := make(map[string]Policy)

	return &RethinkManager{
		Session:  session,
		Table:    r.Table(table),
		Policies: policies,
	}
}

// RethinkManager is a rethinkdb implementation of Manager to store policies persistently.
type RethinkManager struct {
	Session *r.Session
	Table   r.Term
	sync.RWMutex

	Policies map[string]Policy
}

// ColdStart loads all policies from rethinkdb into memory.
func (m *RethinkManager) ColdStart() error {
	m.Policies = map[string]Policy{}
	policies, err := m.Table.Run(m.Session)
	if err != nil {
		return errors.WithStack(err)
	}

	m.Lock()
	defer m.Unlock()
	var tbl rdbSchema
	for policies.Next(&tbl) {
		policy, err := rdbToPolicy(&tbl)
		if err != nil {
			return err
		}
		m.Policies[tbl.ID] = policy
	}

	return nil
}

// Create inserts a new policy.
func (m *RethinkManager) Create(policy Policy) error {
	if err := m.publishCreate(policy); err != nil {
		return err
	}

	return nil
}

// Get retrieves a policy.
func (m *RethinkManager) Get(id string) (Policy, error) {
	m.RLock()
	defer m.RUnlock()

	p, ok := m.Policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *RethinkManager) Delete(id string) error {
	if err := m.publishDelete(id); err != nil {
		return err
	}

	return nil
}

// FindPoliciesForSubject returns Policies (an array of policy) for a given subject.
func (m *RethinkManager) FindPoliciesForSubject(subject string) (Policies, error) {
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

func (m *RethinkManager) fetch() error {
	m.Policies = map[string]Policy{}
	policies, err := m.Table.Run(m.Session)
	if err != nil {
		return errors.WithStack(err)
	}

	var policy DefaultPolicy
	m.Lock()
	defer m.Unlock()
	for policies.Next(&policy) {
		m.Policies[policy.ID] = &policy
	}

	return nil
}

func (m *RethinkManager) publishCreate(policy Policy) error {
	p, err := rdbFromPolicy(policy)
	if err != nil {
		return err
	}
	if _, err := m.Table.Insert(p).RunWrite(m.Session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *RethinkManager) publishDelete(id string) error {
	if _, err := m.Table.Get(id).Delete().RunWrite(m.Session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Watch is used to watch for changes on rethinkdb (which happens
// asynchronous) and updates manager's policy accordingly.
func (m *RethinkManager) Watch(ctx context.Context) {
	go retry(time.Second*15, time.Minute, func() error {
		policies, err := m.Table.Changes().Run(m.Session)
		if err != nil {
			return errors.WithStack(err)
		}

		defer policies.Close()
		var update = make(map[string]*rdbSchema)
		for policies.Next(&update) {
			logrus.Debug("Received update from RethinkDB Cluster in policy manager.")
			newVal, err := rdbToPolicy(update["new_val"])
			if err != nil {
				logrus.Error(err)
				continue
			}

			oldVal, err := rdbToPolicy(update["old_val"])
			if err != nil {
				logrus.Error(err)
				continue
			}

			m.Lock()
			if newVal == nil && oldVal != nil {
				delete(m.Policies, oldVal.GetID())
			} else if newVal != nil && oldVal != nil {
				delete(m.Policies, oldVal.GetID())
				m.Policies[newVal.GetID()] = newVal
			} else {
				m.Policies[newVal.GetID()] = newVal
			}
			m.Unlock()
		}

		if policies.Err() != nil {
			logrus.Error(errors.Wrap(policies.Err(), ""))
		}
		return nil
	})
}

func retry(maxWait time.Duration, failAfter time.Duration, f func() error) (err error) {
	var lastStart time.Time
	err = errors.New("Did not connect.")
	loopWait := time.Millisecond * 500
	retryStart := time.Now()
	for retryStart.Add(failAfter).After(time.Now()) {
		lastStart = time.Now()
		if err = f(); err == nil {
			return nil
		}

		if lastStart.Add(maxWait * 2).Before(time.Now()) {
			retryStart = time.Now()
		}

		logrus.Infof("Retrying in %f seconds...", loopWait.Seconds())
		time.Sleep(loopWait)
		loopWait = loopWait * time.Duration(int64(2))
		if loopWait > maxWait {
			loopWait = maxWait
		}
	}
	return err
}
