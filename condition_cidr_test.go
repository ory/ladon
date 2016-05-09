package ladon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCIDRMatch(t *testing.T) {
	for _, c := range []struct {
		cidr string
		ip   string
		pass bool
	}{
		{ip: "192.168.1.67", cidr: "192.168.1.0/24", pass: true},
		{ip: "192.168.1.67", cidr: "192.168.1.0/28", pass: false},
		{ip: "192.168.1.67", cidr: "1", pass: false},
		{ip: "1", cidr: "192.168.1.0/28", pass: false},
		{ip: "192.168.1.67", cidr: "0.0.0.0/0", pass: true},
	} {
		condition := &CIDRCondition{
			CIDR: c.cidr,
		}

		assert.Equal(t, c.pass, condition.Fulfills(c.ip, new(Request)), "%s; %s", c.ip, c.cidr)
	}
}
