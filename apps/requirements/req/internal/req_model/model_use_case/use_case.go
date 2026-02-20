package model_use_case

import (
	"errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	pkgerrors "github.com/pkg/errors"
)

const (
	_USE_CASE_LEVEL_SKY = "sky" // A high-level organizational user story.
	_USE_CASE_LEVEL_SEA = "sea" // A straight forward user story.
	_USE_CASE_LEVEL_MUD = "mud" // A small, reusable part of another user story (e.g. a login flow).
)

// UseCase is a user story for the system.
type UseCase struct {
	Key             identity.Key
	Name            string        `validate:"required"`
	Details         string        // Markdown.
	Level           string        `validate:"required,oneof=sky sea mud"` // How high cocept or tightly focused the user case is.
	ReadOnly        bool          // This is a user story that does not change the state of the system.
	SuperclassOfKey *identity.Key // If this use case is part of a generalization as the superclass.
	SubclassOfKey   *identity.Key // If this use case is part of a generalization as a subclass.
	UmlComment      string
	// Children
	Actors    map[identity.Key]Actor
	Scenarios map[identity.Key]model_scenario.Scenario
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
	// Validate the key.
	if err := uc.Key.Validate(); err != nil {
		return err
	}
	if uc.Key.KeyType != identity.KEY_TYPE_USE_CASE {
		return errors.New("invalid key type for use_case")
	}
	// Validate struct tags (Name required, Level required+oneof).
	if err := _validate.Struct(uc); err != nil {
		return err
	}
	return nil
}

func (uc *UseCase) SetActors(actors map[identity.Key]Actor) {
	uc.Actors = actors
}

func (uc *UseCase) SetScenarios(scenarios map[identity.Key]model_scenario.Scenario) {
	uc.Scenarios = scenarios
}

// ValidateWithParent validates the UseCase, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (uc *UseCase) ValidateWithParent(parent *identity.Key) error {
	return uc.ValidateWithParentAndClasses(parent, nil, nil)
}

// ValidateWithParentAndClasses validates the UseCase with access to classes for cross-reference validation.
// The parent must be a Subdomain.
// The classes map is used to validate that scenario object ClassKey references exist.
// The actorClasses map contains class keys that have an ActorKey defined (i.e., classes that represent actors).
func (uc *UseCase) ValidateWithParentAndClasses(parent *identity.Key, classes map[identity.Key]bool, actorClasses map[identity.Key]bool) error {
	// Validate the object itself.
	if err := uc.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := uc.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	// Validate that each actor key references a class that has an ActorKey defined.
	for actorClassKey, actor := range uc.Actors {
		if err := actor.ValidateWithParent(); err != nil {
			return err
		}
		if !actorClasses[actorClassKey] {
			return pkgerrors.Errorf("use case '%s' actor references class '%s' which is not an actor class (no ActorKey defined)", uc.Key.String(), actorClassKey.String())
		}
	}
	for _, scenario := range uc.Scenarios {
		if err := scenario.ValidateWithParentAndClasses(&uc.Key, classes); err != nil {
			return err
		}
	}
	return nil
}
