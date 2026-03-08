package model_expression_type

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
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeEnumValuesRequired,
			Message: "EnumType Values is required and must have at least one element",
			Field:   "Values",
		}
	}
	return nil
}

// --- Collection type validation ---

func (t *SetType) Validate() error {
	if t.ElementType == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeSetElementRequired,
			Message: "SetType.ElementType: is required",
			Field:   "ElementType",
		}
	}
	if err := t.ElementType.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeSetElementInvalid,
			Message: fmt.Sprintf("SetType.ElementType: %s", err.Error()),
			Field:   "ElementType",
		}
	}
	return nil
}

func (t *SequenceType) Validate() error {
	if t.ElementType == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeSequenceElementRequired,
			Message: "SequenceType.ElementType: is required",
			Field:   "ElementType",
		}
	}
	if err := t.ElementType.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeSequenceElementInvalid,
			Message: fmt.Sprintf("SequenceType.ElementType: %s", err.Error()),
			Field:   "ElementType",
		}
	}
	return nil
}

func (t *BagType) Validate() error {
	if t.ElementType == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeBagElementRequired,
			Message: "BagType.ElementType: is required",
			Field:   "ElementType",
		}
	}
	if err := t.ElementType.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeBagElementInvalid,
			Message: fmt.Sprintf("BagType.ElementType: %s", err.Error()),
			Field:   "ElementType",
		}
	}
	return nil
}

// --- Compound type validation ---

func (t *TupleType) Validate() error {
	if len(t.ElementTypes) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeTupleElementsRequired,
			Message: "TupleType ElementTypes is required and must have at least one element",
			Field:   "ElementTypes",
		}
	}
	for i, elem := range t.ElementTypes {
		if elem == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeTupleElementNil,
				Message: fmt.Sprintf("TupleType.ElementTypes[%d]: is required", i),
				Field:   "ElementTypes",
			}
		}
		if err := elem.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeTupleElementInvalid,
				Message: fmt.Sprintf("TupleType.ElementTypes[%d]: %s", i, err.Error()),
				Field:   "ElementTypes",
			}
		}
	}
	return nil
}

func (t *RecordType) Validate() error {
	if len(t.Fields) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeRecordFieldsRequired,
			Message: "RecordType Fields is required and must have at least one field",
			Field:   "Fields",
		}
	}
	for i, field := range t.Fields {
		if field.Name == "" {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeRecordFieldNameRequired,
				Message: fmt.Sprintf("RecordType.Fields[%d].Name: is required", i),
				Field:   "Fields",
			}
		}
		if field.Type == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeRecordFieldTypeRequired,
				Message: fmt.Sprintf("RecordType.Fields[%d].Type: is required", i),
				Field:   "Fields",
			}
		}
		if err := field.Type.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeRecordFieldTypeInvalid,
				Message: fmt.Sprintf("RecordType.Fields[%d].Type: %s", i, err.Error()),
				Field:   "Fields",
			}
		}
	}
	return nil
}

func (t *FunctionType) Validate() error {
	if t.Return == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeFunctionReturnRequired,
			Message: "FunctionType.Return: is required",
			Field:   "Return",
		}
	}
	for i, param := range t.Params {
		if param == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeFunctionParamNil,
				Message: fmt.Sprintf("FunctionType.Params[%d]: is required", i),
				Field:   "Params",
			}
		}
		if err := param.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprtypeFunctionParamInvalid,
				Message: fmt.Sprintf("FunctionType.Params[%d]: %s", i, err.Error()),
				Field:   "Params",
			}
		}
	}
	if err := t.Return.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeFunctionReturnInvalid,
			Message: fmt.Sprintf("FunctionType.Return: %s", err.Error()),
			Field:   "Return",
		}
	}
	return nil
}

// --- Reference type validation ---

func (t *ObjectType) Validate() error {
	if err := t.ClassKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprtypeObjectClasskeyInvalid,
			Message: fmt.Sprintf("ObjectType.ClassKey: %s", err.Error()),
			Field:   "ClassKey",
		}
	}
	return nil
}
