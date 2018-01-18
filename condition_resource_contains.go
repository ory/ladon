package ladon

import "strings"

// ResourceContainsCondition is fulfilled if the context matches a substring within the resource name
type ResourceContainsCondition struct{}

// Fulfills returns true if the request's resouce contains the given value string
func (c *ResourceContainsCondition) Fulfills(value interface{}, r *Request) bool {

	filter, ok := value.(map[string]interface{})
	if !ok {
		return false
	}

	valueString, ok := filter["value"].(string)
	if !ok || len(valueString) < 1 {
		return false
	}

	//If no delimiter provided default to "equals" check
	delimiterString, ok := filter["delimiter"].(string)
	if !ok || len(delimiterString) < 1 {
		delimiterString = ""
	}

	// Append delimiter to strings to prevent delim+1 being interpreted as delim+10 being present
	filterValue := delimiterString + valueString + delimiterString
	resourceString := delimiterString + r.Resource + delimiterString

	matches := strings.Contains(resourceString, filterValue)
	return matches

}

// GetName returns the condition's name.
func (c *ResourceContainsCondition) GetName() string {
	return "ResourceContainsCondition"
}
