package model_domain

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// When a domain enforces requirements on another domain.
type Association struct {
	Key               identity.Key
	ProblemDomainKey  identity.Key
	SolutionDomainKey identity.Key
	UmlComment        string
}

func NewAssociation(key, problemDomainKey, solutionDomainKey identity.Key, umlComment string) (association Association, err error) {

	association = Association{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        umlComment,
	}

	err = validation.ValidateStruct(&association,
		validation.Field(&association.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_DOMAIN_ASSOCIATION {
				return errors.Errorf("invalid key type '%s' for domain association", k.KeyType())
			}
			return nil
		})),
		validation.Field(&association.ProblemDomainKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			return k.Validate()
		})),
		validation.Field(&association.SolutionDomainKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			return k.Validate()
		})),
	)
	if err != nil {
		return Association{}, errors.WithStack(err)
	}

	return association, nil
}
