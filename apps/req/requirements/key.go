package requirements

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

func PreenKey(key string) (preened string, err error) {

	preened = key
	preened = strings.ToLower(preened)
	preened = strings.TrimSpace(preened)

	err = validation.Validate(preened,
		validation.Required, // not empty
	)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return preened, nil
}
