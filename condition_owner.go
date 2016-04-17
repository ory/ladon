package ladon

// SubjectIsOwnerCondition is a condition which is fulfilled if the subject of a warden request is also the owner
// of the requested resource.
type SubjectIsOwnerCondition struct{}

// Fulfills returns true if the the request is fulfilled by the condition.
func (c *SubjectIsOwnerCondition) Fulfills(r *Request) bool {
	return r.Subject == r.Context.Owner
}

// GetName returns the condition's name.
func (c *SubjectIsOwnerCondition) GetName() string {
	return "SubjectIsOwnerCondition"
}

// SubjectIsNotOwnerCondition is a condition which is fulfilled if the subject of a warden request is not the owner.
// of the requested resource.
type SubjectIsNotOwnerCondition struct{}

// Fulfills returns true if the the request is fulfilled by the condition.
func (c *SubjectIsNotOwnerCondition) Fulfills(r *Request) bool {
	return r.Subject != r.Context.Owner
}

// GetName returns the condition's name.
func (c *SubjectIsNotOwnerCondition) GetName() string {
	return "SubjectIsNotOwnerCondition"
}
