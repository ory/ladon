package ladon_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/ory/ladon"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWardenIsGranted(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockManager(ctrl)
	defer ctrl.Finish()

	w := &Ladon{
		Manager: m,
	}

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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{Subject: "peter"})).Return(Policies{}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because lookup failure when accessing policies for peter",
			r:           &Request{Subject: "peter"},
			setup: func() {
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{Subject: "peter"})).Return(Policies{}, errors.New("asdf"))
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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{
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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{
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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{
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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{
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
				m.EXPECT().FindRequestCandidates(gomock.Eq(&Request{
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
			err := w.IsAllowed(c.r)
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
