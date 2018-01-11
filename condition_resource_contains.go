package ladon

import (
	"regexp"
)

type ResourceFilter struct {
	Delimiter string
	Value     string
}

// ResourceContainsCondition is fulfilled if the context matches a substring within the resource name
type ResourceContainsCondition struct{}


// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceContainsCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(*ResourceFilter)

	if !ok || len(filter.Value) < 1 || len(filter.Delimiter) < 1 {
		//Default to equal
		return false
	}

	// Append delimiter to strings to prevent delim+1 being interpreted as delim+10 being present
	filterValue := filter.Value + filter.Delimiter
	resourceString := r.Resource + filter.Delimiter

	matches, _ := regexp.MatchString(filterValue, resourceString)
	return matches

}

// GetName returns the condition's name.
func (c *ResourceContainsCondition) GetName() string {
	return "ResourceContainsCondition"
}
