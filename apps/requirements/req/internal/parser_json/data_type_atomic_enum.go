package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

// atomicEnumInOut represents an allowed value in an enumeration.
type atomicEnumInOut struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// ToRequirements converts the atomicEnumInOut to model_data_type.AtomicEnum.
func (a atomicEnumInOut) ToRequirements() model_data_type.AtomicEnum {
	return model_data_type.AtomicEnum{
		Value:     a.Value,
		SortOrder: a.SortOrder,
	}
}

// FromRequirements creates a atomicEnumInOut from model_data_type.AtomicEnum.
func FromRequirementsAtomicEnum(a model_data_type.AtomicEnum) atomicEnumInOut {
	return atomicEnumInOut{
		Value:     a.Value,
		SortOrder: a.SortOrder,
	}
}
