package req_model

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key        string              `validate:"required"` // Models do not have keys like other entitites. They just need to be unique to other models in the system.
	Name       string              `validate:"required"`
	Details    string              // Markdown.
	Invariants []model_logic.Logic // Invariants that must be true for this model.
	// Global functions that can be referenced from other expressions.
	GlobalFunctions map[identity.Key]model_logic.GlobalFunction
	// Children
	Actors               map[identity.Key]model_actor.Actor
	ActorGeneralizations map[identity.Key]model_actor.Generalization
	Domains              map[identity.Key]model_domain.Domain
	DomainAssociations   map[identity.Key]model_domain.Association
	ClassAssociations    map[identity.Key]model_class.Association // Associations between classes that span domains.
}

func NewModel(key, name, details string, invariants []model_logic.Logic, globalFunctions map[identity.Key]model_logic.GlobalFunction) (model Model, err error) {

	model = Model{
		Key:             strings.TrimSpace(strings.ToLower(key)),
		Name:            name,
		Details:         details,
		Invariants:      invariants,
		GlobalFunctions: globalFunctions,
	}

	if err = model.Validate(); err != nil {
		return Model{}, err
	}

	return model, nil
}

// Validate validates the Model struct and all its children.
// This is the entry point for validating the entire model tree.
func (m *Model) Validate() error {
	// Validate the model's own fields.
	if err := _validate.Struct(m); err != nil {
		return err
	}

	// Build a set of actor keys for reference validation.
	actorKeys := make(map[identity.Key]bool)
	for actorKey := range m.Actors {
		actorKeys[actorKey] = true
	}

	// Build a set of domain keys for domain association reference validation.
	domainKeys := make(map[identity.Key]bool)
	for domainKey := range m.Domains {
		domainKeys[domainKey] = true
	}

	// Build a set of all class keys in the model for association reference validation.
	// Classes only exist in subdomains.
	classKeys := make(map[identity.Key]bool)
	for _, domain := range m.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey := range subdomain.Classes {
				classKeys[classKey] = true
			}
		}
	}

	// Validate invariants.
	for i, inv := range m.Invariants {
		if err := inv.ValidateWithParent(nil); err != nil {
			return errors.Wrapf(err, "invariant %d", i)
		}
		if inv.Type != model_logic.LogicTypeAssessment {
			return errors.Errorf("invariant %d: logic kind must be '%s', got '%s'", i, model_logic.LogicTypeAssessment, inv.Type)
		}
	}

	// Validate global functions.
	for gfKey, gf := range m.GlobalFunctions {
		if err := gf.ValidateWithParent(); err != nil {
			return errors.Wrapf(err, "global function '%s'", gfKey.String())
		}
		// Ensure the map key matches the function key.
		if gfKey != gf.Key {
			return errors.Errorf("global function map key '%s' does not match function key '%s'", gfKey.String(), gf.Key.String())
		}
	}

	// Build a set of actor generalization keys for reference validation.
	actorGeneralizationKeys := make(map[identity.Key]bool)
	for agKey := range m.ActorGeneralizations {
		actorGeneralizationKeys[agKey] = true
	}

	// Validate actor generalizations.
	for _, ag := range m.ActorGeneralizations {
		if err := ag.ValidateWithParent(nil); err != nil {
			return err
		}
	}

	// Validate all children - they all have nil as their parent since Model
	// doesn't have an identity.Key.
	for _, actor := range m.Actors {
		if err := actor.ValidateWithParent(nil); err != nil {
			return err
		}
		if err := actor.ValidateReferences(actorGeneralizationKeys); err != nil {
			return err
		}
	}

	// Check that each actor generalization is in use by exactly one superclass and at least one subclass.
	for _, ag := range m.ActorGeneralizations {
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
			return errors.Errorf("actor generalization '%s' must have exactly one superclass, found %d", ag.Key.String(), superCount)
		}
		if subCount < 1 {
			return errors.Errorf("actor generalization '%s' must have at least one subclass, found %d", ag.Key.String(), subCount)
		}
	}
	for _, domain := range m.Domains {
		if err := domain.ValidateWithParentAndActorsAndClasses(nil, actorKeys, classKeys); err != nil {
			return err
		}
	}
	// DomainAssociations need to be validated.
	for _, domainAssoc := range m.DomainAssociations {
		if err := domainAssoc.ValidateWithParent(nil); err != nil {
			return err
		}
		if err := domainAssoc.ValidateReferences(domainKeys); err != nil {
			return err
		}
	}
	// Model-level Associations (spanning domains) have nil parent.
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
	for k, v := range m.ClassAssociations {
		result[k] = v
	}
	// Add associations from all domains.
	for _, domain := range m.Domains {
		for k, v := range domain.GetClassAssociations() {
			result[k] = v
		}
	}
	return result
}
