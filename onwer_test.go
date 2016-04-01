package ladon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var extra = map[string]interface{}{"subject": "foo"}

func TestSubjectIsOwner(t *testing.T) {
	assert.True(t, SubjectIsOwner(extra, &Context{Owner: "foo"}))
	assert.False(t, SubjectIsOwner(extra, &Context{Owner: "not-foo"}))
	assert.False(t, SubjectIsOwner(map[string]interface{}{}, &Context{Owner: "not-foo"}))
	assert.False(t, SubjectIsOwner(map[string]interface{}{}, &Context{}))
}

func TestSubjectIsOwnerEdgeCases(t *testing.T) {
	assert.False(t, SubjectIsOwner(nil, &Context{}))
	assert.False(t, SubjectIsOwner(nil, &Context{Owner: "not-foo"}))
	assert.False(t, SubjectIsOwner(map[string]interface{}{}, nil))
	assert.False(t, SubjectIsOwner(extra, nil))
}

func TestSubjectIsNotOwner(t *testing.T) {
	assert.False(t, SubjectIsNotOwner(extra, &Context{Owner: "foo"}))
	assert.True(t, SubjectIsNotOwner(extra, &Context{Owner: "not-foo"}))
	assert.False(t, SubjectIsNotOwner(map[string]interface{}{}, &Context{Owner: "not-foo"}))
	assert.False(t, SubjectIsNotOwner(map[string]interface{}{}, &Context{}))
}

func TestSubjectIsNotOwnerEdgeCases(t *testing.T) {
	assert.False(t, SubjectIsNotOwner(nil, &Context{Owner: "foo"}))
	assert.False(t, SubjectIsNotOwner(nil, &Context{}))
	assert.False(t, SubjectIsNotOwner(map[string]interface{}{}, nil))
	assert.False(t, SubjectIsNotOwner(extra, nil))
}
