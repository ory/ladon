package ladon

type item struct {
	addr    string
	cidr    string
	matches bool
}

var testdata = []item{
	item{"192.168.1.67", "192.168.1.0/24", true},
	item{"192.168.1.67", "192.168.1.0/28", false},
	item{"192.168.1.67", "0.0.0.0/0", true},
}

func TestCIDRMatch(t *testing.T) {
	for _, it := range testdata {
		_, cidrnet, err := net.ParseCIDR(it.cidr)
		if err != nil {
			panic(err) // assuming I did it right above
		}
		myaddr := net.ParseIP(it.addr)
		if cidrnet.Contains(myaddr) != it.matches {
			t.Fatalf("Wrong on %+v")
		}
	}
}
