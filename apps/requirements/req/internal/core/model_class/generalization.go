package model_class

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key             identity.Key
	Name            string
	Details         string // Markdown.
	UnfinishedNotes string // Scratch notes not yet placed in final requirement locations.
	IsComplete      bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic        bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment      string
}

// GeneralizationDetails holds the human-authored name and description from a generalization file.
type GeneralizationDetails struct {
	Name    string
	Details string
}

// GeneralizationTraits holds completeness and staticity flags for a generalization.
type GeneralizationTraits struct {
	IsComplete bool
	IsStatic   bool
}

func NewGeneralization(key identity.Key, details GeneralizationDetails, unfinishedNotes string, traits GeneralizationTraits, umlComment string) Generalization {
	return Generalization{
		Key:             key,
		Name:            details.Name,
		Details:         details.Details,
		UnfinishedNotes: unfinishedNotes,
		IsComplete:      traits.IsComplete,
		IsStatic:        traits.IsStatic,
		UmlComment:      umlComment,
	}
}

// Validate validates the Generalization struct.
func (g *Generalization) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := g.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.CgenKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if g.Key.KeyType != identity.KEY_TYPE_CLASS_GENERALIZATION {
		return coreerr.NewWithValues(ctx, coreerr.CgenKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for generalization", g.Key.KeyType), "Key", g.Key.KeyType, identity.KEY_TYPE_CLASS_GENERALIZATION)
	}

	// Name is required.
	if g.Name == "" {
		return coreerr.New(ctx, coreerr.CgenNameRequired, "Name is required", "Name")
	}

	return nil
}

// ValidateWithParent validates the Generalization, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (g *Generalization) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := g.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := g.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Generalization has no children with keys that need validation.
	return nil
}
