package ladon

eval_condition("StringPairsEqualCondition", request, options, key) {
    cast_array(request.context[key], pairs)
    cast_array(pairs[_], pair)

    count(pair, c)
    c == 2
    pair[0] == pair[1]
}

		{pairs: "junk", pass: false},
		{pairs: []interface{}{[]interface{}{}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "1", "2"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "2"}}, pass: false},
		{pairs: []interface{}{[]interface{}{"1", "1"}, []interface{}{"2", "3"}}, pass: false},
		{pairs: []interface{}{}, pass: true},
		{pairs: []interface{}{[]interface{}{"1", "1"}}, pass: true},
		{pairs: []interface{}{[]interface{}{"1", "1"}, []interface{}{"2", "2"}}, pass: true},

#test_condition_string_pairs_eqal {
#    eval_condition("StringPairsEqualCondition", { "subject": "some-subject", "context": { "foobar": "some-subject" } }, {}, "foobar")
#}
