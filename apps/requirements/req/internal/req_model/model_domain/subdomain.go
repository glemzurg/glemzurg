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
	// For parsing only, not stored here.
	Generalizations []model_class.Generalization // Generalizations for the classes and use cases in this subdomain.
	Classes         []model_class.Class          // Classes in this subdomain.
	UseCases        []model_use_case.UseCase     // Use cases in this subdomain.
	Associations    []model_class.Association    // Associations between classes in this subdomain.
}

func NewSubdomain(key identity.Key, name, details, umlComment string) (subdomain Subdomain, err error) {

	subdomain = Subdomain{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&subdomain,
		validation.Field(&subdomain.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_SUBDOMAIN {
				return errors.Errorf("invalid key type '%s' for subdomain", k.KeyType())
			}
			return nil
		})),
		validation.Field(&subdomain.Name, validation.Required),
	)
	if err != nil {
		return Subdomain{}, errors.WithStack(err)
	}

	return subdomain, nil
}

// ValidateWithParent validates the Subdomain and verifies its key has the correct parent.
// The parent must be a Domain.
func (s *Subdomain) ValidateWithParent(parent *identity.Key) error {
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range s.Generalizations {
		if err := s.Generalizations[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for i := range s.Classes {
		if err := s.Classes[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for i := range s.UseCases {
		if err := s.UseCases[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	for i := range s.Associations {
		if err := s.Associations[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	return nil
}
