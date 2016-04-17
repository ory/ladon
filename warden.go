package ladon

type Request struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Subject  string `json:"subject"`
	Context  *Context
}

// Warden is responsible for deciding if subject s can perform action a on resource r with context c.
type Warden interface {
	// IsAllowed returns nil if subject s can perform action a on resource r with context c or an error otherwise.
	//  if err := guard.IsAllowed(&Request{Resource: "article/1234", Action: "update", Subject: "peter"}); err != nil {
	//    return errors.New("Not allowed")
	//  }
	IsAllowed(r *Request) error
}
