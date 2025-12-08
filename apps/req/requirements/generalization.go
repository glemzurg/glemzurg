package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        string
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
	// Part of the data in a parsed file.
	SuperclassKey string   // If this generalization is classes, the superclass for it.
	SubclassKeys  []string // If this generalization is classes, the subclasses for it.
}

func NewGeneralization(key, name, details string, isComplete, isStatic bool, umlComment string) (generalization Generalization, err error) {

	generalization = Generalization{
		Key:        key,
		Name:       name,
		Details:    details,
		IsComplete: isComplete,
		IsStatic:   isStatic,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&generalization,
		validation.Field(&generalization.Key, validation.Required),
		validation.Field(&generalization.Name, validation.Required),
	)
	if err != nil {
		return Generalization{}, errors.WithStack(err)
	}

	return generalization, nil
}

func (g *Generalization) SetSuperSubclassKeys(superclassKey string, subclassKeys []string) {
	g.SuperclassKey = superclassKey
	g.SubclassKeys = subclassKeys
}

func createKeyGeneralizationLookup(domainClasses map[string][]Class, items []Generalization) (lookup map[string]Generalization) {

	// Classes that are part of generalizations.
	superclassKeyOf := map[string]string{}
	subclassKeysOf := map[string][]string{}
	for _, classes := range domainClasses {
		for _, class := range classes {
			if class.SuperclassOfKey != "" {
				superclassKeyOf[class.SuperclassOfKey] = class.Key
			}
			if class.SubclassOfKey != "" {
				subclassKeys := subclassKeysOf[class.SubclassOfKey]
				subclassKeys = append(subclassKeys, class.Key)
				subclassKeysOf[class.SubclassOfKey] = subclassKeys
			}
		}
	}

	lookup = map[string]Generalization{}
	for _, item := range items {

		item.SetSuperSubclassKeys(superclassKeyOf[item.Key], subclassKeysOf[item.Key])

		lookup[item.Key] = item
	}
	return lookup
}
