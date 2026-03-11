package model_class

import (
	"fmt"

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

func NewClass(key identity.Key, name, details string, actorKey, superclassOfKey, subclassOfKey *identity.Key, umlComment string) Class {
	return Class{
		Key:             key,
		Name:            name,
		Details:         details,
		ActorKey:        actorKey,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      umlComment,
	}
}

// Validate validates the Class struct.
func (c *Class) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := c.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.ClassKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if c.Key.KeyType != identity.KEY_TYPE_CLASS {
		return coreerr.NewWithValues(ctx, coreerr.ClassKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for class", c.Key.KeyType), "Key", c.Key.KeyType, identity.KEY_TYPE_CLASS)
	}

	// Name is required.
	if c.Name == "" {
		return coreerr.New(ctx, coreerr.ClassNameRequired, "Name is required", "Name")
	}

	// Validate FK key types.
	if c.ActorKey != nil {
		if err := c.ActorKey.ValidateWithContext(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ClassActorkeyInvalid, fmt.Sprintf("ActorKey: %s", err.Error()), "ActorKey")
		}
		if c.ActorKey.KeyType != identity.KEY_TYPE_ACTOR {
			return coreerr.NewWithValues(ctx, coreerr.ClassActorkeyTypeInvalid, fmt.Sprintf("ActorKey: invalid key type '%s' for actor", c.ActorKey.KeyType), "ActorKey", c.ActorKey.KeyType, identity.KEY_TYPE_ACTOR)
		}
	}
	if c.SuperclassOfKey != nil {
		if err := c.SuperclassOfKey.ValidateWithContext(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ClassSuperkeyInvalid, fmt.Sprintf("SuperclassOfKey: %s", err.Error()), "SuperclassOfKey")
		}
		if c.SuperclassOfKey.KeyType != identity.KEY_TYPE_CLASS_GENERALIZATION {
			return coreerr.NewWithValues(ctx, coreerr.ClassSuperkeyTypeInvalid, fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for class generalization", c.SuperclassOfKey.KeyType), "SuperclassOfKey", c.SuperclassOfKey.KeyType, identity.KEY_TYPE_CLASS_GENERALIZATION)
		}
	}
	if c.SubclassOfKey != nil {
		if err := c.SubclassOfKey.ValidateWithContext(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ClassSubkeyInvalid, fmt.Sprintf("SubclassOfKey: %s", err.Error()), "SubclassOfKey")
		}
		if c.SubclassOfKey.KeyType != identity.KEY_TYPE_CLASS_GENERALIZATION {
			return coreerr.NewWithValues(ctx, coreerr.ClassSubkeyTypeInvalid, fmt.Sprintf("SubclassOfKey: invalid key type '%s' for class generalization", c.SubclassOfKey.KeyType), "SubclassOfKey", c.SubclassOfKey.KeyType, identity.KEY_TYPE_CLASS_GENERALIZATION)
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if c.SuperclassOfKey != nil && c.SubclassOfKey != nil && *c.SuperclassOfKey == *c.SubclassOfKey {
		return coreerr.New(ctx, coreerr.ClassSuperSubSame, "SuperclassOfKey and SubclassOfKey cannot be the same", "SuperclassOfKey")
	}
	return nil
}

// ValidateReferences validates that the class's reference keys point to valid entities.
// - ActorKey must exist in the actors map
// - SuperclassOfKey must exist in the generalizations map and be in the same subdomain
// - SubclassOfKey must exist in the generalizations map and be in the same subdomain.
func (c *Class) ValidateReferences(ctx *coreerr.ValidationContext, actors map[identity.Key]bool, generalizations map[identity.Key]bool) error {
	// Validate ActorKey references a real actor.
	if c.ActorKey != nil {
		if !actors[*c.ActorKey] {
			return coreerr.NewWithValues(ctx, coreerr.ClassActorNotfound, fmt.Sprintf("class '%s' references non-existent actor '%s'", c.Key.String(), c.ActorKey.String()), "ActorKey", c.ActorKey.String(), "")
		}
	}

	// Get this class's subdomain from its parent key.
	classSubdomainKey := c.Key.ParentKey

	// Validate SuperclassOfKey references a real generalization in the same subdomain.
	if c.SuperclassOfKey != nil {
		if !generalizations[*c.SuperclassOfKey] {
			return coreerr.NewWithValues(ctx, coreerr.ClassSupergenNotfound, fmt.Sprintf("class '%s' references non-existent generalization '%s'", c.Key.String(), c.SuperclassOfKey.String()), "SuperclassOfKey", c.SuperclassOfKey.String(), "")
		}
		// Check same subdomain.
		generalizationSubdomainKey := c.SuperclassOfKey.ParentKey
		if classSubdomainKey != generalizationSubdomainKey {
			return coreerr.NewWithValues(ctx, coreerr.ClassSupergenWrongSubdomain, fmt.Sprintf("class '%s' generalization '%s' must be in the same subdomain", c.Key.String(), c.SuperclassOfKey.String()), "SuperclassOfKey", c.SuperclassOfKey.String(), "")
		}
	}

	// Validate SubclassOfKey references a real generalization in the same subdomain.
	if c.SubclassOfKey != nil {
		if !generalizations[*c.SubclassOfKey] {
			return coreerr.NewWithValues(ctx, coreerr.ClassSubgenNotfound, fmt.Sprintf("class '%s' references non-existent generalization '%s'", c.Key.String(), c.SubclassOfKey.String()), "SubclassOfKey", c.SubclassOfKey.String(), "")
		}
		// Check same subdomain.
		generalizationSubdomainKey := c.SubclassOfKey.ParentKey
		if classSubdomainKey != generalizationSubdomainKey {
			return coreerr.NewWithValues(ctx, coreerr.ClassSubgenWrongSubdomain, fmt.Sprintf("class '%s' generalization '%s' must be in the same subdomain", c.Key.String(), c.SubclassOfKey.String()), "SubclassOfKey", c.SubclassOfKey.String(), "")
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
func (c *Class) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	if err := c.Validate(ctx); err != nil {
		return err
	}
	if err := c.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	if err := c.validateClassInvariants(ctx); err != nil {
		return err
	}
	if err := c.validateClassChildren(ctx); err != nil {
		return err
	}
	if err := c.validateActionGuarantees(ctx); err != nil {
		return err
	}
	if err := c.validateTransitions(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Class) validateClassInvariants(ctx *coreerr.ValidationContext) error {
	letTargets := make(map[string]bool)
	for i, inv := range c.Invariants {
		invCtx := ctx.Child("invariant", fmt.Sprintf("%d", i))
		if err := inv.ValidateWithParent(invCtx, &c.Key); err != nil {
			return coreerr.New(invCtx, coreerr.ClassInvariantTypeInvalid, fmt.Sprintf("invariant %d: %s", i, err.Error()), "Invariants")
		}
		if inv.Type != model_logic.LogicTypeAssessment && inv.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(invCtx, coreerr.ClassInvariantTypeInvalid, fmt.Sprintf("invariant %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, inv.Type), "Invariants", inv.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if inv.Type == model_logic.LogicTypeLet {
			if letTargets[inv.Target] {
				return coreerr.NewWithValues(invCtx, coreerr.ClassInvariantDuplicateLet, fmt.Sprintf("invariant %d: duplicate let target %q", i, inv.Target), "Invariants", inv.Target, "")
			}
			letTargets[inv.Target] = true
		}
	}
	return nil
}

func (c *Class) validateClassChildren(ctx *coreerr.ValidationContext) error {
	actionKeys := make(map[identity.Key]bool)
	for actionKey := range c.Actions {
		actionKeys[actionKey] = true
	}
	for _, attr := range c.Attributes {
		attrCtx := ctx.Child("attribute", attr.Key.String())
		if err := attr.ValidateWithParent(attrCtx, &c.Key); err != nil {
			return err
		}
	}
	for _, state := range c.States {
		stateCtx := ctx.Child("state", state.Key.String())
		if err := state.ValidateWithParentAndActions(stateCtx, &c.Key, actionKeys); err != nil {
			return err
		}
	}
	for _, event := range c.Events {
		eventCtx := ctx.Child("event", event.Key.String())
		if err := event.ValidateWithParent(eventCtx, &c.Key); err != nil {
			return err
		}
	}
	for _, guard := range c.Guards {
		guardCtx := ctx.Child("guard", guard.Key.String())
		if err := guard.ValidateWithParent(guardCtx, &c.Key); err != nil {
			return err
		}
	}
	for _, query := range c.Queries {
		queryCtx := ctx.Child("query", query.Key.String())
		if err := query.ValidateWithParent(queryCtx, &c.Key); err != nil {
			return err
		}
	}
	return nil
}

func (c *Class) validateActionGuarantees(ctx *coreerr.ValidationContext) error {
	attrSubKeys := make(map[string]bool)
	for _, attr := range c.Attributes {
		attrSubKeys[attr.Key.SubKey] = true
	}
	for _, action := range c.Actions {
		actionCtx := ctx.Child("action", action.Key.String())
		if err := action.ValidateWithParent(actionCtx, &c.Key); err != nil {
			return err
		}
		for i, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet {
				continue
			}
			if guar.Target != "" && !attrSubKeys[guar.Target] {
				guarCtx := actionCtx.Child("guarantee", fmt.Sprintf("%d", i))
				return coreerr.NewWithValues(guarCtx, coreerr.ClassGuaranteeInvalidTarget, fmt.Sprintf("action %q guarantee %d: target %q is not a valid attribute on class %q", action.Key.String(), i, guar.Target, c.Key.String()), "Guarantees", guar.Target, "")
			}
		}
	}
	return nil
}

func (c *Class) validateTransitions(ctx *coreerr.ValidationContext) error {
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
		transCtx := ctx.Child("transition", transition.Key.String())
		if err := transition.ValidateWithParent(transCtx, &c.Key); err != nil {
			return err
		}
		if err := transition.ValidateReferences(transCtx, stateKeys, eventKeys, guardKeys, actionKeys); err != nil {
			return err
		}
	}
	return nil
}
