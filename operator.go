package ladon

type Operator func(extra map[string]interface{}, ctx *Context) bool
