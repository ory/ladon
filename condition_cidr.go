package ladon

import (
	"net"
)

// CIDRCondition makes sure that the warden requests' IP address is in the given CIDR.
type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

// Fulfills returns true if the the request is fulfilled by the condition.
func (c *CIDRCondition) Fulfills(r *Request) bool {
	_, cidrnet, err := net.ParseCIDR(c.CIDR)
	if err != nil {
		return false
	}

	ip := net.ParseIP(r.Context.ClientIP)
	if ip == nil {
		return false
	}

	return cidrnet.Contains(ip)
}

// GetName returns the condition's name.
func (c *CIDRCondition) GetName() string {
	return "CIDRCondition"
}
