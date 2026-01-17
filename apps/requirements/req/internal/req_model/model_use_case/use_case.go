package model_use_case

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
)

const (
	_USE_CASE_LEVEL_SKY = "sky" // A high-level organizational user story.
	_USE_CASE_LEVEL_SEA = "sea" // A straight forward user story.
	_USE_CASE_LEVEL_MUD = "mud" // A small, reusable part of another user story (e.g. a login flow).
)

// UseCase is a user story for the system.
type UseCase struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	Level      string // How high cocept or tightly focused the user case is.
	ReadOnly   bool   // This is a user story that does not change the state of the system.
	UmlComment string
	// Children
	Actors    map[identity.Key]Actor
	Scenarios []model_scenario.Scenario
}

func NewUseCase(key identity.Key, name, details, level string, readOnly bool, umlComment string) (useCase UseCase, err error) {

	useCase = UseCase{
		Key:        key,
		Name:       name,
		Details:    details,
		Level:      level,
		ReadOnly:   readOnly,
		UmlComment: umlComment,
	}

	if err = useCase.Validate(); err != nil {
		return UseCase{}, err
	}

	return useCase, nil
}

// Validate validates the UseCase struct.
func (uc *UseCase) Validate() error {
	return validation.ValidateStruct(uc,
		validation.Field(&uc.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_USE_CASE {
				return errors.New("invalid key type for use_case")
			}
			return nil
		})),
		validation.Field(&uc.Name, validation.Required),
		validation.Field(&uc.Level, validation.Required, validation.In(_USE_CASE_LEVEL_SKY, _USE_CASE_LEVEL_SEA, _USE_CASE_LEVEL_MUD)),
	)
}

func (uc *UseCase) SetDomainKey(domainKey identity.Key) {
	uc.DomainKey = domainKey
}

func (uc *UseCase) SetActors(actors map[identity.Key]Actor) {
	uc.Actors = actors
}

func (uc *UseCase) SetScenarios(scenarios []model_scenario.Scenario) {
	uc.Scenarios = scenarios
}

// ValidateWithParent validates the UseCase, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (uc *UseCase) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := uc.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := uc.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range uc.Scenarios {
		if err := uc.Scenarios[i].ValidateWithParent(&uc.Key); err != nil {
			return err
		}
	}
	return nil
}
