package domain

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/use_case"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Domain is a root category of the mode.
type Domain struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Part of the data in a parsed file.
	Associations []DomainAssociation
	Classes      []class.Class
	UseCases     []use_case.UseCase
	Subdomains   []Subdomain
}

func NewDomain(key, name, details string, realized bool, umlComment string) (domain Domain, err error) {

	domain = Domain{
		Key:        key,
		Name:       name,
		Details:    details,
		Realized:   realized,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&domain,
		validation.Field(&domain.Key, validation.Required),
		validation.Field(&domain.Name, validation.Required),
	)
	if err != nil {
		return Domain{}, errors.WithStack(err)
	}

	return domain, nil
}

func createKeyDomainLookup(domainClasses map[string][]class.Class, domainUseCases map[string][]use_case.UseCase, items []Domain) (lookup map[string]Domain) {

	lookup = map[string]Domain{}
	for _, item := range items {

		item.Classes = domainClasses[item.Key]
		item.UseCases = domainUseCases[item.Key]

		lookup[item.Key] = item
	}
	return lookup
}
