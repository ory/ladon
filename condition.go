package ladon

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Condition either do or do not fulfill an access request.
type Condition interface {
	// GetName returns the condition's name.
	GetName() string

	// Fulfills returns true if the request is fulfilled by the condition.
	Fulfills(interface{}, *Request) bool
}

// Conditions is a collection of conditions.
type Conditions map[string]Condition

// AddCondition adds a condition to the collection.
func (cs Conditions) AddCondition(key string, c Condition) {
	cs[key] = c
}

// MarshalJSON marshals a list of conditions to json.
func (cs Conditions) MarshalJSON() ([]byte, error) {
	out := make(map[string]*jsonCondition, len(cs))
	for k, c := range cs {
		raw, err := json.Marshal(c)
		if err != nil {
			return []byte{}, errors.WithStack(err)
		}

		out[k] = &jsonCondition{
			Type:    c.GetName(),
			Options: json.RawMessage(raw),
		}
	}

	return json.Marshal(out)
}

// UnmarshalJSON unmarshals a list of conditions from json.
func (cs Conditions) UnmarshalJSON(data []byte) error {
	if cs == nil {
		return errors.New("Can not be nil")
	}

	var jcs map[string]jsonCondition
	var dc Condition

	if err := json.Unmarshal(data, &jcs); err != nil {
		return errors.WithStack(err)
	}

	for k, jc := range jcs {
		var found bool
		for name, c := range ConditionFactories {
			if name == jc.Type {
				found = true
				dc = c()

				if len(jc.Options) == 0 {
					cs[k] = dc
					break
				}

				if err := json.Unmarshal(jc.Options, dc); err != nil {
					return errors.WithStack(err)
				}

				cs[k] = dc
				break
			}
		}

		if !found {
			return errors.Errorf("Could not find condition type %s", jc.Type)
		}
	}

	return nil
}

type jsonCondition struct {
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
}

// ConditionFactories is where you can add custom conditions
var ConditionFactories = map[string]func() Condition{
	new(StringEqualCondition).GetName(): func() Condition {
		return new(StringEqualCondition)
	},
	new(CIDRCondition).GetName(): func() Condition {
		return new(CIDRCondition)
	},
	new(EqualsSubjectCondition).GetName(): func() Condition {
		return new(EqualsSubjectCondition)
	},
	new(StringPairsEqualCondition).GetName(): func() Condition {
		return new(StringPairsEqualCondition)
	},
}
