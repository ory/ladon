package memory

import (
	"testing"

	"github.com/ory-am/ladon"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	m := New()
	_, err := m.Get("123")
	assert.NotNil(t, err)

	p := &ladon.DefaultPolicy{
		ID:       "123",
		Subjects: []string{"<ken|zac>", "peter"},
	}
	assert.Nil(t, m.Create(p))

	r, err := m.Get("123")
	assert.Nil(t, err)
	assert.Equal(t, p, r)

	ps, err := m.FindPoliciesForSubject("peter")
	assert.Nil(t, err)
	assert.Equal(t, ps[0], r)

	ps, err = m.FindPoliciesForSubject("zac")
	assert.Nil(t, err)
	assert.Equal(t, ps[0], r)

	ps, err = m.FindPoliciesForSubject("max")
	assert.Nil(t, err)
	assert.Len(t, ps, 0)

	err = m.Delete("123")
	assert.Nil(t, err)

	_, err = m.Get("123")
	assert.NotNil(t, err)
}
