package ladon

// AuditLoggerNoOp is the default AuditLogger, that tracks nothing.
type AuditLoggerNoOp struct{}

func (*AuditLoggerNoOp) LogRejectedAccessRequest(r *Request, p Policies, d Policies) {}
func (*AuditLoggerNoOp) LogGrantedAccessRequest(r *Request, p Policies, d Policies)  {}

var DefaultAuditLogger = &AuditLoggerNoOp{}
