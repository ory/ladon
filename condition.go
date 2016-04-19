package ladon

import (
	"encoding/json"
	"fmt"

	"github.com/go-errors/errors"
)

// Condition either do or do not fulfill an access request.
type Condition interface {
	// GetName returns the condition's name.
	GetName() string

	// Fulfills returns true if the request is fulfilled by the condition.
	Fulfills(interface{}) bool
}

// Conditions is a collection of conditions.
type Conditions map[string]Condition

// AddCondition adds a condition to the collection.
func (cs *Conditions) AddCondition(c Condition) {
	*cs = append(*cs, c)
}

// MarshalJSON marshals a list of conditions to json.
func (cs *Conditions) MarshalJSON() ([]byte, error) {
	out := make(map[string]jsonCondition, len(*cs))
	for k, c := range *cs {
		raw, err := json.Marshal(c)
		if err != nil {
			return []byte{}, errors.New(err)
		}

		out[k] = jsonCondition{
			Type:    c.GetName(),
			Options: json.RawMessage(raw),
		}
	}

	return json.Marshal(out)
}

// UnmarshalJSON unmarshals a list of conditions from json.
func (cs *Conditions) UnmarshalJSON(data []byte) error {
	var jcs map[string]jsonCondition
	var dc Condition
	if err := json.Unmarshal(data, &jcs); err != nil {
		return errors.New(err)
	}

	for _, jc := range jcs {
		for name, c := range conditionFactories {
			if name == jc.Type {
				dc = c()
				if err := json.Unmarshal(jc.Options, dc); err != nil {
					return err
				}

				*cs = append(*cs, dc)
				break
			}
		}
	}

	return nil
}

type jsonCondition struct {
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
}

var conditionFactories = map[string]func() Condition{
	new(StringMatchCondition).GetName(): func() Condition {
		return new(StringMatchCondition)
	},
	new(CIDRCondition).GetName(): func() Condition {
		return new(CIDRCondition)
	},
}
