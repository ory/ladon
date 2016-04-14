package ladon

import (
	"encoding/json"

	"github.com/go-errors/errors"
)

type jsonCondition struct {
	Name    string          `json:"name"`
	Options json.RawMessage `json:"options"`
}

var conditionFactory = map[string]func() Condition{
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

type Condition interface {
	GetName() string
	FullFills(*Request) bool
}

type Conditions []Condition

func (cs *Conditions) AddCondition(c Condition) {
	*cs = append(*cs, c)
}

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
	return json.Marshal(out)
}

func (cs *Conditions) UnmarshalJSON(data []byte) error {
	var jcs []jsonCondition
	var dc Condition
	if err := json.Unmarshal(data, &jcs); err != nil {
		return errors.New(err)
	}

	for _, jc := range jcs {
		for name, c := range conditionFactory {
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
