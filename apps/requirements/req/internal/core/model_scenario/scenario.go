package model_scenario

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Scenario is a documented scenario for a use case, such as a sequence diagram.
type Scenario struct {
	Key     identity.Key
	Name    string
	Details string // Markdown.
	// Children
	Steps   *Step // The "abstract syntax tree" of the scenario.
	Objects map[identity.Key]Object
}

func NewScenario(key identity.Key, name, details string) Scenario {
	return Scenario{
		Key:     key,
		Name:    name,
		Details: details,
	}
}

// Validate validates the Scenario struct.
func (s *Scenario) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := s.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.ScenarioKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if s.Key.KeyType != identity.KEY_TYPE_SCENARIO {
		return coreerr.NewWithValues(ctx, coreerr.ScenarioKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for scenario", s.Key.KeyType), "Key", s.Key.KeyType, identity.KEY_TYPE_SCENARIO)
	}
	// Validate Name required.
	if s.Name == "" {
		return coreerr.New(ctx, coreerr.ScenarioNameRequired, "Name is required", "Name")
	}
	return nil
}

func (s *Scenario) SetObjects(objects map[identity.Key]Object) {
	s.Objects = objects
}

// ValidateWithParent validates the Scenario, its key's parent relationship, and all children.
// The parent must be a UseCase.
func (s *Scenario) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	return s.ValidateWithParentAndClasses(ctx, parent, nil)
}

// ValidateWithParentAndClasses validates the Scenario with access to classes for cross-reference validation.
// The parent must be a UseCase.
// The classes map is used to validate that Object ClassKey references exist.
func (s *Scenario) ValidateWithParentAndClasses(ctx *coreerr.ValidationContext, parent *identity.Key, classes map[identity.Key]bool) error {
	// Validate the object itself.
	if err := s.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Validate all children.
	for _, obj := range s.Objects {
		childCtx := ctx.Child("object", obj.Key.String())
		if err := obj.ValidateWithParent(childCtx, &s.Key); err != nil {
			return err
		}
		if err := obj.ValidateReferences(childCtx, classes); err != nil {
			return err
		}
	}
	// Validate Steps if there is content.
	if s.Steps != nil {
		childCtx := ctx.Child("steps", s.Key.String())
		if err := s.Steps.ValidateWithParent(childCtx, &s.Key); err != nil {
			return err
		}
	}
	return nil
}
