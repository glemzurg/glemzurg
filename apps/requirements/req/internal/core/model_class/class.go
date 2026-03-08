package model_class

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Class is a thing in the system.
type Class struct {
	Key             identity.Key
	Name            string
	Details         string        // Markdown.
	ActorKey        *identity.Key // If this class is an Actor this is the key of that actor.
	SuperclassOfKey *identity.Key // If this class is part of a generalization as the superclass.
	SubclassOfKey   *identity.Key // If this class is part of a generalization as a subclass.
	UmlComment      string
	// Children
	Invariants  []model_logic.Logic        // Invariants that must be true for all objects of this class.
	Attributes  map[identity.Key]Attribute // The attributes of a class.
	States      map[identity.Key]model_state.State
	Events      map[identity.Key]model_state.Event
	Guards      map[identity.Key]model_state.Guard
	Actions     map[identity.Key]model_state.Action
	Queries     map[identity.Key]model_state.Query
	Transitions map[identity.Key]model_state.Transition
}

func NewClass(key identity.Key, name, details string, actorKey, superclassOfKey, subclassOfKey *identity.Key, umlComment string) (class Class, err error) {
	class = Class{
		Key:             key,
		Name:            name,
		Details:         details,
		ActorKey:        actorKey,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      umlComment,
	}

	if err = class.Validate(); err != nil {
		return Class{}, err
	}

	return class, nil
}

// Validate validates the Class struct.
func (c *Class) Validate() error {
	// Validate the key.
	if err := c.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ClassKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if c.Key.KeyType != identity.KEY_TYPE_CLASS {
		return &coreerr.ValidationError{
			Code:    coreerr.ClassKeyTypeInvalid,
			Message: fmt.Sprintf("key: invalid key type '%s' for class", c.Key.KeyType),
			Field:   "Key",
			Got:     c.Key.KeyType,
			Want:    identity.KEY_TYPE_CLASS,
		}
	}

	// Name is required.
	if c.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ClassNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}

	// Validate FK key types.
	if c.ActorKey != nil {
		if err := c.ActorKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassActorkeyInvalid,
				Message: fmt.Sprintf("ActorKey: %s", err.Error()),
				Field:   "ActorKey",
			}
		}
		if c.ActorKey.KeyType != identity.KEY_TYPE_ACTOR {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassActorkeyTypeInvalid,
				Message: fmt.Sprintf("ActorKey: invalid key type '%s' for actor", c.ActorKey.KeyType),
				Field:   "ActorKey",
				Got:     c.ActorKey.KeyType,
				Want:    identity.KEY_TYPE_ACTOR,
			}
		}
	}
	if c.SuperclassOfKey != nil {
		if err := c.SuperclassOfKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSuperkeyInvalid,
				Message: fmt.Sprintf("SuperclassOfKey: %s", err.Error()),
				Field:   "SuperclassOfKey",
			}
		}
		if c.SuperclassOfKey.KeyType != identity.KEY_TYPE_CLASS_GENERALIZATION {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSuperkeyTypeInvalid,
				Message: fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for class generalization", c.SuperclassOfKey.KeyType),
				Field:   "SuperclassOfKey",
				Got:     c.SuperclassOfKey.KeyType,
				Want:    identity.KEY_TYPE_CLASS_GENERALIZATION,
			}
		}
	}
	if c.SubclassOfKey != nil {
		if err := c.SubclassOfKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSubkeyInvalid,
				Message: fmt.Sprintf("SubclassOfKey: %s", err.Error()),
				Field:   "SubclassOfKey",
			}
		}
		if c.SubclassOfKey.KeyType != identity.KEY_TYPE_CLASS_GENERALIZATION {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSubkeyTypeInvalid,
				Message: fmt.Sprintf("SubclassOfKey: invalid key type '%s' for class generalization", c.SubclassOfKey.KeyType),
				Field:   "SubclassOfKey",
				Got:     c.SubclassOfKey.KeyType,
				Want:    identity.KEY_TYPE_CLASS_GENERALIZATION,
			}
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if c.SuperclassOfKey != nil && c.SubclassOfKey != nil && *c.SuperclassOfKey == *c.SubclassOfKey {
		return &coreerr.ValidationError{
			Code:    coreerr.ClassSuperSubSame,
			Message: "SuperclassOfKey and SubclassOfKey cannot be the same",
			Field:   "SuperclassOfKey",
		}
	}
	return nil
}

// ValidateReferences validates that the class's reference keys point to valid entities.
// - ActorKey must exist in the actors map
// - SuperclassOfKey must exist in the generalizations map and be in the same subdomain
// - SubclassOfKey must exist in the generalizations map and be in the same subdomain.
func (c *Class) ValidateReferences(actors map[identity.Key]bool, generalizations map[identity.Key]bool) error {
	// Validate ActorKey references a real actor.
	if c.ActorKey != nil {
		if !actors[*c.ActorKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassActorNotfound,
				Message: fmt.Sprintf("class '%s' references non-existent actor '%s'", c.Key.String(), c.ActorKey.String()),
				Field:   "ActorKey",
				Got:     c.ActorKey.String(),
			}
		}
	}

	// Get this class's subdomain from its parent key.
	classSubdomainKey := c.Key.ParentKey

	// Validate SuperclassOfKey references a real generalization in the same subdomain.
	if c.SuperclassOfKey != nil {
		if !generalizations[*c.SuperclassOfKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSupergenNotfound,
				Message: fmt.Sprintf("class '%s' references non-existent generalization '%s'", c.Key.String(), c.SuperclassOfKey.String()),
				Field:   "SuperclassOfKey",
				Got:     c.SuperclassOfKey.String(),
			}
		}
		// Check same subdomain.
		generalizationSubdomainKey := c.SuperclassOfKey.ParentKey
		if classSubdomainKey != generalizationSubdomainKey {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSupergenWrongSubdomain,
				Message: fmt.Sprintf("class '%s' generalization '%s' must be in the same subdomain", c.Key.String(), c.SuperclassOfKey.String()),
				Field:   "SuperclassOfKey",
				Got:     c.SuperclassOfKey.String(),
			}
		}
	}

	// Validate SubclassOfKey references a real generalization in the same subdomain.
	if c.SubclassOfKey != nil {
		if !generalizations[*c.SubclassOfKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSubgenNotfound,
				Message: fmt.Sprintf("class '%s' references non-existent generalization '%s'", c.Key.String(), c.SubclassOfKey.String()),
				Field:   "SubclassOfKey",
				Got:     c.SubclassOfKey.String(),
			}
		}
		// Check same subdomain.
		generalizationSubdomainKey := c.SubclassOfKey.ParentKey
		if classSubdomainKey != generalizationSubdomainKey {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassSubgenWrongSubdomain,
				Message: fmt.Sprintf("class '%s' generalization '%s' must be in the same subdomain", c.Key.String(), c.SubclassOfKey.String()),
				Field:   "SubclassOfKey",
				Got:     c.SubclassOfKey.String(),
			}
		}
	}

	return nil
}

func (c *Class) SetInvariants(invariants []model_logic.Logic) {
	c.Invariants = invariants
}

func (c *Class) SetAttributes(attributes map[identity.Key]Attribute) {
	c.Attributes = attributes
}

func (c *Class) SetStates(states map[identity.Key]model_state.State) {
	c.States = states
}

func (c *Class) SetEvents(events map[identity.Key]model_state.Event) {
	c.Events = events
}

func (c *Class) SetGuards(guards map[identity.Key]model_state.Guard) {
	c.Guards = guards
}

func (c *Class) SetActions(actions map[identity.Key]model_state.Action) {
	c.Actions = actions
}

func (c *Class) SetQueries(queries map[identity.Key]model_state.Query) {
	c.Queries = queries
}

func (c *Class) SetTransitions(transitions map[identity.Key]model_state.Transition) {
	c.Transitions = transitions
}

// ValidateWithParent validates the Class, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (c *Class) ValidateWithParent(parent *identity.Key) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := c.Key.ValidateParent(parent); err != nil {
		return err
	}
	if err := c.validateClassInvariants(); err != nil {
		return err
	}
	if err := c.validateClassChildren(); err != nil {
		return err
	}
	if err := c.validateActionGuarantees(); err != nil {
		return err
	}
	if err := c.validateTransitions(); err != nil {
		return err
	}
	return nil
}

func (c *Class) validateClassInvariants() error {
	letTargets := make(map[string]bool)
	for i, inv := range c.Invariants {
		if err := inv.ValidateWithParent(&c.Key); err != nil {
			return errors.Wrapf(err, "invariant %d", i)
		}
		if inv.Type != model_logic.LogicTypeAssessment && inv.Type != model_logic.LogicTypeLet {
			return &coreerr.ValidationError{
				Code:    coreerr.ClassInvariantTypeInvalid,
				Message: fmt.Sprintf("invariant %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, inv.Type),
				Field:   "Invariants",
				Got:     inv.Type,
				Want:    fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet),
			}
		}
		if inv.Type == model_logic.LogicTypeLet {
			if letTargets[inv.Target] {
				return &coreerr.ValidationError{
					Code:    coreerr.ClassInvariantDuplicateLet,
					Message: fmt.Sprintf("invariant %d: duplicate let target %q", i, inv.Target),
					Field:   "Invariants",
					Got:     inv.Target,
				}
			}
			letTargets[inv.Target] = true
		}
	}
	return nil
}

func (c *Class) validateClassChildren() error {
	actionKeys := make(map[identity.Key]bool)
	for actionKey := range c.Actions {
		actionKeys[actionKey] = true
	}
	for _, attr := range c.Attributes {
		if err := attr.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, state := range c.States {
		if err := state.ValidateWithParentAndActions(&c.Key, actionKeys); err != nil {
			return err
		}
	}
	for _, event := range c.Events {
		if err := event.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, guard := range c.Guards {
		if err := guard.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, query := range c.Queries {
		if err := query.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	return nil
}

func (c *Class) validateActionGuarantees() error {
	attrSubKeys := make(map[string]bool)
	for _, attr := range c.Attributes {
		attrSubKeys[attr.Key.SubKey] = true
	}
	for _, action := range c.Actions {
		if err := action.ValidateWithParent(&c.Key); err != nil {
			return err
		}
		for i, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet {
				continue
			}
			if guar.Target != "" && !attrSubKeys[guar.Target] {
				return &coreerr.ValidationError{
					Code:    coreerr.ClassGuaranteeInvalidTarget,
					Message: fmt.Sprintf("action %q guarantee %d: target %q is not a valid attribute on class %q", action.Key.String(), i, guar.Target, c.Key.String()),
					Field:   "Guarantees",
					Got:     guar.Target,
				}
			}
		}
	}
	return nil
}

func (c *Class) validateTransitions() error {
	stateKeys := make(map[identity.Key]bool)
	for stateKey := range c.States {
		stateKeys[stateKey] = true
	}
	eventKeys := make(map[identity.Key]bool)
	for eventKey := range c.Events {
		eventKeys[eventKey] = true
	}
	guardKeys := make(map[identity.Key]bool)
	for guardKey := range c.Guards {
		guardKeys[guardKey] = true
	}
	actionKeys := make(map[identity.Key]bool)
	for actionKey := range c.Actions {
		actionKeys[actionKey] = true
	}
	for _, transition := range c.Transitions {
		if err := transition.ValidateWithParent(&c.Key); err != nil {
			return err
		}
		if err := transition.ValidateReferences(stateKeys, eventKeys, guardKeys, actionKeys); err != nil {
			return err
		}
	}
	return nil
}
