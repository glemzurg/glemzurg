package logic_expression_type

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// --- Scalar type validation ---

func (t *BooleanType) Validate(_ *coreerr.ValidationContext) error  { return nil }
func (t *IntegerType) Validate(_ *coreerr.ValidationContext) error  { return nil }
func (t *RationalType) Validate(_ *coreerr.ValidationContext) error { return nil }
func (t *StringType) Validate(_ *coreerr.ValidationContext) error   { return nil }

func (t *EnumType) Validate(ctx *coreerr.ValidationContext) error {
	if len(t.Values) == 0 {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeEnumValuesRequired,
			"EnumType Values is required and must have at least one element",
			"Values",
		)
	}
	return nil
}

// --- Collection type validation ---

func (t *SetType) Validate(ctx *coreerr.ValidationContext) error {
	if t.ElementType == nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeSetElementRequired,
			"SetType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(ctx); err != nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeSetElementInvalid,
			fmt.Sprintf("SetType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

func (t *SequenceType) Validate(ctx *coreerr.ValidationContext) error {
	if t.ElementType == nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeSequenceElementRequired,
			"SequenceType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(ctx); err != nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeSequenceElementInvalid,
			fmt.Sprintf("SequenceType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

func (t *BagType) Validate(ctx *coreerr.ValidationContext) error {
	if t.ElementType == nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeBagElementRequired,
			"BagType.ElementType: is required",
			"ElementType",
		)
	}
	if err := t.ElementType.Validate(ctx); err != nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeBagElementInvalid,
			fmt.Sprintf("BagType.ElementType: %s", err.Error()),
			"ElementType",
		)
	}
	return nil
}

// --- Compound type validation ---

func (t *TupleType) Validate(ctx *coreerr.ValidationContext) error {
	if len(t.ElementTypes) == 0 {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeTupleElementsRequired,
			"TupleType ElementTypes is required and must have at least one element",
			"ElementTypes",
		)
	}
	for i, elem := range t.ElementTypes {
		if elem == nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeTupleElementNil,
				fmt.Sprintf("TupleType.ElementTypes[%d]: is required", i),
				"ElementTypes",
			)
		}
		if err := elem.Validate(ctx); err != nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeTupleElementInvalid,
				fmt.Sprintf("TupleType.ElementTypes[%d]: %s", i, err.Error()),
				"ElementTypes",
			)
		}
	}
	return nil
}

func (t *RecordType) Validate(ctx *coreerr.ValidationContext) error {
	if len(t.Fields) == 0 {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeRecordFieldsRequired,
			"RecordType Fields is required and must have at least one field",
			"Fields",
		)
	}
	for i, field := range t.Fields {
		if field.Name == "" {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeRecordFieldNameRequired,
				fmt.Sprintf("RecordType.Fields[%d].Name: is required", i),
				"Fields",
			)
		}
		if field.Type == nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeRecordFieldTypeRequired,
				fmt.Sprintf("RecordType.Fields[%d].Type: is required", i),
				"Fields",
			)
		}
		if err := field.Type.Validate(ctx); err != nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeRecordFieldTypeInvalid,
				fmt.Sprintf("RecordType.Fields[%d].Type: %s", i, err.Error()),
				"Fields",
			)
		}
	}
	return nil
}

func (t *FunctionType) Validate(ctx *coreerr.ValidationContext) error {
	if t.Return == nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeFunctionReturnRequired,
			"FunctionType.Return: is required",
			"Return",
		)
	}
	for i, param := range t.Params {
		if param == nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeFunctionParamNil,
				fmt.Sprintf("FunctionType.Params[%d]: is required", i),
				"Params",
			)
		}
		if err := param.Validate(ctx); err != nil {
			return coreerr.New(
				ctx,
				coreerr.ExprtypeFunctionParamInvalid,
				fmt.Sprintf("FunctionType.Params[%d]: %s", i, err.Error()),
				"Params",
			)
		}
	}
	if err := t.Return.Validate(ctx); err != nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeFunctionReturnInvalid,
			fmt.Sprintf("FunctionType.Return: %s", err.Error()),
			"Return",
		)
	}
	return nil
}

// --- Reference type validation ---

func (t *ObjectType) Validate(ctx *coreerr.ValidationContext) error {
	if err := t.ClassKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(
			ctx,
			coreerr.ExprtypeObjectClasskeyInvalid,
			fmt.Sprintf("ObjectType.ClassKey: %s", err.Error()),
			"ClassKey",
		)
	}
	return nil
}
