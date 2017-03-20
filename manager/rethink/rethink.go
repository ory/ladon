package rethink

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	r "gopkg.in/dancannon/gorethink.v2"

	"github.com/ory/ladon/log"
	"github.com/ory/ladon/manager"
	"github.com/ory/ladon/policy"
)

func init() {
	manager.DefaultManagers["rethinkdb"] = NewManager
}

// RethinkManager is a rethinkdb implementation of Manager to store policies persistently.
type RethinkManager struct {
	mu       sync.RWMutex
	session  *r.Session
	table    r.Term
	policies map[string]policy.Policy
}

// NewManager initializes a new RethinkManager for given session, table name defaults
// to "policies".
func NewManager(opts ...manager.Option) (manager.Manager, error) {
	var o manager.Options
	for _, opt := range opts {
		opt(&o)
	}
	table := o.PolicyTable
	if o.PolicyTable == "" {
		table = "policies"
	}

	// Connect to Rethink and create table if it does not exist
	session, err := getSession(o)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if _, err := r.TableCreate(table).RunWrite(session); err != nil {
		return nil, errors.WithStack(err)
	}

	mgr := &RethinkManager{
		session: session,
		table:   r.Table(table),
	}
	if err := mgr.loadToMem(); err != nil {
		return nil, errors.WithStack(err)
	}

	return mgr, nil
}

func getSession(o manager.Options) (*r.Session, error) {
	// Check if a session was already initialized and stored
	if conn, ok := o.GetConnection(); ok {
		switch t := conn.(type) {
		default:
			err := fmt.Sprintf("Expected Connection option of type %T, got %T",
				&r.Session{}, t)
			return nil, errors.New(err)
		case *r.Session:
			return t, nil
		}
	}

	// Set default options
	addrs := strings.Split(o.Connection, ",")
	if len(addrs) == 0 {
		addrs = []string{"localhost:28015"}
	}
	db := o.Database
	if db == "" {
		db = "ladon"
	}

	// Connect to database
	connOpts := r.ConnectOpts{Addresses: addrs}
	session, err := r.Connect(connOpts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create DB if it does not exist
	if _, err := r.DBCreate(db).RunWrite(session); err != nil {
		return nil, errors.WithStack(err)
	}

	// Select database and return session
	session.Use(db)
	return session, nil
}

func (m *RethinkManager) loadToMem() error {
	if m.policies != nil {
		return nil
	}

	policies, err := m.table.Run(m.session)
	if err != nil {
		return errors.WithStack(err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies = make(map[string]policy.Policy)
	var tbl rdbSchema
	for policies.Next(&tbl) {
		policy, err := rdbToPolicy(&tbl)
		if err != nil {
			m.policies = nil
			return errors.WithStack(err)
		}
		m.policies[tbl.ID] = policy
	}

	return nil
}

// Create inserts a new policy.
func (m *RethinkManager) Create(policy policy.Policy) error {
	if err := m.publishCreate(policy); err != nil {
		return err
	}

	return nil
}

// Get retrieves a policy.
func (m *RethinkManager) Get(id string) (policy.Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.policies[id]
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
func (m *RethinkManager) FindPoliciesForSubject(subject string) (policy.Policies, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ps policy.Policies
	for _, p := range m.policies {
		if ok, err := policy.Match(p, p.GetSubjects(), subject); err != nil {
			return nil, errors.WithStack(err)
		} else if !ok {
			continue
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func (m *RethinkManager) fetch() error {
	policies, err := m.table.Run(m.session)
	if err != nil {
		return errors.WithStack(err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var pol policy.DefaultPolicy
	m.policies = make(map[string]policy.Policy)
	for policies.Next(&pol) {
		m.policies[pol.GetID()] = &pol
	}

	return nil
}

func (m *RethinkManager) publishCreate(policy policy.Policy) error {
	p, err := rdbFromPolicy(policy)
	if err != nil {
		return err
	}
	if _, err := m.table.Insert(p).RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *RethinkManager) publishDelete(id string) error {
	if _, err := m.table.Get(id).Delete().RunWrite(m.session); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Watch is used to watch for changes on rethinkdb (which happens
// asynchronous) and updates manager's policy accordingly.
func (m *RethinkManager) Watch(ctx context.Context) {
	go retry(time.Second*15, time.Minute, func() error {
		policies, err := m.table.Changes().Run(m.session)
		if err != nil {
			return errors.WithStack(err)
		}

		defer policies.Close()
		update := make(map[string]*rdbSchema)
		for policies.Next(&update) {
			log.Debug("Received update from RethinkDB Cluster in policy manager.")
			newVal, err := rdbToPolicy(update["new_val"])
			if err != nil {
				log.Printf("error updating RethinkDB policy data: %v", err)
				continue
			}

			oldVal, err := rdbToPolicy(update["old_val"])
			if err != nil {
				log.Printf("error updating RethinkDB policy data: %v", err)
				continue
			}

			m.mu.Lock()
			if newVal == nil && oldVal != nil {
				delete(m.policies, oldVal.GetID())
			} else if newVal != nil && oldVal != nil {
				delete(m.policies, oldVal.GetID())
				m.policies[newVal.GetID()] = newVal
			} else {
				m.policies[newVal.GetID()] = newVal
			}
			m.mu.Unlock()
		}

		if policies.Err() != nil {
			// Errored but should not retry
			log.Print(errors.Wrap(policies.Err(), ""))
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

		log.Printf("Retrying in %f seconds...", loopWait.Seconds())
		time.Sleep(loopWait)
		loopWait = loopWait * time.Duration(int64(2))
		if loopWait > maxWait {
			loopWait = maxWait
		}
	}
	return err
}
