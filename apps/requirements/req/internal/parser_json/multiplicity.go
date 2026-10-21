package parser_json

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
)

// multiplicityFromCore emits a token NewMultiplicity accepts.
// Multiplicity.String uses diagram spellings ("*", "1..*") that do not parse back.
func multiplicityFromCore(multiplicity model_class.Multiplicity) (string, error) {
	token := multiplicity.ParsedString()
	if strings.HasSuffix(token, "..*") {
		token = strings.TrimSuffix(token, "*") + "many"
	}
	parsed, err := model_class.NewMultiplicity(token)
	if err != nil {
		return "", errors.Wrapf(err, "multiplicity %q", multiplicity.String())
	}
	if parsed != multiplicity {
		return "", errors.Errorf("multiplicity %q encoded as %q does not round-trip", multiplicity.String(), token)
	}
	return token, nil
}

func multiplicityToCore(field, token string) (model_class.Multiplicity, error) {
	parsed, err := model_class.NewMultiplicity(token)
	if err != nil {
		return model_class.Multiplicity{}, errors.Wrapf(err, "%s", field)
	}
	return parsed, nil
}
