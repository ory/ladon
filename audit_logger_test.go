package ladon_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/ory/ladon"
	. "github.com/ory/ladon/manager/memory"
	"github.com/stretchr/testify/assert"
)

func TestAuditLogger(t *testing.T) {
	var output bytes.Buffer

	warden := &Ladon{
		Manager: NewMemoryManager(),
		AuditLogger: &AuditLoggerInfo{
			Logger: log.New(&output, "", 0),
		},
	}

	warden.Manager.Create(&DefaultPolicy{
		ID:        "no-updates",
		Subjects:  []string{"<.*>"},
		Actions:   []string{"update"},
		Resources: []string{"<.*>"},
		Effect:    DenyAccess,
	})
	warden.Manager.Create(&DefaultPolicy{
		ID:        "yes-deletes",
		Subjects:  []string{"<.*>"},
		Actions:   []string{"delete"},
		Resources: []string{"<.*>"},
		Effect:    AllowAccess,
	})
	warden.Manager.Create(&DefaultPolicy{
		ID:        "no-bob",
		Subjects:  []string{"bob"},
		Actions:   []string{"delete"},
		Resources: []string{"<.*>"},
		Effect:    DenyAccess,
	})

	r := &Request{}
	assert.NotNil(t, warden.IsAllowed(r))
	assert.Equal(t, "no policy allowed access\n", output.String())

	output.Reset()

	r = &Request{
		Action: "update",
	}
	assert.NotNil(t, warden.IsAllowed(r))
	assert.Equal(t, "policy no-updates forcefully denied the access\n", output.String())

	output.Reset()

	r = &Request{
		Subject: "bob",
		Action:  "delete",
	}
	assert.NotNil(t, warden.IsAllowed(r))
	assert.Equal(t, "policies yes-deletes allow access, but policy no-bob forcefully denied it\n", output.String())

	output.Reset()

	r = &Request{
		Subject: "alice",
		Action:  "delete",
	}
	assert.Nil(t, warden.IsAllowed(r))
	assert.Equal(t, "policies yes-deletes allow access\n", output.String())
}
