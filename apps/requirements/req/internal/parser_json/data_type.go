package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

// dataTypeInOut represents the main data type structure.
type dataTypeInOut struct {
	Key              string       `json:"key"`
	CollectionType   string       `json:"collection_type"`
	CollectionUnique *bool        `json:"collection_unique"`
	CollectionMin    *int         `json:"collection_min"`
	CollectionMax    *int         `json:"collection_max"`
	Atomic           *atomicInOut `json:"atomic"`
	RecordFields     []fieldInOut `json:"record_fields"`
}

// ToRequirements converts the dataTypeInOut to data_type.DataType.
func (d dataTypeInOut) ToRequirements() data_type.DataType {
	dt := data_type.DataType{
		Key:              d.Key,
		CollectionType:   d.CollectionType,
		CollectionUnique: d.CollectionUnique,
		CollectionMin:    d.CollectionMin,
		CollectionMax:    d.CollectionMax,
		Atomic:           nil,
		RecordFields:     nil,
	}
	if d.Atomic != nil {
		a := d.Atomic.ToRequirements()
		dt.Atomic = &a
	}
	for _, f := range d.RecordFields {
		dt.RecordFields = append(dt.RecordFields, f.ToRequirements())
	}
	return dt
}

// FromRequirements creates a dataTypeInOut from data_type.DataType.
func FromRequirementsDataType(d data_type.DataType) dataTypeInOut {
	dt := dataTypeInOut{
		Key:              d.Key,
		CollectionType:   d.CollectionType,
		CollectionUnique: d.CollectionUnique,
		CollectionMin:    d.CollectionMin,
		CollectionMax:    d.CollectionMax,
		Atomic:           nil,
		RecordFields:     nil,
	}
	if d.Atomic != nil {
		a := FromRequirementsAtomic(*d.Atomic)
		dt.Atomic = &a
	}
	for _, f := range d.RecordFields {
		dt.RecordFields = append(dt.RecordFields, FromRequirementsField(f))
	}
	return dt
}
