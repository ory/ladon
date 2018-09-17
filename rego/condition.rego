package ladon

all_conditions_true(policy) {
    not any_condition_false(policy)
}

any_condition_false(policy) {
    c := policy.conditions[condition_key]
    not eval_condition(c.type, input, c.options, condition_key)
}
