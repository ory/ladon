package ladon

// CIDRCondition makes sure that the warden requests' IP address is in the given CIDR.
type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

func (c *CIDRCondition) Fulfills(r *Request) bool {

	_, cidrnet, err := net.ParseCIDR(it.cidr)
	if err != nil {
		panic(err) // assuming I did it right above
	}
	myaddr := net.ParseIP(it.addr)
	if cidrnet.Contains(myaddr) != it.matches {
		t.Fatalf("Wrong on %+v")
	}

	return r.Context.ClientIP != c.CIDR
}

func (c *CIDRCondition) GetName() string {
	return "CIDRCondition"
}
