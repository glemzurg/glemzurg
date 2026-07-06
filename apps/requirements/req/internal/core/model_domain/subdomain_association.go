package model_domain

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// SubdomainAssociation records when a subdomain enforces requirements on another subdomain in the same domain.
type SubdomainAssociation struct {
	Key                  identity.Key
	ProblemSubdomainKey  identity.Key
	SolutionSubdomainKey identity.Key
	UmlComment           string
}

func NewSubdomainAssociation(key, problemSubdomainKey, solutionSubdomainKey identity.Key, umlComment string) SubdomainAssociation {
	return SubdomainAssociation{
		Key:                  key,
		ProblemSubdomainKey:  problemSubdomainKey,
		SolutionSubdomainKey: solutionSubdomainKey,
		UmlComment:           umlComment,
	}
}

// Validate validates the SubdomainAssociation struct.
func (a *SubdomainAssociation) Validate(ctx *coreerr.ValidationContext) error {
	if err := a.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.SassocKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_SUBDOMAIN_ASSOCIATION {
		return coreerr.NewWithValues(ctx, coreerr.SassocKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for subdomain association", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_SUBDOMAIN_ASSOCIATION)
	}
	if err := a.ProblemSubdomainKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.SassocProblemkeyInvalid, fmt.Sprintf("ProblemSubdomainKey: %s", err.Error()), "ProblemSubdomainKey")
	}
	if a.ProblemSubdomainKey.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return coreerr.NewWithValues(ctx, coreerr.SassocProblemkeyType, fmt.Sprintf("ProblemSubdomainKey: invalid key type '%s' for subdomain", a.ProblemSubdomainKey.KeyType), "ProblemSubdomainKey", a.ProblemSubdomainKey.KeyType, identity.KEY_TYPE_SUBDOMAIN)
	}
	if err := a.SolutionSubdomainKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.SassocSolutionkeyInvalid, fmt.Sprintf("SolutionSubdomainKey: %s", err.Error()), "SolutionSubdomainKey")
	}
	if a.SolutionSubdomainKey.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return coreerr.NewWithValues(ctx, coreerr.SassocSolutionkeyType, fmt.Sprintf("SolutionSubdomainKey: invalid key type '%s' for subdomain", a.SolutionSubdomainKey.KeyType), "SolutionSubdomainKey", a.SolutionSubdomainKey.KeyType, identity.KEY_TYPE_SUBDOMAIN)
	}
	if a.ProblemSubdomainKey == a.SolutionSubdomainKey {
		return coreerr.New(ctx, coreerr.SassocSameSubdomains, "ProblemSubdomainKey and SolutionSubdomainKey cannot be the same", "ProblemSubdomainKey")
	}
	return nil
}

// ValidateWithParent validates the subdomain association and its key's parent relationship.
// The parent must be the owning domain.
func (a *SubdomainAssociation) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	if err := a.Validate(ctx); err != nil {
		return err
	}
	if err := a.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	return nil
}

// ValidateReferences validates that both subdomain keys exist in the domain and share its parent.
func (a *SubdomainAssociation) ValidateReferences(ctx *coreerr.ValidationContext, domainKey identity.Key, subdomains map[identity.Key]bool) error {
	if !subdomains[a.ProblemSubdomainKey] {
		return coreerr.NewWithValues(ctx, coreerr.SassocProblemNotfound, fmt.Sprintf("subdomain association '%s' references non-existent problem subdomain '%s'", a.Key.String(), a.ProblemSubdomainKey.String()), "ProblemSubdomainKey", a.ProblemSubdomainKey.String(), "")
	}
	if !subdomains[a.SolutionSubdomainKey] {
		return coreerr.NewWithValues(ctx, coreerr.SassocSolutionNotfound, fmt.Sprintf("subdomain association '%s' references non-existent solution subdomain '%s'", a.Key.String(), a.SolutionSubdomainKey.String()), "SolutionSubdomainKey", a.SolutionSubdomainKey.String(), "")
	}
	if !a.ProblemSubdomainKey.IsParent(domainKey) {
		return coreerr.NewWithValues(ctx, coreerr.SassocCrossDomain, fmt.Sprintf("subdomain association '%s' problem subdomain '%s' is not in domain '%s'", a.Key.String(), a.ProblemSubdomainKey.String(), domainKey.String()), "ProblemSubdomainKey", a.ProblemSubdomainKey.String(), domainKey.String())
	}
	if !a.SolutionSubdomainKey.IsParent(domainKey) {
		return coreerr.NewWithValues(ctx, coreerr.SassocCrossDomain, fmt.Sprintf("subdomain association '%s' solution subdomain '%s' is not in domain '%s'", a.Key.String(), a.SolutionSubdomainKey.String(), domainKey.String()), "SolutionSubdomainKey", a.SolutionSubdomainKey.String(), domainKey.String())
	}
	return nil
}
