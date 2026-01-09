package identity

import (
	"fmt"
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

// HasPrefix returns a validation rule that checks if the string value has the specified prefix.
// This is used for validating derivative keys that should start with their parent key.
func HasPrefix(prefix string) validation.Rule {
	return validation.By(func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("must be a string")
		}
		if !strings.HasPrefix(s, prefix) {
			return fmt.Errorf("must have prefix %s", prefix)
		}
		return nil
	})
}
