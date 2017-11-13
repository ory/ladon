package ladon

// AuditLogger tracks denied and granted authorizations.
type AuditLogger interface {
	LogRejectedAccessRequest(request *Request, pool Policies, deciders Policies)
	LogGrantedAccessRequest(request *Request, pool Policies, deciders Policies)
}
