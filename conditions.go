package ladon

// SubjectIsOwnerCondition is a condition which is fulfilled if the subject of a warden request is also the owner
// of the requested resource.
type SubjectIsOwnerCondition struct{}

func (c *SubjectIsOwnerCondition) Fulfills(r *Request) bool {
	return r.Subject == r.Context.Owner
}

// GetName returns the name of this condition.
func (c *SubjectIsOwnerCondition) GetName() string {
	return "SubjectIsOwnerCondition"
}

// SubjectIsNotOwnerCondition is a condition which is fulfilled if the subject of a warden request is not the owner.
// of the requested resource.
type SubjectIsNotOwnerCondition struct{}

func (c *SubjectIsNotOwnerCondition) Fulfills(r *Request) bool {
	return r.Subject != r.Context.Owner
}

// GetName returns the name of this condition.
func (c *SubjectIsNotOwnerCondition) GetName() string {
	return "SubjectIsNotOwnerCondition"
}
