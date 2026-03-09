package logic_expression_type

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// --- Scalar type validation ---

func (t *BooleanType) Validate() error  { return nil }
func (t *IntegerType) Validate() error  { return nil }
func (t *RationalType) Validate() error { return nil }
func (t *StringType) Validate() error   { return nil }

func (t *EnumType) Validate() error {
	if len(t.Values) == 0 {
		return coreerr.New(
			coreerr.ExprtypeEnumValuesRequired,
			"EnumType Values is required and must have at least one element",
			"Values",
		)
	}
	return nil
}

// --- Collection type validation ---

func (t *SetType) Validate() error {
	if t.ElementType == nil {
		return coreerr.New(
			coreerr.ExprtypeSetElementRequired,
			"SetType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(); err != nil {
		return coreerr.New(
			coreerr.ExprtypeSetElementInvalid,
			fmt.Sprintf("SetType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

func (t *SequenceType) Validate() error {
	if t.ElementType == nil {
		return coreerr.New(
			coreerr.ExprtypeSequenceElementRequired,
			"SequenceType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(); err != nil {
		return coreerr.New(
			coreerr.ExprtypeSequenceElementInvalid,
			fmt.Sprintf("SequenceType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

func (t *BagType) Validate() error {
	if t.ElementType == nil {
		return coreerr.New(
			coreerr.ExprtypeBagElementRequired,
			"BagType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(); err != nil {
		return coreerr.New(
			coreerr.ExprtypeBagElementInvalid,
			fmt.Sprintf("BagType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

// --- Compound type validation ---

func (t *TupleType) Validate() error {
	if len(t.ElementTypes) == 0 {
		return coreerr.New(
			coreerr.ExprtypeTupleElementsRequired,
			"TupleType ElementTypes is required and must have at least one element",
			"ElementTypes",
		)
	}
	for i, elem := range t.ElementTypes {
		if elem == nil {
			return coreerr.New(
				coreerr.ExprtypeTupleElementNil,
				fmt.Sprintf("TupleType.ElementTypes[%d]: is required", i),
				"ElementTypes",
			)
		}
		if err := elem.Validate(); err != nil {
			return coreerr.New(
				coreerr.ExprtypeTupleElementInvalid,
				fmt.Sprintf("TupleType.ElementTypes[%d]: %s", i, err.Error()),
				"ElementTypes",
			)
		}
	}
	return nil
}

func (t *RecordType) Validate() error {
	if len(t.Fields) == 0 {
		return coreerr.New(
			coreerr.ExprtypeRecordFieldsRequired,
			"RecordType Fields is required and must have at least one field",
			"Fields",
		)
	}
	for i, field := range t.Fields {
		if field.Name == "" {
			return coreerr.New(
				coreerr.ExprtypeRecordFieldNameRequired,
				fmt.Sprintf("RecordType.Fields[%d].Name: is required", i),
				"Fields",
			)
		}
		if field.Type == nil {
			return coreerr.New(
				coreerr.ExprtypeRecordFieldTypeRequired,
				fmt.Sprintf("RecordType.Fields[%d].Type: is required", i),
				"Fields",
			)
		}
		if err := field.Type.Validate(); err != nil {
			return coreerr.New(
				coreerr.ExprtypeRecordFieldTypeInvalid,
				fmt.Sprintf("RecordType.Fields[%d].Type: %s", i, err.Error()),
				"Fields",
			)
		}
	}
	return nil
}

func (t *FunctionType) Validate() error {
	if t.Return == nil {
		return coreerr.New(
			coreerr.ExprtypeFunctionReturnRequired,
			"FunctionType.Return: is required",
			"Return",
		)
	}
	for i, param := range t.Params {
		if param == nil {
			return coreerr.New(
				coreerr.ExprtypeFunctionParamNil,
				fmt.Sprintf("FunctionType.Params[%d]: is required", i),
				"Params",
			)
		}
		if err := param.Validate(); err != nil {
			return coreerr.New(
				coreerr.ExprtypeFunctionParamInvalid,
				fmt.Sprintf("FunctionType.Params[%d]: %s", i, err.Error()),
				"Params",
			)
		}
	}
	if err := t.Return.Validate(); err != nil {
		return coreerr.New(
			coreerr.ExprtypeFunctionReturnInvalid,
			fmt.Sprintf("FunctionType.Return: %s", err.Error()),
			"Return",
		)
	}
	return nil
}

// --- Reference type validation ---

func (t *ObjectType) Validate() error {
	if err := t.ClassKey.Validate(); err != nil {
		return coreerr.New(
			coreerr.ExprtypeObjectClasskeyInvalid,
			fmt.Sprintf("ObjectType.ClassKey: %s", err.Error()),
			"ClassKey",
		)
	}
	return nil
}
