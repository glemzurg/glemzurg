package model_expression_type

import "fmt"

// --- Scalar type validation ---

func (t *BooleanType) Validate() error  { return nil }
func (t *IntegerType) Validate() error  { return nil }
func (t *RationalType) Validate() error { return nil }
func (t *StringType) Validate() error   { return nil }

func (t *EnumType) Validate() error {
	return _validate.Struct(t)
}

// --- Collection type validation ---

func (t *SetType) Validate() error {
	if t.ElementType == nil {
		return fmt.Errorf("SetType.ElementType: is required")
	}
	return t.ElementType.Validate()
}

func (t *SequenceType) Validate() error {
	if t.ElementType == nil {
		return fmt.Errorf("SequenceType.ElementType: is required")
	}
	return t.ElementType.Validate()
}

func (t *BagType) Validate() error {
	if t.ElementType == nil {
		return fmt.Errorf("BagType.ElementType: is required")
	}
	return t.ElementType.Validate()
}

// --- Compound type validation ---

func (t *TupleType) Validate() error {
	if err := _validate.Struct(t); err != nil {
		return err
	}
	for i, elem := range t.ElementTypes {
		if elem == nil {
			return fmt.Errorf("TupleType.ElementTypes[%d]: is required", i)
		}
		if err := elem.Validate(); err != nil {
			return fmt.Errorf("TupleType.ElementTypes[%d]: %w", i, err)
		}
	}
	return nil
}

func (t *RecordType) Validate() error {
	if err := _validate.Struct(t); err != nil {
		return err
	}
	for i, field := range t.Fields {
		if field.Name == "" {
			return fmt.Errorf("RecordType.Fields[%d].Name: is required", i)
		}
		if field.Type == nil {
			return fmt.Errorf("RecordType.Fields[%d].Type: is required", i)
		}
		if err := field.Type.Validate(); err != nil {
			return fmt.Errorf("RecordType.Fields[%d].Type: %w", i, err)
		}
	}
	return nil
}

func (t *FunctionType) Validate() error {
	if t.Return == nil {
		return fmt.Errorf("FunctionType.Return: is required")
	}
	for i, param := range t.Params {
		if param == nil {
			return fmt.Errorf("FunctionType.Params[%d]: is required", i)
		}
		if err := param.Validate(); err != nil {
			return fmt.Errorf("FunctionType.Params[%d]: %w", i, err)
		}
	}
	return t.Return.Validate()
}

// --- Reference type validation ---

func (t *ObjectType) Validate() error {
	if err := t.ClassKey.Validate(); err != nil {
		return fmt.Errorf("ObjectType.ClassKey: %w", err)
	}
	return nil
}

