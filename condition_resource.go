package ladon

import (
	"strings"
)

type ResourceFilter struct {
	Delimiter string
	Value     string
}

// ResourceCondition is fulfilled if the context matches a substring within the resource name
type ResourceCondition struct{}

// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(*ResourceFilter)

	if !ok || len(filter.Value) < 1 || len(filter.Delimiter) < 1 {
		//Default to equal
		return false
	}

	// Append delimiter to strings to prevent delim+1 being interpreted as delim+10 being present
	filterValue := filter.Value + filter.Delimiter
	resourceString := r.Resource + filter.Delimiter

	return strings.Contains(resourceString, filterValue)

}

// GetName returns the condition's name.
func (c *ResourceCondition) GetName() string {
	return "ResourceCondition"
}
