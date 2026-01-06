package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

// atomicEnumInOut represents an allowed value in an enumeration.
type atomicEnumInOut struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// ToRequirements converts the atomicEnumInOut to data_type.AtomicEnum.
func (a atomicEnumInOut) ToRequirements() data_type.AtomicEnum {
	return data_type.AtomicEnum{
		Value:     a.Value,
		SortOrder: a.SortOrder,
	}
}

// FromRequirements creates a atomicEnumInOut from data_type.AtomicEnum.
func FromRequirementsAtomicEnum(a data_type.AtomicEnum) atomicEnumInOut {
	return atomicEnumInOut{
		Value:     a.Value,
		SortOrder: a.SortOrder,
	}
}
