package ladon

// SubjectIsOwnerCondition is a condition which is fulfilled if the subject of a warden request is also the owner
// of the requested resource.
type EqualsSubjectCondition struct {
	Equals string `json:"equals"`
}

// Fulfills returns true if the the request is fulfilled by the condition.
func (c *EqualsSubjectCondition) Fulfills(value interface{}, r *Request) bool {
	s, ok := value.(string)

	return ok && s == c.Equals
}

// GetName returns the condition's name.
func (c *EqualsSubjectCondition) GetName() string {
	return "StringMatch"
}
