package ladon

type matcher interface {
	Matches(p Policy, haystack []string, needle string) (matches bool, error error)
}

var DefaultMatcher = NewRegexpMatcher(512)
