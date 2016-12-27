package ladon

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Policies is an array of policies.
type Policies []Policy

// Policy represent a policy model.
type Policy interface {
	// GetID returns the policies id.
	GetID() string

	// GetDescription returns the policies description.
	GetDescription() string

	// GetSubjects returns the policies subjects.
	GetSubjects() []string

	// AllowAccess returns true if the policy effect is allow, otherwise false.
	AllowAccess() bool

	// GetEffect returns the policies effect which might be 'allow' or 'deny'.
	GetEffect() string

	// GetResources returns the policies resources.
	GetResources() []string

	// GetActions returns the policies actions.
	GetActions() []string

	// GetConditions returns the policies conditions.
	GetConditions() Conditions

	// GetStartDelimiter returns the delimiter which identifies the beginning of a regular expression.
	GetStartDelimiter() byte

	// GetEndDelimiter returns the delimiter which identifies the end of a regular expression.
	GetEndDelimiter() byte
}

// DefaultPolicy is the default implementation of the policy interface.
type DefaultPolicy struct {
	ID          string     `json:"id" gorethink:"id"`
	Description string     `json:"description" gorethink:"description"`
	Subjects    []string   `json:"subjects" gorethink:"subjects"`
	Effect      string     `json:"effect" gorethink:"effect"`
	Resources   []string   `json:"resources" gorethink:"resources"`
	Actions     []string   `json:"actions" gorethink:"actions"`
	Conditions  Conditions `json:"conditions" gorethink:"conditions"`
}

// UnmarshalJSON overwrite own policy with values of the given in policy in JSON format
func (p *DefaultPolicy) UnmarshalJSON(data []byte) error {
	var pol = struct {
		ID          string     `json:"id" gorethink:"id"`
		Description string     `json:"description" gorethink:"description"`
		Subjects    []string   `json:"subjects" gorethink:"subjects"`
		Effect      string     `json:"effect" gorethink:"effect"`
		Resources   []string   `json:"resources" gorethink:"resources"`
		Actions     []string   `json:"actions" gorethink:"actions"`
		Conditions  Conditions `json:"conditions" gorethink:"conditions"`
	}{
		Conditions: Conditions{},
	}

	if err := json.Unmarshal(data, &pol); err != nil {
		return errors.WithStack(err)
	}

	*p = *&DefaultPolicy{
		ID:          pol.ID,
		Description: pol.Description,
		Subjects:    pol.Subjects,
		Effect:      pol.Effect,
		Resources:   pol.Resources,
		Actions:     pol.Actions,
		Conditions:  pol.Conditions,
	}
	return nil
}

// GetID returns the policies id.
func (p *DefaultPolicy) GetID() string {
	return p.ID
}

// GetDescription returns the policies description.
func (p *DefaultPolicy) GetDescription() string {
	return p.Description
}

// GetSubjects returns the policies subjects.
func (p *DefaultPolicy) GetSubjects() []string {
	return p.Subjects
}

// AllowAccess returns true if the policy effect is allow, otherwise false.
func (p *DefaultPolicy) AllowAccess() bool {
	return p.Effect == AllowAccess
}

// GetEffect returns the policies effect which might be 'allow' or 'deny'.
func (p *DefaultPolicy) GetEffect() string {
	return p.Effect
}

// GetResources returns the policies resources.
func (p *DefaultPolicy) GetResources() []string {
	return p.Resources
}

// GetActions returns the policies actions.
func (p *DefaultPolicy) GetActions() []string {
	return p.Actions
}

// GetConditions returns the policies conditions.
func (p *DefaultPolicy) GetConditions() Conditions {
	return p.Conditions
}

// GetEndDelimiter returns the delimiter which identifies the end of a regular expression.
func (p *DefaultPolicy) GetEndDelimiter() byte {
	return '>'
}

// GetStartDelimiter returns the delimiter which identifies the beginning of a regular expression.
func (p *DefaultPolicy) GetStartDelimiter() byte {
	return '<'
}
