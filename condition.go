package ladon

import "encoding/json"

type Condition interface {
	GetName() string
	FullFills(*Request) bool
}

type Conditions []Condition

func (cs *Conditions) AddCondition(c Condition) {
	*cs = append(*cs, c)
}

func (cs *Conditions) MarshalJSON() ([]byte, error) {
	out := make(map[string]Condition)
	for _, c := range cs {
		out[c.GetName()] = cs
	}
	return json.Marshal(out)
}
