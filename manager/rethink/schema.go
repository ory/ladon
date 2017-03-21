package rethink

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/ladon/access"
	"github.com/ory/ladon/policy"
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

func rdbToPolicy(s *rdbSchema) (*policy.DefaultPolicy, error) {
	if s == nil {
		return nil, nil
	}

	ret := &policy.DefaultPolicy{
		ID:          s.ID,
		Description: s.Description,
		Subjects:    s.Subjects,
		Effect:      s.Effect,
		Resources:   s.Resources,
		Actions:     s.Actions,
		Conditions:  access.Conditions{},
	}

	if err := ret.Conditions.UnmarshalJSON(s.Conditions); err != nil {
		return nil, errors.WithStack(err)
	}

	return ret, nil

}

func rdbFromPolicy(p policy.Policy) (*rdbSchema, error) {
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
	}, nil
}
