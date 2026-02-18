package identity

import (
	"strings"

	"github.com/pkg/errors"
)

func PreenKey(key string) (preened string, err error) {

	preened = key
	preened = strings.ToLower(preened)
	preened = strings.TrimSpace(preened)

	if preened == "" {
		return "", errors.New("cannot be blank")
	}

	return preened, nil
}
