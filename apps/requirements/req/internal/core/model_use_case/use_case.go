package model_use_case

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

const (
	_USE_CASE_LEVEL_SKY = "sky" // A high-level organizational user story.
	_USE_CASE_LEVEL_SEA = "sea" // A straight forward user story.
	_USE_CASE_LEVEL_MUD = "mud" // A small, reusable part of another user story (e.g. a login flow).
)

// UseCase is a user story for the system.
type UseCase struct {
	Key             identity.Key
	Name            string
	Details         string        // Markdown.
	Level           string        // How high cocept or tightly focused the user case is.
	ReadOnly        bool          // This is a user story that does not change the state of the system.
	SuperclassOfKey *identity.Key // If this use case is part of a generalization as the superclass.
	SubclassOfKey   *identity.Key // If this use case is part of a generalization as a subclass.
	UmlComment      string
	// Children
	Actors    map[identity.Key]Actor
	Scenarios map[identity.Key]model_scenario.Scenario
}

// GeneralizationRefs holds the optional generalization references for a use case.
type GeneralizationRefs struct {
	SuperclassOfKey *identity.Key
	SubclassOfKey   *identity.Key
}

func NewUseCase(key identity.Key, name, details, level string, readOnly bool, genRefs GeneralizationRefs, umlComment string) (useCase UseCase, err error) {
	useCase = UseCase{
		Key:             key,
		Name:            name,
		Details:         details,
		Level:           level,
		ReadOnly:        readOnly,
		SuperclassOfKey: genRefs.SuperclassOfKey,
		SubclassOfKey:   genRefs.SubclassOfKey,
		UmlComment:      umlComment,
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
		return coreerr.New(coreerr.UcKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if uc.Key.KeyType != identity.KEY_TYPE_USE_CASE {
		return coreerr.NewWithValues(coreerr.UcKeyTypeInvalid, "invalid key type for use_case", "Key", uc.Key.KeyType, identity.KEY_TYPE_USE_CASE)
	}
	// Validate Name required.
	if uc.Name == "" {
		return coreerr.New(coreerr.UcNameRequired, "Name is required", "Name")
	}
	// Validate Level required.
	if uc.Level == "" {
		return coreerr.New(coreerr.UcLevelRequired, "Level is required", "Level")
	}
	// Validate Level is one of valid values.
	if uc.Level != _USE_CASE_LEVEL_SKY && uc.Level != _USE_CASE_LEVEL_SEA && uc.Level != _USE_CASE_LEVEL_MUD {
		return coreerr.NewWithValues(coreerr.UcLevelInvalid, "Level must be one of: sky, sea, mud", "Level", uc.Level, "one of: sky, sea, mud")
	}
	// Validate FK key types.
	if uc.SuperclassOfKey != nil {
		if err := uc.SuperclassOfKey.Validate(); err != nil {
			return coreerr.New(coreerr.UcSuperkeyInvalid, fmt.Sprintf("SuperclassOfKey: %s", err.Error()), "SuperclassOfKey")
		}
		if uc.SuperclassOfKey.KeyType != identity.KEY_TYPE_USE_CASE_GENERALIZATION {
			return coreerr.NewWithValues(coreerr.UcSuperkeyTypeInvalid, fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for use case generalization", uc.SuperclassOfKey.KeyType), "SuperclassOfKey", uc.SuperclassOfKey.KeyType, identity.KEY_TYPE_USE_CASE_GENERALIZATION)
		}
	}
	if uc.SubclassOfKey != nil {
		if err := uc.SubclassOfKey.Validate(); err != nil {
			return coreerr.New(coreerr.UcSubkeyInvalid, fmt.Sprintf("SubclassOfKey: %s", err.Error()), "SubclassOfKey")
		}
		if uc.SubclassOfKey.KeyType != identity.KEY_TYPE_USE_CASE_GENERALIZATION {
			return coreerr.NewWithValues(coreerr.UcSubkeyTypeInvalid, fmt.Sprintf("SubclassOfKey: invalid key type '%s' for use case generalization", uc.SubclassOfKey.KeyType), "SubclassOfKey", uc.SubclassOfKey.KeyType, identity.KEY_TYPE_USE_CASE_GENERALIZATION)
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if uc.SuperclassOfKey != nil && uc.SubclassOfKey != nil && *uc.SuperclassOfKey == *uc.SubclassOfKey {
		return coreerr.New(coreerr.UcSuperSubSame, "SuperclassOfKey and SubclassOfKey cannot be the same", "SuperclassOfKey")
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
			return coreerr.NewWithValues(coreerr.UcActorNotActorClass, fmt.Sprintf("use case '%s' actor references class '%s' which is not an actor class (no ActorKey defined)", uc.Key.String(), actorClassKey.String()), "Actors", actorClassKey.String(), "")
		}
	}
	for _, scenario := range uc.Scenarios {
		if err := scenario.ValidateWithParentAndClasses(&uc.Key, classes); err != nil {
			return err
		}
	}
	return nil
}

// ValidateReferences validates that the use case's reference keys point to valid entities.
// - SuperclassOfKey must exist in the generalizations map
// - SubclassOfKey must exist in the generalizations map.
func (uc *UseCase) ValidateReferences(generalizations map[identity.Key]bool) error {
	if uc.SuperclassOfKey != nil {
		if !generalizations[*uc.SuperclassOfKey] {
			return coreerr.NewWithValues(coreerr.UcSupergenNotfound, fmt.Sprintf("use case '%s' references non-existent generalization '%s'", uc.Key.String(), uc.SuperclassOfKey.String()), "SuperclassOfKey", uc.SuperclassOfKey.String(), "")
		}
	}
	if uc.SubclassOfKey != nil {
		if !generalizations[*uc.SubclassOfKey] {
			return coreerr.NewWithValues(coreerr.UcSubgenNotfound, fmt.Sprintf("use case '%s' references non-existent generalization '%s'", uc.Key.String(), uc.SubclassOfKey.String()), "SubclassOfKey", uc.SubclassOfKey.String(), "")
		}
	}
	return nil
}
