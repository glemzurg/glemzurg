package requirements

import (
	"regexp"
)

func init() {

	// Regex syntax used by golang: https://github.com/google/re2/wiki/Syntax

	// A requirement header.
	_reqHeaderRegexp = regexp.MustCompile(`^(#+)[[:space:]]*(F|R|A|S|U|N|f|r|a|s|u|n)([[:digit:]]*)[[:space:]]*[.[:punct:]][[:space:]]+([[:alnum:]].*)`)

	// One or more spaces.
	_whitespaceRegexp = regexp.MustCompile(`[[:space:]]+`)

	// One or more punctional (including period).
	_puncRegexp = regexp.MustCompile(`[.[:punct:]]+`)

	// A markdown link to an in file ref in the form of: [F13][] or [F13][F13]
	_linkRegexp = regexp.MustCompile(`\[((F|R|A|S|U|N|f|r|a|s|u|n)([1-9][[:digit:]]*))\](\[((F|R|A|S|U|N|f|r|a|s|u|n)([1-9][[:digit:]]*))?\])`)

	// A markdown ref link to a requirement in the form of: \n[F13] or \n[F13]: ../path/to/file#title "Title"
	// The quotes can be either "Title", 'Title', or (Title)
	_refLinkTextlineRegexp = regexp.MustCompile(`^\[((F|R|A|S|U|N|f|r|a|s|u|n)([1-9][[:digit:]]*))\]:.*$`)

	// A aspect table header line like: |-----------|------------|
	_aspectTableHeaderLineRegexp = regexp.MustCompile(`^[[:space:]]*\|[[:space:]]*\-+[[:space:]]*\|[[:space:]]*\-+[[:space:]]*\|[[:space:]]*$`)

	// A aspect table value like: | Priority | high |
	_aspectTableValueRegexp = regexp.MustCompile(`^[[:space:]]*\|[[:space:]]*([^\|]+)[[:space:]]*\|[[:space:]]*([^\|]+)[[:space:]]*\|[[:space:]]*$`)

	// A bullet line in markdown.
	_markdownBulletLineRegexp = regexp.MustCompile(`^[[:space:]]*(-|\*|\+)+.+$`)
	_markdownNumericBulletLineRegexp = regexp.MustCompile(`^[[:space:]]*[[:digit:]]+\..+$`)
}

// Regex match a header line.
var _reqHeaderRegexp *regexp.Regexp

// Regex match one or more spaces.
var _whitespaceRegexp *regexp.Regexp

// Regex match one or more punc.
var _puncRegexp *regexp.Regexp

// Regex match a single link.
var _linkRegexp *regexp.Regexp

// Regex match a reference link at the bottom of the file.
// This regex is for single textlines.
var _refLinkTextlineRegexp *regexp.Regexp

// Regex match a header line in an aspect table.
var _aspectTableHeaderLineRegexp *regexp.Regexp

// Regex match a value in an aspect table.
var _aspectTableValueRegexp *regexp.Regexp

// Regex matches a bullet in markdown.
var _markdownBulletLineRegexp *regexp.Regexp
var _markdownNumericBulletLineRegexp *regexp.Regexp

func normalizeWhitespace(value string) (normalized string) {
	return _whitespaceRegexp.ReplaceAllString(value, " ")
}

func puncToWhitespace(value string) (normalized string) {
	return _puncRegexp.ReplaceAllString(value, " ")
}
