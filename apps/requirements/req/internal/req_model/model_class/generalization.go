package model_class

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        identity.Key
	Name       string `validate:"required"`
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
		return err
	}
	if g.Key.KeyType() != identity.KEY_TYPE_GENERALIZATION {
		return errors.Errorf("Key: invalid key type '%s' for generalization.", g.Key.KeyType())
	}

	// Validate struct tags (Name required).
	if err := _validate.Struct(g); err != nil {
		return err
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
