package ladon

policies = [
    {
    	"id": "1",
        "resources": [`articles:1`],
        "subjects": [`subjects:1`],
        "actions": [`actions:1`],
        "effect": "allow",
        "conditions": {
        	"foobar": {
		        "type": "StringEqualCondition",
		        "options": {
		            "equals": "the-value-should-be-this"
		        }
        	}
        }
    },
    {
    	"id": "2",
        "resources": [`articles:2`],
        "subjects": [`subjects:2`],
        "actions": [`actions:2`],
        "effect": "deny",
    },
    {
    	"id": "3-1",
        "resources": [`articles:3`],
        "subjects": [`subjects:3`],
        "actions": [`actions:3`],
        "effect": "allow",
    },
    {
    	"id": "3-2",
        "resources": [`articles:3`],
        "subjects": [`subjects:3`],
        "actions": [`actions:3`],
        "effect": "deny",
    },
    {
    	"id": "3-3",
        "resources": [`articles:3`],
        "subjects": [`subjects:3`],
        "actions": [`actions:3`],
        "effect": "allow",
    },
]

test_exact_deny_policy {
    not allow_exact with input as {"resource": "articles:2", "subject": "subjects:2", "action": "actions:2"}
}

test_exact_deny_overrides {
    not allow_exact with input as {"resource": "articles:3", "subject": "subjects:3", "action": "actions:3"}
}
