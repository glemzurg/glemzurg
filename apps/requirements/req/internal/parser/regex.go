package parser

import (
	"regexp"
	"strings"
)

func init() {

	// Regex syntax used by golang: https://github.com/google/re2/wiki/Syntax

	// One or more spaces.
	_whitespaceRegexp = regexp.MustCompile(`[[:space:]]+`)

}

// Regex match one or more spaces.
var _whitespaceRegexp *regexp.Regexp

func normalizeWhitespace(value string) (normalized string) {
	return _whitespaceRegexp.ReplaceAllString(strings.TrimSpace(value), " ")
}
