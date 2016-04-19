package ladon

// SubjectIsOwnerCondition is a condition which is fulfilled if the subject of a warden request is also the owner
// of the requested resource.
type StringMatchCondition struct {
	Matches string `json:"matches"`
}

// Fulfills returns true if the the request is fulfilled by the condition.
func (c *StringMatchCondition) Fulfills(value interface{}) bool {
	s, ok := value.(string)

	return ok && s == c.Matches
}

// GetName returns the condition's name.
func (c *StringMatchCondition) GetName() string {
	return "StringMatch"
}
