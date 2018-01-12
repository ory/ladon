package ladon

import (
	"regexp"
)

// ResourceContainsCondition is fulfilled if the context matches a substring within the resource name
type ResourceContainsCondition struct{}

// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceContainsCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(map[string]interface{})

	valueString := filter["Value"].(string)
	delimiterString := filter["Delimiter"].(string)

	if !ok || len(valueString) < 1 || len(delimiterString) < 1 {
		return false
	}

	// Append delimiter to strings to prevent delim+1 being interpreted as delim+10 being present
	filterValue := valueString + delimiterString
	resourceString := r.Resource + delimiterString

	matches, _ := regexp.MatchString(filterValue, resourceString)
	return matches

}

// GetName returns the condition's name.
func (c *ResourceContainsCondition) GetName() string {
	return "ResourceContainsCondition"
}
