package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

func validateGeneralizationKey(value interface{}) error {
	key, ok := value.(identity.Key)
	if !ok {
		return errors.New("invalid key type")
	}
	if key.KeyType() != identity.KEY_TYPE_GENERALIZATION {
		return errors.Errorf("key must be of type '%s', not '%s'", identity.KEY_TYPE_GENERALIZATION, key.KeyType())
	}
	return nil
}

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
	// Part of the data in a parsed file.
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

	err = validation.ValidateStruct(&generalization,
		validation.Field(&generalization.Key, validation.By(validateGeneralizationKey)),
		validation.Field(&generalization.Name, validation.Required),
	)
	if err != nil {
		return Generalization{}, errors.WithStack(err)
	}

	return generalization, nil
}

func (g *Generalization) SetSuperSubclassKeys(superclassKey identity.Key, subclassKeys []identity.Key) {
	g.SuperclassKey = superclassKey
	g.SubclassKeys = subclassKeys
}
