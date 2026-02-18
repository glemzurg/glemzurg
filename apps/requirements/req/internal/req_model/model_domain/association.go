package model_domain

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return err
	}
	if a.Key.KeyType != identity.KEY_TYPE_DOMAIN_ASSOCIATION {
		return errors.Errorf("Key: invalid key type '%s' for domain association", a.Key.KeyType)
	}
	// Validate ProblemDomainKey.
	if err := a.ProblemDomainKey.Validate(); err != nil {
		return errors.Wrap(err, "ProblemDomainKey")
	}
	if a.ProblemDomainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return errors.Errorf("ProblemDomainKey: invalid key type '%s' for domain", a.ProblemDomainKey.KeyType)
	}
	// Validate SolutionDomainKey.
	if err := a.SolutionDomainKey.Validate(); err != nil {
		return errors.Wrap(err, "SolutionDomainKey")
	}
	if a.SolutionDomainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return errors.Errorf("SolutionDomainKey: invalid key type '%s' for domain", a.SolutionDomainKey.KeyType)
	}
	// ProblemDomainKey and SolutionDomainKey cannot be the same.
	if a.ProblemDomainKey == a.SolutionDomainKey {
		return errors.New("ProblemDomainKey and SolutionDomainKey cannot be the same")
	}
	return nil
}

// ValidateWithParent validates the domain Association, its key's parent relationship, and all children.
// The parent must be nil since domain associations are root-level entities with no parent.
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

// ValidateReferences validates that the association's domain keys reference real domains.
// - ProblemDomainKey must exist in the domains map
// - SolutionDomainKey must exist in the domains map
func (a *Association) ValidateReferences(domains map[identity.Key]bool) error {
	if !domains[a.ProblemDomainKey] {
		return errors.Errorf("domain association '%s' references non-existent problem domain '%s'", a.Key.String(), a.ProblemDomainKey.String())
	}
	if !domains[a.SolutionDomainKey] {
		return errors.Errorf("domain association '%s' references non-existent solution domain '%s'", a.Key.String(), a.SolutionDomainKey.String())
	}
	return nil
}
