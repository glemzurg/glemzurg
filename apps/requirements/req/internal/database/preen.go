package database

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Regex match one or more spaces.
var _whitespaceRegexp *regexp.Regexp = regexp.MustCompile(`[[:space:]]+`)

func normalizeWhitespace(value string) (normalized string) {
	return _whitespaceRegexp.ReplaceAllString(value, "-")
}

func preenKey(key string) (preened string, err error) {

	preened = key
	preened = strings.ToLower(preened)
	preened = strings.TrimSpace(preened)

	if preened == "" {
		return "", errors.New("cannot be blank")
	}

	preened = normalizeWhitespace(preened)

	return preened, nil
}
