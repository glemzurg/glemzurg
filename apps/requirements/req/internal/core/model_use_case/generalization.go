package model_use_case

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
}

func NewGeneralization(key identity.Key, name, details string, isComplete, isStatic bool, umlComment string) (generalization Generalization, err error) {
	generalization = Generalization{
		Key:        key,
		Name:       name,
		Details:    details,
		IsComplete: isComplete,
		IsStatic:   isStatic,
		UmlComment: umlComment,
	}

	if err = generalization.Validate(); err != nil {
		return Generalization{}, err
	}

	return generalization, nil
}

// Validate validates the Generalization struct.
func (g *Generalization) Validate() error {
	// Validate the key.
	if err := g.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.UcgenKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if g.Key.KeyType != identity.KEY_TYPE_USE_CASE_GENERALIZATION {
		return &coreerr.ValidationError{
			Code:    coreerr.UcgenKeyTypeInvalid,
			Message: fmt.Sprintf("key: invalid key type '%s' for use case generalization", g.Key.KeyType),
			Field:   "Key",
			Got:     g.Key.KeyType,
			Want:    identity.KEY_TYPE_USE_CASE_GENERALIZATION,
		}
	}

	// Validate Name required.
	if g.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.UcgenNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}

	return nil
}

// ValidateWithParent validates the Generalization, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (g *Generalization) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := g.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := g.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Generalization has no children with keys that need validation.
	return nil
}
