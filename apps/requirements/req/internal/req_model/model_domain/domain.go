package model_domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// Domain is a root category of the mode.
type Domain struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Children
	DomainAssociations []Association
	Classes            []model_class.Class
	UseCases           []model_use_case.UseCase
	Subdomains         []Subdomain
	ClassAssociations  []model_class.Association // Associations between classes that bridge subdomains in this domain.
}

func NewDomain(key identity.Key, name, details string, realized bool, umlComment string) (domain Domain, err error) {

	domain = Domain{
		Key:        key,
		Name:       name,
		Details:    details,
		Realized:   realized,
		UmlComment: umlComment,
	}

	if err = domain.Validate(); err != nil {
		return Domain{}, err
	}

	return domain, nil
}

// Validate validates the Domain struct.
func (d *Domain) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_DOMAIN {
				return errors.Errorf("invalid key type '%s' for domain", k.KeyType())
			}
			return nil
		})),
		validation.Field(&d.Name, validation.Required),
	)
}

// ValidateWithParent validates the Domain, its key's parent relationship, and all children.
// The parent must be nil (domains are root-level entities).
func (d *Domain) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := d.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := d.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range d.Subdomains {
		if err := d.Subdomains[i].ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	// Domain-level classes (if any) - these would have the domain as implicit parent
	// but since Class expects a subdomain parent, we validate domain-level associations
	// which span subdomains within this domain.
	for i := range d.Associations {
		if err := d.Associations[i].ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	return nil
}
