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
	DomainAssociations map[identity.Key]Association
	Classes            map[identity.Key]model_class.Class
	UseCases           map[identity.Key]model_use_case.UseCase
	Subdomains         map[identity.Key]Subdomain
	ClassAssociations  map[identity.Key]model_class.Association // Associations between classes that bridge subdomains in this domain.
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
	for _, assoc := range d.DomainAssociations {
		if err := assoc.ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	for _, class := range d.Classes {
		if err := class.ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	for _, useCase := range d.UseCases {
		if err := useCase.ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	for _, subdomain := range d.Subdomains {
		if err := subdomain.ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	for _, classAssoc := range d.ClassAssociations {
		if err := classAssoc.ValidateWithParent(&d.Key); err != nil {
			return err
		}
	}
	return nil
}
