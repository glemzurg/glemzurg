package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
	// Children
	SuperclassKey identity.Key   // If this generalization is classes, the superclass for it.
	SubclassKeys  []identity.Key // If this generalization is classes, the subclasses for it.
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
	return validation.ValidateStruct(g,
		validation.Field(&g.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_GENERALIZATION {
				return errors.Errorf("invalid key type '%s' for generalization", k.KeyType())
			}
			return nil
		})),
		validation.Field(&g.Name, validation.Required),
	)
}

func (g *Generalization) SetSuperSubclassKeys(superclassKey identity.Key, subclassKeys []identity.Key) {
	g.SuperclassKey = superclassKey
	g.SubclassKeys = subclassKeys
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
