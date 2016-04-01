package ladon_test

import (
	"testing"

	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	"github.com/ory-am/ladon"
	"github.com/ory-am/ladon/internal"
	"github.com/stretchr/testify/assert"
)

func TestWardenIsGranted(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := internal.NewMockManager(ctrl)
	defer ctrl.Finish()

	w := &ladon.Ladon{
		Manager: m,
	}

	for k, c := range []struct {
		r           *ladon.Request
		description string
		setup       func()
		expectErr   bool
	}{
		{
			description: "should fail because no policies are found for peter",
			r:           &ladon.Request{Subject: "peter"},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("peter").Return(ladon.Policies{}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because lookup failure when accessing policies for peter",
			r:           &ladon.Request{Subject: "peter"},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("peter").Return(ladon.Policies{}, errors.New("asdf"))
			},
			expectErr: true,
		},
		{
			description: "should pass",
			r: &ladon.Request{
				Subject:  "peter",
				Resource: "articles:1234",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("peter").Return(ladon.Policies{
					&ladon.DefaultPolicy{
						Subjects:  []string{"<zac|peter>"},
						Effect:    ladon.AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: false,
		},
		{
			description: "should fail because subjects don't match (unlikely event)",
			r: &ladon.Request{
				Subject:  "ken",
				Resource: "articles:1234",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("ken").Return(ladon.Policies{
					&ladon.DefaultPolicy{
						Subjects:  []string{"<zac|peter>"},
						Effect:    ladon.AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because resources mismatch",
			r: &ladon.Request{
				Subject:  "ken",
				Resource: "printers:321",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("ken").Return(ladon.Policies{
					&ladon.DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    ladon.AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"view"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should fail because action mismatch",
			r: &ladon.Request{
				Subject:  "ken",
				Resource: "articles:321",
				Action:   "view",
			},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("ken").Return(ladon.Policies{
					&ladon.DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    ladon.AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"<foo|bar>"},
					},
				}, nil)
			},
			expectErr: true,
		},
		{
			description: "should pass",
			r: &ladon.Request{
				Subject:  "ken",
				Resource: "articles:321",
				Action:   "foo",
			},
			setup: func() {
				m.EXPECT().FindPoliciesForSubject("ken").Return(ladon.Policies{
					&ladon.DefaultPolicy{
						Subjects:  []string{"ken", "peter"},
						Effect:    ladon.AllowAccess,
						Resources: []string{"articles:<[0-9]+>"},
						Actions:   []string{"<foo|bar>"},
					},
				}, nil)
			},
			expectErr: false,
		},
	} {
		c.setup()
		err := w.IsAllowed(c.r)
		if c.expectErr {
			assert.NotNil(t, err, "(%d) %s", k, c.description)
		} else {
			assert.Nil(t, err, "(%d) %s", k, c.description)
		}
		t.Logf("Passed test case (%d) %s", k, c.description)
	}
}
