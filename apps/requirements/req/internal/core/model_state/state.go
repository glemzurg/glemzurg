package model_state

import (
	"fmt"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Children
	Actions []StateAction
}

func NewState(key identity.Key, name, details, umlComment string) State {
	return State{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}
}

// Validate validates the State struct.
func (s *State) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := s.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.StateKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if s.Key.KeyType != identity.KEY_TYPE_STATE {
		return coreerr.NewWithValues(ctx, coreerr.StateKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for state", s.Key.KeyType), "Key", s.Key.KeyType, identity.KEY_TYPE_STATE)
	}

	if s.Name == "" {
		return coreerr.New(ctx, coreerr.StateNameRequired, "Name is required", "Name")
	}
	if badChar := coreerr.ValidateNameChars(s.Name); badChar != "" {
		return coreerr.NewWithValues(ctx, coreerr.StateNameInvalidChars, fmt.Sprintf("Name contains invalid character %q", badChar), "Name", s.Name, "A-Za-z0-9 space hyphen underscore")
	}

	return nil
}

func (s *State) SetActions(actions []StateAction) {
	sort.Slice(actions, func(i, j int) bool {
		return lessThanStateAction(actions[i], actions[j])
	})

	s.Actions = actions
}

// ValidateWithParent validates the State, its key's parent relationship, and all children.
// The parent must be a Class.
func (s *State) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	return s.ValidateWithParentAndActions(ctx, parent, nil)
}

// ValidateWithParentAndActions validates the State with access to actions for cross-reference validation.
// The parent must be a Class.
// The actions map is used to validate that StateAction ActionKey references exist.
func (s *State) ValidateWithParentAndActions(ctx *coreerr.ValidationContext, parent *identity.Key, actions map[identity.Key]bool) error {
	// Validate the object itself.
	if err := s.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range s.Actions {
		childCtx := ctx.Child("stateAction", s.Actions[i].Key.String())
		if err := s.Actions[i].ValidateWithParent(childCtx, &s.Key); err != nil {
			return err
		}
		if err := s.Actions[i].ValidateReferences(childCtx, actions); err != nil {
			return err
		}
	}
	return nil
}
