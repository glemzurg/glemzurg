package model_scenario

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Scenario is a documented scenario for a use case, such as a sequence diagram.
type Scenario struct {
	Key     identity.Key
	Name    string `validate:"required"`
	Details string // Markdown.
	// Children
	Steps   *Node // The "abstract syntax tree" of the scenario.
	Objects map[identity.Key]Object
}

func NewScenario(key identity.Key, name, details string) (scenario Scenario, err error) {

	scenario = Scenario{
		Key:     key,
		Name:    name,
		Details: details,
	}

	if err = scenario.Validate(); err != nil {
		return Scenario{}, err
	}

	return scenario, nil
}

// Validate validates the Scenario struct.
func (s *Scenario) Validate() error {
	// Validate the key.
	if err := s.Key.Validate(); err != nil {
		return err
	}
	if s.Key.KeyType != identity.KEY_TYPE_SCENARIO {
		return errors.Errorf("Key: invalid key type '%s' for scenario.", s.Key.KeyType)
	}
	// Validate struct tags (Name required).
	if err := _validate.Struct(s); err != nil {
		return err
	}
	return nil
}

func (sc *Scenario) SetObjects(objects map[identity.Key]Object) {
	sc.Objects = objects
}

// ValidateWithParent validates the Scenario, its key's parent relationship, and all children.
// The parent must be a UseCase.
func (s *Scenario) ValidateWithParent(parent *identity.Key) error {
	return s.ValidateWithParentAndClasses(parent, nil)
}

// ValidateWithParentAndClasses validates the Scenario with access to classes for cross-reference validation.
// The parent must be a UseCase.
// The classes map is used to validate that Object ClassKey references exist.
func (s *Scenario) ValidateWithParentAndClasses(parent *identity.Key, classes map[identity.Key]bool) error {
	// Validate the object itself.
	if err := s.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for _, obj := range s.Objects {
		if err := obj.ValidateWithParent(&s.Key); err != nil {
			return err
		}
		if err := obj.ValidateReferences(classes); err != nil {
			return err
		}
	}
	// Validate Steps if there is content.
	if s.Steps != nil {
		if err := s.Steps.ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	return nil
}
