package ladon

import (
	"regexp"
)

// StringMatchCondition is a condition which is fulfilled if the given
// string value matches the regex pattern specified in StringMatchCondition
type StringMatchCondition struct {
	Matches string `json:"matches"`
}

// Fulfills returns true if the given value is a string and matches the regex
// pattern in StringMatchCondition.Matches
func (c *StringMatchCondition) Fulfills(value interface{}, _ *Request) bool {
	s, ok := value.(string)

	matches, _ := regexp.MatchString(c.Matches, s)

	return ok && matches
}

// GetName returns the condition's name.
func (c *StringMatchCondition) GetName() string {
	return "StringMatchCondition"
}
