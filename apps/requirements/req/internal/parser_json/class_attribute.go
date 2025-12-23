package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// attributeInOut is a member of a class.
type attributeInOut struct {
	Key              string `json:"key"`
	Name             string `json:"name"`
	Details          string `json:"details"` // Markdown.
	DataTypeRules    string `json:"data_type_rules"`
	DerivationPolicy string `json:"derivation_policy"`
	Nullable         bool   `json:"nullable"`
	UmlComment       string `json:"uml_comment"`
	// Part of the data in a parsed file.
	IndexNums []uint         `json:"index_nums"` // The indexes this attribute is part of.
	DataType  *dataTypeInOut `json:"data_type"`  // If the DataTypeRules can be parsed, this is the resulting data type.
}

// ToRequirements converts the attributeInOut to requirements.Attribute.
func (a attributeInOut) ToRequirements() requirements.Attribute {
	attr := requirements.Attribute{
		Key:              a.Key,
		Name:             a.Name,
		Details:          a.Details,
		DataTypeRules:    a.DataTypeRules,
		DerivationPolicy: a.DerivationPolicy,
		Nullable:         a.Nullable,
		UmlComment:       a.UmlComment,
		IndexNums:        a.IndexNums,
		DataType:         nil, // TODO: convert
	}
	return attr
}

// FromRequirements creates a attributeInOut from requirements.Attribute.
func FromRequirementsAttribute(a requirements.Attribute) attributeInOut {
	attr := attributeInOut{
		Key:              a.Key,
		Name:             a.Name,
		Details:          a.Details,
		DataTypeRules:    a.DataTypeRules,
		DerivationPolicy: a.DerivationPolicy,
		Nullable:         a.Nullable,
		UmlComment:       a.UmlComment,
		IndexNums:        a.IndexNums,
		DataType:         nil, // TODO: convert
	}
	return attr
}
