package ladon

//
type SubjectIsOwnerCondition struct{}

func (c *SubjectIsOwnerCondition) FullFills(r *Request) bool {
	return r.Subject == r.Context.Owner
}

func (c *SubjectIsOwnerCondition) GetName() string {
	return "SubjectIsOwnerCondition"
}

//
type SubjectIsNotOwnerCondition struct{}

func (c *SubjectIsNotOwnerCondition) FullFills(r *Request) bool {
	return r.Subject != r.Context.Owner
}

func (c *SubjectIsNotOwnerCondition) GetName() string {
	return "SubjectIsNotOwnerCondition"
}

//
type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

func (c *CIDRCondition) FullFills(r *Request) bool {
	return r.Context.ClientIP != c.CIDR
}

func (c *CIDRCondition) GetName() string {
	return "CIDRCondition"
}
