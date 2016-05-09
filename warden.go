package ladon

// Request is the warden's request object.
type Request struct {
	// Resource is the resource that access is requested to.
	Resource string `json:"resource"`

	// Action is the action that is requested on the resource.
	Action string `json:"action"`

	// Subejct is the subject that is requesting access.
	Subject string `json:"subject"`

	// Context is the request's environmental context.
	Context Context `json:"context"`
}

// Warden is responsible for deciding if subject s can perform action a on resource r with context c.
type Warden interface {
	// IsAllowed returns nil if subject s can perform action a on resource r with context c or an error otherwise.
	//  if err := guard.IsAllowed(&Request{Resource: "article/1234", Action: "update", Subject: "peter"}); err != nil {
	//    return errors.New("Not allowed")
	//  }
	IsAllowed(r *Request) error
}
