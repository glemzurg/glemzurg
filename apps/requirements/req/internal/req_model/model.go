package req_model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key     string // Models do not have keys like other entitites. They just need to be unique to other models in the system.
	Name    string
	Details string // Markdown.
	// Data in a parsed file.
	Actors             []model_actor.Actor
	Domains            []model_domain.Domain
	DomainAssociations []model_domain.Association
	Associations       []model_class.Association // Associations between classes that span domains.
}

func NewModel(key, name, details string) (model Model, err error) {

	model = Model{
		Key:     strings.TrimSpace(strings.ToLower(key)),
		Name:    name,
		Details: details,
	}

	if err = model.Validate(); err != nil {
		return Model{}, err
	}

	return model, nil
}

// Validate validates the Model struct.
func (m *Model) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Key, validation.Required),
		validation.Field(&m.Name, validation.Required),
	)
}

// ValidateWithParent validates the Model and all its children.
// This is the entry point for validating the entire model tree.
// For Model, parent should always be nil.
func (m *Model) ValidateWithParent() error {
	// Validate the model itself.
	if err := m.Validate(); err != nil {
		return err
	}
	// Validate all children - they all have nil as their parent since Model
	// doesn't have an identity.Key.
	for i := range m.Actors {
		if err := m.Actors[i].ValidateWithParent(nil); err != nil {
			return err
		}
	}
	for i := range m.Domains {
		if err := m.Domains[i].ValidateWithParent(nil); err != nil {
			return err
		}
	}
	// DomainAssociations need to be validated.
	for i := range m.DomainAssociations {
		if err := m.DomainAssociations[i].ValidateWithParent(nil); err != nil {
			return err
		}
	}
	// Model-level Associations (spanning domains) have nil parent.
	for i := range m.Associations {
		if err := m.Associations[i].ValidateWithParent(nil); err != nil {
			return err
		}
	}
	return nil
}
