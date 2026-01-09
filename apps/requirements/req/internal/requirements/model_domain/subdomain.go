package model_domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
)

// Construct a key that sits correctly in the model shape.
func NewSubdomainKey(domainKey identity.Key, subKey string) (key identity.Key, err error) {
	return identity.NewKey(domainKey.String(), identity.SUBDOMAIN_KEY_TYPE, subKey)
}

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
			if k.ChildType() != identity.SUBDOMAIN_KEY_TYPE {
				return errors.New("invalid child type for subdomain")
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
