/*
 * Copyright © 2016-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright 	2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

package ladon_test

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/ory/ladon"
	. "github.com/ory/ladon/manager/memory"
)

func TestAuditLogger(t *testing.T) {
	var output bytes.Buffer

	warden := &Ladon{
		Manager: NewMemoryManager(),
		AuditLogger: &AuditLoggerInfo{
			Logger: log.New(&output, "", 0),
		},
	}

	ctx := context.Background()

	warden.Manager.Create(ctx, &DefaultPolicy{
		ID:        "no-updates",
		Subjects:  []string{"<.*>"},
		Actions:   []string{"update"},
		Resources: []string{"<.*>"},
		Effect:    DenyAccess,
	})
	warden.Manager.Create(ctx, &DefaultPolicy{
		ID:        "yes-deletes",
		Subjects:  []string{"<.*>"},
		Actions:   []string{"delete"},
		Resources: []string{"<.*>"},
		Effect:    AllowAccess,
	})
	warden.Manager.Create(ctx, &DefaultPolicy{
		ID:        "no-bob",
		Subjects:  []string{"bob"},
		Actions:   []string{"delete"},
		Resources: []string{"<.*>"},
		Effect:    DenyAccess,
	})

	r := &Request{}
	assert.NotNil(t, warden.IsAllowed(ctx, r))
	assert.Equal(t, "no policy allowed access\n", output.String())

	output.Reset()

	r = &Request{
		Action: "update",
	}
	assert.NotNil(t, warden.IsAllowed(ctx, r))
	assert.Equal(t, "policy no-updates forcefully denied the access\n", output.String())

	output.Reset()

	r = &Request{
		Subject: "bob",
		Action:  "delete",
	}
	assert.NotNil(t, warden.IsAllowed(ctx, r))
	assert.Equal(t, "policies yes-deletes allow access, but policy no-bob forcefully denied it\n", output.String())

	output.Reset()

	r = &Request{
		Subject: "alice",
		Action:  "delete",
	}
	assert.Nil(t, warden.IsAllowed(ctx, r))
	assert.Equal(t, "policies yes-deletes allow access\n", output.String())
}
