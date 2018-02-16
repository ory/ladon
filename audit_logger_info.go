package ladon

import (
	"log"
	"os"
	"strings"
)

// AuditLoggerInfo outputs information about granting or rejecting policies.
type AuditLoggerInfo struct {
	Logger *log.Logger
}

func (a *AuditLoggerInfo) logger() *log.Logger {
	if a.Logger == nil {
		a.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return a.Logger
}

func (a *AuditLoggerInfo) LogRejectedAccessRequest(r *Request, p Policies, d Policies) {
	if len(d) > 1 {
		allowed := joinPoliciesNames(d[0 : len(d)-1])
		denied := d[len(d)-1].GetID()
		a.logger().Printf("policies %s allow access, but policy %s forcefully denied it", allowed, denied)
	} else if len(d) == 1 {
		denied := d[len(d)-1].GetID()
		a.logger().Printf("policy %s forcefully denied the access", denied)
	} else {
		a.logger().Printf("no policy allowed access")
	}
}

func (a *AuditLoggerInfo) LogGrantedAccessRequest(r *Request, p Policies, d Policies) {
	a.logger().Printf("policies %s allow access", joinPoliciesNames(d))
}

func joinPoliciesNames(policies Policies) string {
	names := []string{}
	for _, policy := range policies {
		names = append(names, policy.GetID())
	}
	return strings.Join(names, ", ")
}
