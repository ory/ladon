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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/ladon"
	"github.com/ory/ladon/manager/memory"
)

// A bunch of exemplary policies
var pols = []ladon.Policy{
	&ladon.DefaultPolicy{
		ID: "0",
		Description: `This policy allows max, peter, zac and ken to create, delete and get the listed resources,
			but only if the client ip matches and the request states that they are the owner of those resources as well.`,
		Subjects:  []string{"max", "peter", "<zac|ken>"},
		Resources: []string{"myrn:some.domain.com:resource:123", "myrn:some.domain.com:resource:345", "myrn:something:foo:<.+>"},
		Actions:   []string{"<create|delete>", "get"},
		Effect:    ladon.AllowAccess,
		Conditions: ladon.Conditions{
			"owner": &ladon.EqualsSubjectCondition{},
			"clientIP": &ladon.CIDRCondition{
				CIDR: "127.0.0.1/32",
			},
		},
	},
	&ladon.DefaultPolicy{
		ID:          "1",
		Description: "This policy allows max to update any resource",
		Subjects:    []string{"max"},
		Actions:     []string{"update"},
		Resources:   []string{"<.*>"},
		Effect:      ladon.AllowAccess,
	},
	&ladon.DefaultPolicy{
		ID:          "3",
		Description: "This policy denies max to broadcast any of the resources",
		Subjects:    []string{"max"},
		Actions:     []string{"broadcast"},
		Resources:   []string{"<.*>"},
		Effect:      ladon.DenyAccess,
	},
	&ladon.DefaultPolicy{
		ID:          "2",
		Description: "This policy denies max to broadcast any of the resources",
		Subjects:    []string{"max"},
		Actions:     []string{"random"},
		Resources:   []string{"<.*>"},
		Effect:      ladon.DenyAccess,
	},
	&ladon.DefaultPolicy{
		ID:          "4",
		Description: "This policy allows swen to update any resource except `protected` resources",
		Subjects:    []string{"swen"},
		Actions:     []string{"update"},
		Resources:   []string{"myrn:some.domain.com:resource:<(?!protected).*>"},
		Effect:      ladon.AllowAccess,
	},
	&ladon.DefaultPolicy{
		ID:          "5",
		Description: "This policy allows richard to update resources which names consists of digits only",
		Subjects:    []string{"richard"},
		Actions:     []string{"update"},
		Resources:   []string{"myrn:some.domain.com:resource:<[[:digit:]]+>"},
		Effect:      ladon.AllowAccess,
	},
}

// Some test cases
var cases = []struct {
	description   string
	accessRequest *ladon.Request
	expectErr     bool
}{
	{
		description: "should fail because no policy is matching as field clientIP does not satisfy the CIDR condition of policy 1.",
		accessRequest: &ladon.Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: ladon.Context{
				"owner":    "peter",
				"clientIP": "0.0.0.0",
			},
		},
		expectErr: true,
	},
	{
		description: "should fail because no policy is matching as the owner of the resource 123 is zac, not peter!",
		accessRequest: &ladon.Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: ladon.Context{
				"owner":    "zac",
				"clientIP": "127.0.0.1",
			},
		},
		expectErr: true,
	},
	{
		description: "should pass because policy 1 is matching and has effect allow.",
		accessRequest: &ladon.Request{
			Subject:  "peter",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: ladon.Context{
				"owner":    "peter",
				"clientIP": "127.0.0.1",
			},
		},
		expectErr: false,
	},
	{
		description: "should pass because max is allowed to update all resources.",
		accessRequest: &ladon.Request{
			Subject:  "max",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:123",
		},
		expectErr: false,
	},
	{
		description: "should pass because max is allowed to update all resource, even if none is given.",
		accessRequest: &ladon.Request{
			Subject:  "max",
			Action:   "update",
			Resource: "",
		},
		expectErr: false,
	},
	{
		description: "should fail because max is not allowed to broadcast any resource.",
		accessRequest: &ladon.Request{
			Subject:  "max",
			Action:   "broadcast",
			Resource: "myrn:some.domain.com:resource:123",
		},
		expectErr: true,
	},
	{
		description: "should fail because max is not allowed to broadcast any resource, even empty ones!",
		accessRequest: &ladon.Request{
			Subject: "max",
			Action:  "broadcast",
		},
		expectErr: true,
	},
	{
		description: "should pass because swen is allowed to update all resources except `protected` resources.",
		accessRequest: &ladon.Request{
			Subject:  "swen",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:123",
		},
		expectErr: false,
	},
	{
		description: "should fail because swen is not allowed to update `protected` resource",
		accessRequest: &ladon.Request{
			Subject:  "swen",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:protected123",
		},
		expectErr: true,
	},
	{
		description: "should fail because richard is not allowed to update a resource with alphanumeric name",
		accessRequest: &ladon.Request{
			Subject:  "richard",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:protected123",
		},
		expectErr: true,
	},
	{
		description: "should pass because richard is allowed to update a resources with a name containing digits only",
		accessRequest: &ladon.Request{
			Subject:  "richard",
			Action:   "update",
			Resource: "myrn:some.domain.com:resource:25222",
		},
		expectErr: false,
	},
}

func TestLadon(t *testing.T) {
	// Instantiate ladon with the default in-memory store.
	warden := &ladon.Ladon{Manager: memory.NewMemoryManager()}

	// Add the policies defined above to the memory manager.
	for _, pol := range pols {
		require.Nil(t, warden.Manager.Create(pol))
	}

	for i := 0; i < len(pols); i++ {
		polices, err := warden.Manager.GetAll(int64(1), int64(i))
		require.NoError(t, err)
		p, err := warden.Manager.Get(fmt.Sprintf("%d", i))
		if err == nil {
			AssertPolicyEqual(t, p, polices[0])
		}
	}

	for k, c := range cases {
		t.Run(fmt.Sprintf("case=%d-%s", k, c.description), func(t *testing.T) {

			// This is where we ask the warden if the access requests should be granted
			err := warden.IsAllowed(c.accessRequest)

			assert.Equal(t, c.expectErr, err != nil)
		})
	}

}

func TestLadonEmpty(t *testing.T) {
	// If no policy was given, the warden must return an error!
	warden := &ladon.Ladon{Manager: memory.NewMemoryManager()}
	assert.NotNil(t, warden.IsAllowed(&ladon.Request{}))
}
