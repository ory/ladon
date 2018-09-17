package ladon

eval_condition("StringEqualCondition", request, options, key) {
    cast_string(request.context[key], a)
    cast_string(options.equals, b)
    a == b
}

test_condition_string_equal {
    eval_condition("StringEqualCondition", { "context": {"foobar": "the-value-should-be-this" } }, { "equals": "the-value-should-be-this" }, "foobar")

    not eval_condition("StringEqualCondition", { "context": {"not-foobar": "the-value-should-be-this" } }, { "equals": "the-value-should-be-this" }, "foobar")
    not eval_condition("StringEqualCondition", { "context": {"foobar": "the-value-should-be-this" } }, { "not-equals": "the-value-should-be-this" }, "foobar")
    not eval_condition("StringEqualCondition", { "context": {"not-foobar": "the-value-should-be-this" } }, { "not-equals": "the-value-should-be-this" }, "foobar")
    not eval_condition("StringEqualCondition", { "context": {"foobar": "the-value-should-be-this" } }, { "equals": "not-the-value-should-be-this" }, "foobar")
}
