package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"

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

// ToRequirements converts the attributeInOut to model_class.Attribute.
func (a attributeInOut) ToRequirements() model_class.Attribute {
	attr := model_class.Attribute{
		Key:              a.Key,
		Name:             a.Name,
		Details:          a.Details,
		DataTypeRules:    a.DataTypeRules,
		DerivationPolicy: a.DerivationPolicy,
		Nullable:         a.Nullable,
		UmlComment:       a.UmlComment,
		IndexNums:        a.IndexNums,
		DataType:         nil,
	}

	if a.DataType != nil {
		dt := a.DataType.ToRequirements()
		attr.DataType = &dt
	}

	return attr
}

// FromRequirements creates a attributeInOut from model_class.Attribute.
func FromRequirementsAttribute(a model_class.Attribute) attributeInOut {
	attr := attributeInOut{
		Key:              a.Key,
		Name:             a.Name,
		Details:          a.Details,
		DataTypeRules:    a.DataTypeRules,
		DerivationPolicy: a.DerivationPolicy,
		Nullable:         a.Nullable,
		UmlComment:       a.UmlComment,
		IndexNums:        a.IndexNums,
		DataType:         nil,
	}

	if a.DataType != nil {
		dt := FromRequirementsDataType(*a.DataType)
		attr.DataType = &dt
	}

	return attr
}
