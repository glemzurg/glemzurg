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

	if err = association.Validate(); err != nil {
		return Association{}, err
	}

	return association, nil
}

// Validate validates the domain Association struct.
func (a *Association) Validate() error {
	if err := validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_DOMAIN_ASSOCIATION {
				return errors.Errorf("invalid key type '%s' for domain association", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.ProblemDomainKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_DOMAIN {
				return errors.Errorf("invalid key type '%s' for domain", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.SolutionDomainKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_DOMAIN {
				return errors.Errorf("invalid key type '%s' for domain", k.KeyType())
			}
			return nil
		})),
	); err != nil {
		return err
	}
	// ProblemDomainKey and SolutionDomainKey cannot be the same.
	if a.ProblemDomainKey == a.SolutionDomainKey {
		return errors.New("ProblemDomainKey and SolutionDomainKey cannot be the same")
	}
	return nil
}

// ValidateWithParent validates the domain Association, its key's parent relationship, and all children.
// The parent must be nil (domain associations are root-level entities).
func (a *Association) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Association has no children with keys that need validation.
	return nil
}
