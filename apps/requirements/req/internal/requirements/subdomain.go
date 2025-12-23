package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Subdomain is a nested category of the model.
type Subdomain struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// For parsing only, not stored here.
	Generalizations []Generalization // Generalizations for the classes and use cases in this subdomain.
	Classes         []Class          // Classes in this subdomain.
	UseCases        []UseCase        // Use cases in this subdomain.
	Associations    []Association    // Associations between classes in this subdomain.
}

func NewSubdomain(key, name, details, umlComment string) (subdomain Subdomain, err error) {

	subdomain = Subdomain{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&subdomain,
		validation.Field(&subdomain.Key, validation.Required),
		validation.Field(&subdomain.Name, validation.Required),
	)
	if err != nil {
		return Subdomain{}, errors.WithStack(err)
	}

	return subdomain, nil
}
