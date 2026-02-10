package req_model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key           string   // Models do not have keys like other entitites. They just need to be unique to other models in the system.
	Name          string
	Details       string   // Markdown.
	TlaInvariants []string // TLA+ expressions that must be true for this model.
	// Global TLA+ definitions that can be referenced from other TLA+ expressions.
	// Key is the definition Name (case-preserved, e.g., "_Max", "_SetOfValues").
	TlaDefinitions map[string]TlaDefinition
	// Children
	Actors             map[identity.Key]model_actor.Actor
	Domains            map[identity.Key]model_domain.Domain
	DomainAssociations map[identity.Key]model_domain.Association
	ClassAssociations  map[identity.Key]model_class.Association // Associations between classes that span domains.
}

func NewModel(key, name, details string, tlaInvariants []string, tlaDefinitions map[string]TlaDefinition) (model Model, err error) {

	model = Model{
		Key:            strings.TrimSpace(strings.ToLower(key)),
		Name:           name,
		Details:        details,
		TlaInvariants:  tlaInvariants,
		TlaDefinitions: tlaDefinitions,
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
	if err := validation.ValidateStruct(m,
		validation.Field(&m.Key, validation.Required),
		validation.Field(&m.Name, validation.Required),
	); err != nil {
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

	// Validate TLA definitions.
	for name, def := range m.TlaDefinitions {
		if err := def.Validate(); err != nil {
			return errors.Wrapf(err, "TLA definition '%s'", name)
		}
		// Ensure the map key matches the definition name.
		if name != def.Name {
			return errors.Errorf("TLA definition map key '%s' does not match definition name '%s'", name, def.Name)
		}
	}

	// Validate all children - they all have nil as their parent since Model
	// doesn't have an identity.Key.
	for _, actor := range m.Actors {
		if err := actor.ValidateWithParent(nil); err != nil {
			return err
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
