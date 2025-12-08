package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Attribute is a member of a class.
type Attribute struct {
	Key              string
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy string // If this is a derived attribute, how is it derived.
	Nullable         bool   // Is this attribute optional.
	UmlComment       string
	// Part of the data in a parsed file.
	IndexNums []uint // The indexes this attribute is part of.
}

func NewAttribute(key, name, details, dataTypeRules, derivationPolicy string, nullable bool, umlComment string, indexNums []uint) (attribute Attribute, err error) {

	attribute = Attribute{
		Key:              key,
		Name:             name,
		Details:          details,
		DataTypeRules:    dataTypeRules,
		DerivationPolicy: derivationPolicy,
		Nullable:         nullable,
		UmlComment:       umlComment,
		IndexNums:        indexNums,
	}

	err = validation.ValidateStruct(&attribute,
		validation.Field(&attribute.Key, validation.Required),
		validation.Field(&attribute.Name, validation.Required),
	)
	if err != nil {
		return Attribute{}, errors.WithStack(err)
	}

	return attribute, nil
}

func createKeyAttributeLookup(byCategory map[string][]Attribute) (lookup map[string]Attribute) {
	lookup = map[string]Attribute{}
	for _, items := range byCategory {
		for _, item := range items {
			lookup[item.Key] = item
		}
	}
	return lookup
}
