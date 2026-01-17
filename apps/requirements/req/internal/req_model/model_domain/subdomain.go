package model_domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// Subdomain is a nested category of the model.
type Subdomain struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Children
	Generalizations   map[identity.Key]model_class.Generalization // Generalizations for the classes and use cases in this subdomain.
	Classes           map[identity.Key]model_class.Class          // Classes in this subdomain.
	UseCases          map[identity.Key]model_use_case.UseCase     // Use cases in this subdomain.
	ClassAssociations map[identity.Key]model_class.Association    // Associations between classes in this subdomain.
}

func NewSubdomain(key identity.Key, name, details, umlComment string) (subdomain Subdomain, err error) {

	subdomain = Subdomain{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	if err = subdomain.Validate(); err != nil {
		return Subdomain{}, err
	}

	return subdomain, nil
}

// Validate validates the Subdomain struct.
func (s *Subdomain) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_SUBDOMAIN {
				return errors.Errorf("invalid key type '%s' for subdomain", k.KeyType())
			}
			return nil
		})),
		validation.Field(&s.Name, validation.Required),
	)
}

// ValidateWithParent validates the Subdomain, its key's parent relationship, and all children.
// The parent must be a Domain.
func (s *Subdomain) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := s.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for _, gen := range s.Generalizations {
		if err := gen.ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for _, class := range s.Classes {
		if err := class.ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for _, useCase := range s.UseCases {
		if err := useCase.ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for _, classAssoc := range s.ClassAssociations {
		if err := classAssoc.ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	return nil
}
