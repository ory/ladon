package ladon

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
)

// Condition either do or do not fulfill an access request.
type Condition interface {
	GetName() string
	Fulfills(*Request) bool
}

// Conditions is a collection of conditions.
type Conditions []Condition

// AddCondition adds a condition to the collection.
func (cs *Conditions) AddCondition(c Condition) {
	*cs = append(*cs, c)
}

// MarshalJSON marshals a list of conditions to json.
func (cs *Conditions) MarshalJSON() ([]byte, error) {
	out := make([]jsonCondition, len(*cs))
	for k, c := range *cs {
		raw, err := json.Marshal(c)
		if err != nil {
			return []byte{}, errors.New(err)
		}

		out[k] = jsonCondition{
			Name:    c.GetName(),
			Options: json.RawMessage(raw),
		}
	}

	fmt.Printf("%s\n", out)

	return json.Marshal(out)
}

// UnmarshalJSON unmarshals a list of conditions from json.
func (cs *Conditions) UnmarshalJSON(data []byte) error {
	var jcs []jsonCondition
	var dc Condition
	if err := json.Unmarshal(data, &jcs); err != nil {
		return errors.New(err)
	}

	for _, jc := range jcs {
		for name, c := range conditionFactories {
			if name == jc.Name {
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
	Name    string          `json:"name"`
	Options json.RawMessage `json:"options"`
}

var conditionFactories = map[string]func() Condition{
	new(SubjectIsOwnerCondition).GetName(): func() Condition {
		return new(SubjectIsOwnerCondition)
	},
	new(SubjectIsNotOwnerCondition).GetName(): func() Condition {
		return new(SubjectIsNotOwnerCondition)
	},
	new(CIDRCondition).GetName(): func() Condition {
		return new(CIDRCondition)
	},
}
