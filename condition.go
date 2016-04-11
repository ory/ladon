package ladon

type Condition interface {
	FullFills(*Request) bool
}

type Conditions []Condition

func (cs *Conditions) AddCondition(c Condition) {
	*cs = append(*cs, c)
}
