package ladon

import (
	"regexp"
)

// ResourceContainsCondition is fulfilled if the context matches a substring within the resource name
type ResourceContainsCondition struct{}

// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceContainsCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(map[string]interface{})
	if !ok {
		return false
	}

	valueString, ok := filter["Value"].(string)
	if !ok || len(valueString) < 1 {
		return false
	}

	delimiterString, ok := filter["Delimiter"].(string)
	if !ok || len(delimiterString) < 1 {
		return false
	}

	// Append delimiter to strings to prevent delim+1 being interpreted as delim+10 being present
	filterValue := delimiterString + valueString + delimiterString
	resourceString := r.Resource + delimiterString

	matches, _ := regexp.MatchString(filterValue, resourceString)
	return matches

}

// GetName returns the condition's name.
func (c *ResourceContainsCondition) GetName() string {
	return "ResourceContainsCondition"
}
