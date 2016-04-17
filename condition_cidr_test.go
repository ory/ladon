package ladon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCIDRMatch(t *testing.T) {
	for _, c := range []struct{
		cidr string
		ip string
		pass bool
	}{
		{ip: "192.168.1.67", cidr: "192.168.1.0/24",pass: true},
		{ip: "192.168.1.67",cidr:  "192.168.1.0/28", pass:false},
		{ip: "192.168.1.67",cidr:  "0.0.0.0/0", pass: true},
	} {
		condition := &CIDRCondition{
			CIDR: c.cidr,
		}

		assert.Equal(t, c.pass, condition.Fulfills(&Request{
			Context: &Context{
				ClientIP:c.ip,
			},
		}), "%s; %s", c.ip, c.cidr)
	}
}
