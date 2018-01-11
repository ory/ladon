package ladon

import (
	"errors"
	"strings"
)

// ResourceCondition is a condition which is fulfilled if the request's subject is equal to the given value string
type ResourceCondition struct{}

// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(string)
	if !ok {
		//Default to equal
		panic(errors.New("missing resourceFilter"))
	}

	resourceString := r.Resource + ":"
	return strings.Contains(resourceString, filter)

}

// GetName returns the condition's name.
func (c *ResourceCondition) GetName() string {
	return "ResourceCondition"
}
