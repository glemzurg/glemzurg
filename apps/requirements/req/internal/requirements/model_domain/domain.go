package model_domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
)

// Domain is a root category of the mode.
type Domain struct {
	Key        identity.Key // Unique in model. No prefix pattern.
	Name       string
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Part of the data in a parsed file.
	Associations []Association
	Classes      []model_class.Class
	UseCases     []model_use_case.UseCase
	Subdomains   []Subdomain
}

func NewDomain(key identity.Key, name, details string, realized bool, umlComment string) (domain Domain, err error) {

	domain = Domain{
		Key:        key,
		Name:       name,
		Details:    details,
		Realized:   realized,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&domain,
		validation.Field(&domain.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			return k.Validate()
		})),
		validation.Field(&domain.Name, validation.Required),
	)
	if err != nil {
		return Domain{}, errors.WithStack(err)
	}

	return domain, nil
}
