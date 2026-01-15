package req_model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
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
