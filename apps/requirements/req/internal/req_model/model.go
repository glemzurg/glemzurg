package req_model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key     string // Models do not have keys like other entitites. They just need to be unique to other models in the system.
	Name    string
	Details string // Markdown.
	// Children
	Actors             map[identity.Key]model_actor.Actor
	Domains            map[identity.Key]model_domain.Domain
	DomainAssociations map[identity.Key]model_domain.Association
	ClassAssociations  map[identity.Key]model_class.Association // Associations between classes that span domains.
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
	for _, actor := range m.Actors {
		if err := actor.ValidateWithParent(nil); err != nil {
			return err
		}
	}
	for _, domain := range m.Domains {
		if err := domain.ValidateWithParent(nil); err != nil {
			return err
		}
	}
	// DomainAssociations need to be validated.
	for _, domainAssoc := range m.DomainAssociations {
		if err := domainAssoc.ValidateWithParent(nil); err != nil {
			return err
		}
	}
	// Model-level Associations (spanning domains) have nil parent.
	for _, classAssoc := range m.ClassAssociations {
		if err := classAssoc.ValidateWithParent(nil); err != nil {
			return err
		}
	}
	return nil
}
