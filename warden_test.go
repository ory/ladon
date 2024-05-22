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
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	. "github.com/ory/ladon"
)

func TestWardenIsGranted(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockManager(ctrl)
	defer ctrl.Finish()

	w := &Ladon{
		Manager: m,
	}

	ctx := context.Background()

	for k, c := range []struct {
		r           *Request
		description string
		setup       func()
		expectErr   bool
	}{
		{
			description: "should fail because no policies are found for peter",
			r:           &Request{Subject: "peter"},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{Subject: "peter"})).Return(Policies{}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because lookup failure when accessing policies for peter",
			r:           &Request{Subject: "peter"},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{Subject: "peter"})).Return(Policies{}, errors.New("asdf"))
			},
			expectErr: true,
		},
		{
			description: "should pass",
			r: &Request{
				Subject:  "peter",
				Resource: "articles:1234",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{
					Subject:  "peter",
					Resource: "articles:1234",
					Action:   "view",
				})).Return(Policies{
					&DefaultPolicy{
						Subjects:  []string{"<zac|peter>"},
						Effect:    AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: false,
		},
		{
			description: "should fail because subjects don't match (unlikely event)",
			r: &Request{
				Subject:  "ken",
				Resource: "articles:1234",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{
					Subject:  "ken",
					Resource: "articles:1234",
					Action:   "view",
				})).Return(Policies{
					&DefaultPolicy{
						Subjects:  []string{"<zac|peter>"},
						Effect:    AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because resources mismatch",
			r: &Request{
				Subject:  "ken",
				Resource: "printers:321",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{
					Subject:  "ken",
					Resource: "printers:321",
					Action:   "view",
				})).Return(Policies{
					&DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because action mismatch",
			r: &Request{
				Subject:  "ken",
				Resource: "articles:321",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{
					Subject:  "ken",
					Resource: "articles:321",
					Action:   "view",
				})).Return(Policies{
					&DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"<foo|bar>"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should pass",
			r: &Request{
				Subject:  "ken",
				Resource: "articles:321",
				Action:   "foo",
			},
			setup: func() {
				m.EXPECT().FindRequestCandidates(ctx, gomock.Eq(&Request{
					Subject:  "ken",
					Resource: "articles:321",
					Action:   "foo",
				})).Return(Policies{
					&DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"<foo|bar>"},
					},
				}, nil)
			},
			expectErr: false,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, c.description), func(t *testing.T) {
			c.setup()
			err := w.IsAllowed(ctx, c.r)
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
