package model_state

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Event is what triggers a transition between states.
type Event struct {
	Key     identity.Key
	Name    string
	Details string
	// ParameterNames lists payload field names carried by this event, in order.
	// Names may be a superset of any action or query parameters bound on a transition.
	ParameterNames []string
}

func NewEvent(key identity.Key, name, details string, parameterNames []string) Event {
	return Event{
		Key:            key,
		Name:           name,
		Details:        details,
		ParameterNames: parameterNames,
	}
}

// Validate validates the Event struct.
func (e *Event) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := e.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.EventKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if e.Key.KeyType != identity.KEY_TYPE_EVENT {
		return coreerr.NewWithValues(ctx, coreerr.EventKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for event", e.Key.KeyType), "Key", e.Key.KeyType, identity.KEY_TYPE_EVENT)
	}

	if e.Name == "" {
		return coreerr.New(ctx, coreerr.EventNameRequired, "Name is required", "Name")
	}
	if badChar := coreerr.ValidateNameChars(e.Name); badChar != "" {
		return coreerr.NewWithValues(ctx, coreerr.EventNameInvalidChars, fmt.Sprintf("Name contains invalid character %q", badChar), "Name", e.Name, "A-Za-z0-9 space hyphen underscore")
	}

	return validateEventParameterNames(ctx, e.ParameterNames)
}

// ValidateWithParent validates the Event, its key's parent relationship, and parameter names.
// The parent must be a Class.
func (e *Event) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := e.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := e.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	return nil
}

func validateEventParameterNames(ctx *coreerr.ValidationContext, names []string) error {
	if len(names) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(names))
	for i, name := range names {
		childCtx := ctx.Child("parameter_name", fmt.Sprintf("%d", i))
		if strings.TrimSpace(name) == "" {
			return coreerr.New(childCtx, coreerr.EventParameterNameRequired, "Parameter name is required", "ParameterNames")
		}
		normalized := identity.NormalizeSubKey(name)
		if seen[normalized] {
			return coreerr.NewWithValues(
				childCtx,
				coreerr.EventParameterNameDuplicate,
				fmt.Sprintf("duplicate parameter name %q", name),
				"ParameterNames",
				name,
				"",
			)
		}
		seen[normalized] = true
	}
	return nil
}
