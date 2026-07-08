package model_state

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// System event names reserved for implicit initial and final pseudo-states.
// Stored and authored as _new / _destroy; rendered as «new» / «destroy» in diagrams and docs.
const (
	EventNameNew     = "_new"
	EventNameDestroy = "_destroy"

	EventTLANameNew     = "«new»"
	EventTLANameDestroy = "«destroy»"
)

// IsSystemCreationEvent reports whether name is the reserved creation event _new.
func IsSystemCreationEvent(name string) bool {
	return name == EventNameNew
}

// IsSystemFinalEvent reports whether name is the reserved finalization event _destroy.
func IsSystemFinalEvent(name string) bool {
	return name == EventNameDestroy
}

// SystemEventDisplayName returns the UML stereotype label for system events.
func SystemEventDisplayName(name string) string {
	switch name {
	case EventNameNew:
		return EventTLANameNew
	case EventNameDestroy:
		return EventTLANameDestroy
	default:
		return name
	}
}

// SystemEventTLAName returns the canonical TLA+ spelling for a system event.
// Accepts both ASCII authoring names (_new, _destroy) and guillemet forms («new», «destroy»).
func SystemEventTLAName(name string) string {
	switch name {
	case EventNameNew, EventTLANameNew:
		return EventTLANameNew
	case EventNameDestroy, EventTLANameDestroy:
		return EventTLANameDestroy
	default:
		return name
	}
}

// IsSystemEventTLAName reports whether name is a system event in ASCII or TLA form.
func IsSystemEventTLAName(name string) bool {
	return name == EventNameNew || name == EventNameDestroy ||
		name == EventTLANameNew || name == EventTLANameDestroy
}

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

	if err := validateEventName(ctx, e.Name); err != nil {
		return err
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

func isValidEventName(name string) bool {
	if name == EventNameNew || name == EventNameDestroy {
		return true
	}
	return coreerr.ValidateIdentifierName(name)
}

func validateEventName(ctx *coreerr.ValidationContext, name string) error {
	if name == "" {
		return coreerr.New(ctx, coreerr.EventNameRequired, "Name is required", "Name")
	}
	if !isValidEventName(name) {
		return coreerr.NewWithValues(
			ctx,
			coreerr.EventNameInvalidChars,
			fmt.Sprintf("Name %q must be _new, _destroy, or match ^[a-zA-Z][a-zA-Z0-9_]*$", name),
			"Name",
			name,
			"_new, _destroy, or ^[a-zA-Z][a-zA-Z0-9_]*$",
		)
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
		if !coreerr.ValidateIdentifierName(name) {
			return coreerr.NewWithValues(
				childCtx,
				coreerr.EventParameterNameInvalidChars,
				fmt.Sprintf("Parameter name %q must match ^[a-zA-Z][a-zA-Z0-9_]*$", name),
				"ParameterNames",
				name,
				"^[a-zA-Z][a-zA-Z0-9_]*$",
			)
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
