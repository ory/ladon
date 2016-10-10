package ladon

// StringEqualCondition is a condition which is fulfilled if the given
// string value is the same as specified in StringEqualCondition
type StringEqualCondition struct {
	Equals string `json:"equals"`
}

// Fulfills returns true if the given value is a string and is the
// same as in StringEqualCondition.Equals
func (c *StringEqualCondition) Fulfills(value interface{}, _ *Request) bool {
	s, ok := value.(string)

	return ok && s == c.Equals
}

// GetName returns the condition's name.
func (c *StringEqualCondition) GetName() string {
	return "StringEqualCondition"
}
