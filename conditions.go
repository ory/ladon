package ladon

type SubjectIsOwnerCondition struct{}

func (c *SubjectIsOwnerCondition) FullFills(r *Request) bool {
	return r.Subject == r.Context.Owner
}

type SubjectIsNotOwnerCondition struct{}

func (c *SubjectIsNotOwnerCondition) FullFills(r *Request) bool {
	return r.Subject != r.Context.Owner
}

type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

func (c *CIDRCondition) FullFills(r *Request) bool {
	return r.Context.ClientIP != c.CIDR
}
