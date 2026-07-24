package schema

import (
	"maps"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ForEachClass calls fn for every in-scope class (model tree type).
func (s *Schema) ForEachClass(fn func(model_class.Class)) {
	if s == nil || fn == nil {
		return
	}
	for _, class := range s.classes {
		fn(class)
	}
}

// ForEachAssociation calls fn for every association on the owned model.
func (s *Schema) ForEachAssociation(fn func(model_class.Association)) {
	if s == nil || fn == nil {
		return
	}
	for _, assoc := range s.associations {
		fn(assoc)
	}
}

// ModelInvariants returns model-level invariants from the owned model.
func (s *Schema) ModelInvariants() []model_logic.Logic {
	if s == nil || s.model == nil {
		return nil
	}
	return s.model.Invariants
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

// globalFunctions returns model-level global functions from the owned model.
func (s *Schema) globalFunctions() map[identity.Key]model_logic.GlobalFunction {
	if s == nil || s.model == nil || len(s.model.GlobalFunctions) == 0 {
		return nil
	}
	out := make(map[identity.Key]model_logic.GlobalFunction, len(s.model.GlobalFunctions))
	maps.Copy(out, s.model.GlobalFunctions)
	return out
}

// ForEachUseCase calls fn for every use case on the owned model.
func (s *Schema) ForEachUseCase(fn func(model_use_case.UseCase)) {
	if s == nil || s.model == nil || fn == nil {
		return
	}
	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, uc := range subdomain.UseCases {
				fn(uc)
			}
		}
	}
}
