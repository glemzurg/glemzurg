package model_expression_type

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// --- Supporting types ---

// RecordFieldType is a name-type pair within a RecordType.
type RecordFieldType struct {
	Name string         `validate:"required"`
	Type ExpressionType // Required — validated manually since interface fields can't use struct tags.
}

// --- Scalar types ---

// BooleanType represents the boolean type (TRUE/FALSE).
type BooleanType struct{}

func (t *BooleanType) expressionType()   {}
func (t *BooleanType) TypeName() string  { return TypeBoolean }

// IntegerType represents the integer type (Nat, Int).
type IntegerType struct{}

func (t *IntegerType) expressionType()   {}
func (t *IntegerType) TypeName() string  { return TypeInteger }

// RationalType represents the rational number type (Real).
type RationalType struct{}

func (t *RationalType) expressionType()   {}
func (t *RationalType) TypeName() string  { return TypeRational }

// StringType represents the string type (STRING).
type StringType struct{}

func (t *StringType) expressionType()   {}
func (t *StringType) TypeName() string  { return TypeString }

// EnumType represents a finite enumeration of string values.
type EnumType struct {
	Values []string `validate:"required,min=1"`
}

func (t *EnumType) expressionType()   {}
func (t *EnumType) TypeName() string  { return TypeEnum }

// --- Collection types ---

// SetType represents a set of elements of a given type.
type SetType struct {
	ElementType ExpressionType // Required.
}

func (t *SetType) expressionType()   {}
func (t *SetType) TypeName() string  { return TypeSet }

// SequenceType represents an ordered sequence. Unique=true means no duplicates.
type SequenceType struct {
	ElementType ExpressionType // Required.
	Unique      bool
}

func (t *SequenceType) expressionType()   {}
func (t *SequenceType) TypeName() string  { return TypeSequence }

// BagType represents a multiset (bag) of elements of a given type.
type BagType struct {
	ElementType ExpressionType // Required.
}

func (t *BagType) expressionType()   {}
func (t *BagType) TypeName() string  { return TypeBag }

// --- Compound types ---

// TupleType represents a fixed-length tuple of typed elements.
type TupleType struct {
	ElementTypes []ExpressionType `validate:"required,min=1"`
}

func (t *TupleType) expressionType()   {}
func (t *TupleType) TypeName() string  { return TypeTuple }

// RecordType represents a record with named, typed fields.
type RecordType struct {
	Fields []RecordFieldType `validate:"required,min=1"`
}

func (t *RecordType) expressionType()   {}
func (t *RecordType) TypeName() string  { return TypeRecord }

// FunctionType represents a function from parameter types to a return type.
type FunctionType struct {
	Params []ExpressionType // May be empty (zero-arg function).
	Return ExpressionType   // Required.
}

func (t *FunctionType) expressionType()   {}
func (t *FunctionType) TypeName() string  { return TypeFunction }

// --- Reference types ---

// ObjectType represents a reference to a class instance by class key.
type ObjectType struct {
	ClassKey identity.Key // Required.
}

func (t *ObjectType) expressionType()   {}
func (t *ObjectType) TypeName() string  { return TypeObject }
