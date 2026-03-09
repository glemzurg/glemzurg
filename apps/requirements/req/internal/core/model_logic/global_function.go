package model_logic

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// GlobalFunction represents a global definition that can be referenced
// from other expressions throughout the model.
//
// Definitions can be:
//   - A set for membership checks: x \in _SetOfValues
//   - A function for data transformation: _Max(x, y)
//   - A common boolean predicate: _HasAChild(x)
//
// All global definitions must have a leading underscore to distinguish them
// from class-scoped actions.
type GlobalFunction struct {
	Key        identity.Key
	Name       string   // The definition name (e.g., _Max, _SetOfValues). Must start with underscore.
	Parameters []string // The parameter names (e.g., ["x", "y"] for _Max(x, y)).
	Logic      Logic    // The logic specification for this global function.
}

// NewGlobalFunction creates a new GlobalFunction.
func NewGlobalFunction(key identity.Key, name string, parameters []string, logic Logic) GlobalFunction {
	return GlobalFunction{
		Key:        key,
		Name:       name,
		Parameters: parameters,
		Logic:      logic,
	}
}

// Validate validates the GlobalFunction struct.
func (gf *GlobalFunction) Validate() error {
	// Validate the key.
	if err := gf.Key.Validate(); err != nil {
		return err
	}
	if gf.Key.KeyType != identity.KEY_TYPE_GLOBAL_FUNCTION {
		return coreerr.NewWithValues(coreerr.GfuncKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for global function", gf.Key.KeyType), "Key.KeyType", gf.Key.KeyType, identity.KEY_TYPE_GLOBAL_FUNCTION)
	}

	// Validate the specification logic.
	if err := gf.Logic.Validate(); err != nil {
		return err
	}

	// Logic must use the global function's exact key.
	if gf.Logic.Key != gf.Key {
		return coreerr.NewWithValues(coreerr.GfuncLogicKeyMismatch, fmt.Sprintf("logic key '%s' does not match global function key '%s'", gf.Logic.Key.String(), gf.Key.String()), "Logic.Key", gf.Logic.Key.String(), gf.Key.String())
	}

	// Global function logic must be of kind "value".
	if gf.Logic.Type != LogicTypeValue {
		return coreerr.NewWithValues(coreerr.GfuncLogicTypeInvalid, fmt.Sprintf("global function logic kind must be '%s', got '%s'", LogicTypeValue, gf.Logic.Type), "Logic.Type", gf.Logic.Type, LogicTypeValue)
	}

	// Name is required.
	if gf.Name == "" {
		return coreerr.New(coreerr.GfuncNameRequired, "Name is required", "Name")
	}

	// Name must start with underscore.
	if !strings.HasPrefix(gf.Name, "_") {
		return coreerr.NewWithValues(coreerr.GfuncNameNoUnderscore, fmt.Sprintf("global function name '%s' must start with underscore", gf.Name), "Name", gf.Name, "name starting with '_'")
	}

	return nil
}

// ValidateWithParent validates the GlobalFunction and its key's parent relationship.
// Global function keys are root-level (nil parent).
func (gf *GlobalFunction) ValidateWithParent() error {
	if err := gf.Validate(); err != nil {
		return err
	}
	if err := gf.Key.ValidateParent(nil); err != nil {
		return err
	}
	// Validate the logic's key parent relationship.
	// The spec shares the global function's exact key (root-level, nil parent).
	if err := gf.Logic.ValidateWithParent(nil); err != nil {
		return errors.Wrap(err, "specification")
	}
	return nil
}
