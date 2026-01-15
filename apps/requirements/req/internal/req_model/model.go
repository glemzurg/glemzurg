package req_model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

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

	err = validation.ValidateStruct(&model,
		validation.Field(&model.Key, validation.Required),
		validation.Field(&model.Name, validation.Required),
	)
	if err != nil {
		return Model{}, errors.WithStack(err)
	}

	return model, nil
}

// ValidateWithParent validates the Model and all its children.
// This is the entry point for validating the entire model tree.
// For Model, parent should always be nil.
func (m *Model) ValidateWithParent() error {
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
	// DomainAssociations are validated when Domains are validated.
	// Model-level Associations (spanning domains) have nil parent.
	for i := range m.Associations {
		if err := m.Associations[i].ValidateWithParent(nil); err != nil {
			return err
		}
	}
	return nil
}
