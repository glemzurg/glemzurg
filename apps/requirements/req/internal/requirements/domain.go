package requirements

import (
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
	Classes      []Class
	UseCases     []UseCase
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

func createKeyDomainLookup(domainClasses map[string][]Class, domainUseCases map[string][]UseCase, items []Domain) (lookup map[string]Domain) {

	lookup = map[string]Domain{}
	for _, item := range items {

		item.Classes = domainClasses[item.Key]
		item.UseCases = domainUseCases[item.Key]

		lookup[item.Key] = item
	}
	return lookup
}
