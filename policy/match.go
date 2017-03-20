package policy

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/ory/common/compiler"
)

// Match matches a needle with an array of regular expressions and returns true
// if a match was found.
func Match(p Policy, haystack []string, needle string) (bool, error) {
	var reg *regexp.Regexp
	var err error
	for _, h := range haystack {
		reg, err = compiler.CompileRegex(h, p.GetStartDelimiter(), p.GetEndDelimiter())
		if err != nil {
			return false, errors.WithStack(err)
		}

		if reg.MatchString(needle) {
			return true, nil
		}
	}
	return false, nil
}
