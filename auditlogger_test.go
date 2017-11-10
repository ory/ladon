package ladon_test

import (
    "testing"

    . "github.com/ory/ladon"
    . "github.com/ory/ladon/manager/memory"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockAuditLogger struct {
    mock.Mock
}

func (v *MockAuditLogger) LogRejectedAccessRequest(r *Request, p Policies, d Policies) {
    v.Called(r, p, d)
}

func (v *MockAuditLogger) LogGrantedAccessRequest(r *Request, p Policies, d Policies) {
    v.Called(r, p, d)
}

func TestAuditLogger(t *testing.T) {
    auditLogger := &MockAuditLogger{}
    auditLogger.On("LogRejectedAccessRequest", mock.Anything, mock.Anything, mock.Anything)

    warden := &Ladon{
        Manager: NewMemoryManager(),
        AuditLogger: auditLogger,
    }

    r := &Request{}
    assert.NotNil(t, warden.IsAllowed(r))

    auditLogger.AssertCalled(t, "LogRejectedAccessRequest", r, Policies{}, Policies{})
}
