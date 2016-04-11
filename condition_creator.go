package ladon

import (
	"fmt"
	"github.com/go-errors/errors"
)

type ConditionCreators []ConditionCreator

type ConditionCreator func(map[string]interface{}) (Condition, error)

var DefaultConditionCreators = map[string]ConditionCreator{
	"SubjectIsOwnerCondition": func(_ map[string]interface{}) (Condition, error) {
		return new(SubjectIsOwnerCondition), nil
	},
	"SubjectIsNotOwnerCondition": func(_ map[string]interface{}) (Condition, error) {
		return new(SubjectIsNotOwnerCondition), nil
	},
	"CIDRCondition": func(data map[string]interface{}) (Condition, error) {
		var cidr string
		var err error
		if cidr, err = toString("cidr", data); err != nil {
			return nil, errors.New(err)
		}

		return &CIDRCondition{
			CIDR: cidr,
		}, nil
	},
}

func CreateCondition(allowedConditionCreators map[string]ConditionCreator, data map[string]interface{}) (c Condition, err error) {
	var name string
	if name, err = toString("condition", data); err != nil {
		return nil, errors.New(err)
	}

	condition, ok := allowedConditionCreators[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown condition %s", name))
	}

	return condition(data)
}

func toString(field string, m map[string]interface{}) (string, error) {
	var s string
	if t, k := m["cidr"]; !k {
		return "", errors.New(`Field "`+field+`" is missing`)
	} else if s, k = t.(string); !k {
		return "", errors.New(`Field "`+field+`" is not of type string`)
	}

	return s, nil
}
