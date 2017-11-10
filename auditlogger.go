package ladon

// AuditLogger tracks denied and granted authorizations.
type AuditLogger interface {
	LogRejectedAccessRequest(request *Request, pool Policies, deciders Policies)
	LogGrantedAccessRequest(request *Request, pool Policies, deciders Policies)
}

// AuditLoggerNoOp is the default AuditLogger, that tracks nothing.
type AuditLoggerNoOp struct{}

func (*AuditLoggerNoOp) LogRejectedAccessRequest(r *Request, p Policies, d Policies)  {}
func (*AuditLoggerNoOp) LogGrantedAccessRequest(r *Request, p Policies, d Policies) {}

var DefaultAuditLogger = &AuditLoggerNoOp{}
