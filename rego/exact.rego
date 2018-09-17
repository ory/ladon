package ladon

import data.policies as policies

default allow_exact = false

allow_exact {
	effects := [effect | effect := policies[i].effect
			policies[i].resources[_] == input.resource
			policies[i].subjects[_] == input.subject
			policies[i].actions[_] == input.action
			all_conditions_true(policies[i])
		]

	effect_allow(effects)
}
