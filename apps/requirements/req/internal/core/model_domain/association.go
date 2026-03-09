package model_domain

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// When a domain enforces requirements on another domain.
type Association struct {
	Key               identity.Key
	ProblemDomainKey  identity.Key
	SolutionDomainKey identity.Key
	UmlComment        string
}

func NewAssociation(key, problemDomainKey, solutionDomainKey identity.Key, umlComment string) Association {
	return Association{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        umlComment,
	}
}

// Validate validates the domain Association struct.
func (a *Association) Validate() error {
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return coreerr.New(coreerr.DassocKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_DOMAIN_ASSOCIATION {
		return coreerr.NewWithValues(coreerr.DassocKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for domain association", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_DOMAIN_ASSOCIATION)
	}
	// Validate ProblemDomainKey.
	if err := a.ProblemDomainKey.Validate(); err != nil {
		return coreerr.New(coreerr.DassocProblemkeyInvalid, fmt.Sprintf("ProblemDomainKey: %s", err.Error()), "ProblemDomainKey")
	}
	if a.ProblemDomainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return coreerr.NewWithValues(coreerr.DassocProblemkeyType, fmt.Sprintf("ProblemDomainKey: invalid key type '%s' for domain", a.ProblemDomainKey.KeyType), "ProblemDomainKey", a.ProblemDomainKey.KeyType, identity.KEY_TYPE_DOMAIN)
	}
	// Validate SolutionDomainKey.
	if err := a.SolutionDomainKey.Validate(); err != nil {
		return coreerr.New(coreerr.DassocSolutionkeyInvalid, fmt.Sprintf("SolutionDomainKey: %s", err.Error()), "SolutionDomainKey")
	}
	if a.SolutionDomainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return coreerr.NewWithValues(coreerr.DassocSolutionkeyType, fmt.Sprintf("SolutionDomainKey: invalid key type '%s' for domain", a.SolutionDomainKey.KeyType), "SolutionDomainKey", a.SolutionDomainKey.KeyType, identity.KEY_TYPE_DOMAIN)
	}
	// ProblemDomainKey and SolutionDomainKey cannot be the same.
	if a.ProblemDomainKey == a.SolutionDomainKey {
		return coreerr.New(coreerr.DassocSameDomains, "ProblemDomainKey and SolutionDomainKey cannot be the same", "ProblemDomainKey")
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
// - SolutionDomainKey must exist in the domains map.
func (a *Association) ValidateReferences(domains map[identity.Key]bool) error {
	if !domains[a.ProblemDomainKey] {
		return coreerr.NewWithValues(coreerr.DassocProblemNotfound, fmt.Sprintf("domain association '%s' references non-existent problem domain '%s'", a.Key.String(), a.ProblemDomainKey.String()), "ProblemDomainKey", a.ProblemDomainKey.String(), "")
	}
	if !domains[a.SolutionDomainKey] {
		return coreerr.NewWithValues(coreerr.DassocSolutionNotfound, fmt.Sprintf("domain association '%s' references non-existent solution domain '%s'", a.Key.String(), a.SolutionDomainKey.String()), "SolutionDomainKey", a.SolutionDomainKey.String(), "")
	}
	return nil
}
