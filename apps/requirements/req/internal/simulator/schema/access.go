package schema

import (
	"maps"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ModelInvariants returns run-scoped model-level invariants.
func (s *Schema) ModelInvariants() []model_logic.Logic {
	if s == nil {
		return nil
	}
	return s.modelInvariants
}

// NamedSets returns model-level named sets from the owned model.
func (s *Schema) NamedSets() map[identity.Key]model_logic.NamedSet {
	if s == nil || s.model == nil || len(s.model.NamedSets) == 0 {
		return nil
	}
	out := make(map[identity.Key]model_logic.NamedSet, len(s.model.NamedSets))
	maps.Copy(out, s.model.NamedSets)
	return out
}

// AllUseCases returns every use case on the owned model.
func (s *Schema) AllUseCases() []model_use_case.UseCase {
	if s == nil || s.model == nil {
		return nil
	}
	var out []model_use_case.UseCase
	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, uc := range subdomain.UseCases {
				out = append(out, uc)
			}
		}
	}
	return out
}
