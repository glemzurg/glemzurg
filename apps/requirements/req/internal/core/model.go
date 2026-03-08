package core

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key        string // Models do not have keys like other entitites. They just need to be unique to other models in the system.
	Name       string
	Details    string              // Markdown.
	Invariants []model_logic.Logic // Invariants that must be true for this model.
	// Global functions that can be referenced from other expressions.
	GlobalFunctions map[identity.Key]model_logic.GlobalFunction
	// Named sets that can be referenced from behavioral logic.
	NamedSets map[identity.Key]model_named_set.NamedSet
	// Children
	Actors               map[identity.Key]model_actor.Actor
	ActorGeneralizations map[identity.Key]model_actor.Generalization
	Domains              map[identity.Key]model_domain.Domain
	DomainAssociations   map[identity.Key]model_domain.Association
	ClassAssociations    map[identity.Key]model_class.Association // Associations between classes that span domains.
}

func NewModel(key, name, details string, invariants []model_logic.Logic, globalFunctions map[identity.Key]model_logic.GlobalFunction, namedSets map[identity.Key]model_named_set.NamedSet) (model Model, err error) {
	model = Model{
		Key:             strings.TrimSpace(strings.ToLower(key)),
		Name:            name,
		Details:         details,
		Invariants:      invariants,
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
	}

	if err = model.Validate(); err != nil {
		return Model{}, err
	}

	return model, nil
}

// Validate validates the Model struct and all its children.
// This is the entry point for validating the entire model tree.
func (m *Model) Validate() error {
	if m.Key == "" {
		return coreerr.NewWithValues(coreerr.ModelKeyRequired, "Key is required", "Key", "", "non-empty string")
	}
	if m.Name == "" {
		return coreerr.NewWithValues(coreerr.ModelNameRequired, "Name is required", "Name", "", "non-empty string")
	}
	if err := m.validateInvariants(); err != nil {
		return err
	}
	if err := m.validateGlobalFunctions(); err != nil {
		return err
	}
	if err := m.validateNamedSets(); err != nil {
		return err
	}
	if err := m.validateActors(); err != nil {
		return err
	}
	if err := m.validateDomains(); err != nil {
		return err
	}
	if err := m.validateDomainAssociations(); err != nil {
		return err
	}
	if err := m.validateClassAssociations(); err != nil {
		return err
	}
	return nil
}

func (m *Model) validateInvariants() error {
	letTargets := make(map[string]bool)
	for i, inv := range m.Invariants {
		if err := inv.ValidateWithParent(nil); err != nil {
			return errors.Wrapf(err, "invariant %d", i)
		}
		if inv.Type != model_logic.LogicTypeAssessment && inv.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.ModelInvariantTypeInvalid,
				fmt.Sprintf("invariant %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, inv.Type),
				"Invariants", inv.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if inv.Type == model_logic.LogicTypeLet {
			if letTargets[inv.Target] {
				return coreerr.NewWithValues(coreerr.ModelInvariantDuplicateLet,
					fmt.Sprintf("invariant %d: duplicate let target %q", i, inv.Target),
					"Invariants", inv.Target, "")
			}
			letTargets[inv.Target] = true
		}
	}
	return nil
}

func (m *Model) validateGlobalFunctions() error {
	for gfKey, gf := range m.GlobalFunctions {
		if err := gf.ValidateWithParent(); err != nil {
			return errors.Wrapf(err, "global function '%s'", gfKey.String())
		}
		if gfKey != gf.Key {
			return coreerr.NewWithValues(coreerr.ModelGfuncKeyMismatch,
				fmt.Sprintf("global function map key '%s' does not match function key '%s'", gfKey.String(), gf.Key.String()),
				"GlobalFunctions", gfKey.String(), gf.Key.String())
		}
	}
	return nil
}

func (m *Model) validateNamedSets() error {
	for nsKey, ns := range m.NamedSets {
		if err := ns.ValidateWithParent(); err != nil {
			return errors.Wrapf(err, "named set '%s'", nsKey.String())
		}
		if nsKey != ns.Key {
			return coreerr.NewWithValues(coreerr.ModelNsetKeyMismatch,
				fmt.Sprintf("named set map key '%s' does not match named set key '%s'", nsKey.String(), ns.Key.String()),
				"NamedSets", nsKey.String(), ns.Key.String())
		}
	}
	return nil
}

func (m *Model) validateActors() error {
	actorGeneralizationKeys := make(map[identity.Key]bool)
	for agKey := range m.ActorGeneralizations {
		actorGeneralizationKeys[agKey] = true
	}
	for _, ag := range m.ActorGeneralizations {
		if err := ag.ValidateWithParent(nil); err != nil {
			return err
		}
	}
	for _, actor := range m.Actors {
		if err := actor.ValidateWithParent(nil); err != nil {
			return err
		}
		if err := actor.ValidateReferences(actorGeneralizationKeys); err != nil {
			return err
		}
	}
	for _, ag := range m.ActorGeneralizations {
		if err := m.validateGeneralizationUsage(ag); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) validateGeneralizationUsage(ag model_actor.Generalization) error {
	superCount := 0
	subCount := 0
	for _, actor := range m.Actors {
		if actor.SuperclassOfKey != nil && *actor.SuperclassOfKey == ag.Key {
			superCount++
		}
		if actor.SubclassOfKey != nil && *actor.SubclassOfKey == ag.Key {
			subCount++
		}
	}
	if superCount != 1 {
		return coreerr.NewWithValues(coreerr.ModelAgenSuperclassCount,
			fmt.Sprintf("actor generalization '%s' must have exactly one superclass, found %d", ag.Key.String(), superCount),
			"ActorGeneralizations", fmt.Sprintf("%d", superCount), "1")
	}
	if subCount < 1 {
		return coreerr.NewWithValues(coreerr.ModelAgenSubclassCount,
			fmt.Sprintf("actor generalization '%s' must have at least one subclass, found %d", ag.Key.String(), subCount),
			"ActorGeneralizations", fmt.Sprintf("%d", subCount), "at least 1")
	}
	return nil
}

func (m *Model) validateDomains() error {
	actorKeys := make(map[identity.Key]bool)
	for actorKey := range m.Actors {
		actorKeys[actorKey] = true
	}
	classKeys := m.buildClassKeys()
	for _, domain := range m.Domains {
		if err := domain.ValidateWithParentAndActorsAndClasses(nil, actorKeys, classKeys); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) buildClassKeys() map[identity.Key]bool {
	classKeys := make(map[identity.Key]bool)
	for _, domain := range m.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey := range subdomain.Classes {
				classKeys[classKey] = true
			}
		}
	}
	return classKeys
}

func (m *Model) validateDomainAssociations() error {
	domainKeys := make(map[identity.Key]bool)
	for domainKey := range m.Domains {
		domainKeys[domainKey] = true
	}
	for _, domainAssoc := range m.DomainAssociations {
		if err := domainAssoc.ValidateWithParent(nil); err != nil {
			return err
		}
		if err := domainAssoc.ValidateReferences(domainKeys); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) validateClassAssociations() error {
	classKeys := m.buildClassKeys()
	for _, classAssoc := range m.ClassAssociations {
		if err := classAssoc.ValidateWithParent(nil); err != nil {
			return err
		}
		if err := classAssoc.ValidateReferences(classKeys); err != nil {
			return err
		}
	}
	return nil
}

// SetClassAssociations sets the class associations for the model and routes them to domains.
// Associations with a domain (or subdomain within a domain) as parent are routed to that domain.
// Associations with no parent are kept at the model level.
// Associations with no parent that don't span domains return an error.
func (m *Model) SetClassAssociations(associations map[identity.Key]model_class.Association) error {
	// Initialize model-level associations map.
	modelAssociations := make(map[identity.Key]model_class.Association)

	// Group associations by their parent domain.
	domainAssociations := make(map[identity.Key]map[identity.Key]model_class.Association)
	for domainKey := range m.Domains {
		domainAssociations[domainKey] = make(map[identity.Key]model_class.Association)
	}

	for key, assoc := range associations {
		// Check if the association belongs to a domain (either directly or via subdomain).
		routedToDomain := false
		for domainKey := range m.Domains {
			if assoc.Key.IsParent(domainKey) {
				domainAssociations[domainKey][key] = assoc
				routedToDomain = true
				break
			}
		}

		if routedToDomain {
			continue
		}

		// Association doesn't belong to any domain - must be model-level (no parent).
		if !assoc.Key.HasNoParent() {
			return errors.Errorf("association '%s' has a parent that does not match any domain in the model", key.String())
		}

		// Model-level association - keep at model level.
		modelAssociations[key] = assoc
	}

	// Set model-level associations.
	m.ClassAssociations = modelAssociations

	// Route associations to domains.
	for domainKey, assocs := range domainAssociations {
		if len(assocs) > 0 {
			domain := m.Domains[domainKey]
			if err := domain.SetClassAssociations(assocs); err != nil {
				return err
			}
			m.Domains[domainKey] = domain
		}
	}

	return nil
}

// GetClassAssociations returns a copy of all class associations from this model and its domains.
func (m *Model) GetClassAssociations() map[identity.Key]model_class.Association {
	result := make(map[identity.Key]model_class.Association)
	// Add model-level associations.
	maps.Copy(result, m.ClassAssociations)
	// Add associations from all domains.
	for _, domain := range m.Domains {
		maps.Copy(result, domain.GetClassAssociations())
	}
	return result
}
