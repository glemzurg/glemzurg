package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key     string
	Name    string
	Details string // Markdown.
}

func NewModel(key, name, details string) (model Model, err error) {

	model = Model{
		Key:     key,
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
