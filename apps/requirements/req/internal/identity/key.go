package identity

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

// HasPrefix returns a validation rule that checks if the string value has a prefix constructed from parent and childType.
// The prefix is parent/childType/, and both parent and childType must be non-blank.
// This is used for validating derivative keys that should start with their parent key followed by the child type.
func HasPrefix(parent, childType string) validation.Rule {
	if parent == "" {
		return validation.By(func(interface{}) error {
			return errors.New("parent cannot be blank")
		})
	}
	if childType == "" {
		return validation.By(func(interface{}) error {
			return errors.New("childType cannot be blank")
		})
	}
	prefix := parent + "/" + childType + "/"
	return validation.By(func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("must be a string")
		}
		if !strings.HasPrefix(s, prefix) {
			return errors.Errorf("must have prefix %s", prefix)
		}
		if strings.Contains(s[len(prefix):], "/") {
			return errors.Errorf("must not contain '/' after prefix %s", prefix)
		}
		return nil
	})
}
