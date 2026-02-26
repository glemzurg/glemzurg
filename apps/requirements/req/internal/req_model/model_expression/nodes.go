package model_expression

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// --- Supporting types (not Expression nodes themselves) ---

// RecordField is a name-value pair within a RecordLiteral.
type RecordField struct {
	Name  string     `validate:"required"`
	Value Expression
}

// FieldAlteration is a field update within a RecordUpdate.
type FieldAlteration struct {
	Field string     `validate:"required"`
	Value Expression
}

// CaseBranch is a condition-result pair within a Case expression.
type CaseBranch struct {
	Condition Expression
	Result    Expression
}

// --- Literals ---

// BoolLiteral represents a boolean constant (TRUE or FALSE).
type BoolLiteral struct {
	Value bool
}

func (n *BoolLiteral) expressionNode()    {}
func (n *BoolLiteral) NodeType() string   { return NodeBoolLiteral }

// IntLiteral represents an integer constant.
type IntLiteral struct {
	Value int64
}

func (n *IntLiteral) expressionNode()    {}
func (n *IntLiteral) NodeType() string   { return NodeIntLiteral }

// RationalLiteral represents a rational number (numerator/denominator).
type RationalLiteral struct {
	Numerator   int64
	Denominator int64
}

func (n *RationalLiteral) expressionNode()    {}
func (n *RationalLiteral) NodeType() string   { return NodeRationalLiteral }

// StringLiteral represents a string constant.
type StringLiteral struct {
	Value string
}

func (n *StringLiteral) expressionNode()    {}
func (n *StringLiteral) NodeType() string   { return NodeStringLiteral }

// SetLiteral represents a finite set of elements: {e1, e2, ...}.
type SetLiteral struct {
	Elements []Expression
}

func (n *SetLiteral) expressionNode()    {}
func (n *SetLiteral) NodeType() string   { return NodeSetLiteral }

// TupleLiteral represents an ordered tuple: <<e1, e2, ...>>.
type TupleLiteral struct {
	Elements []Expression `validate:"required,min=1"`
}

func (n *TupleLiteral) expressionNode()    {}
func (n *TupleLiteral) NodeType() string   { return NodeTupleLiteral }

// RecordLiteral represents a record: [field1 |-> v1, field2 |-> v2, ...].
type RecordLiteral struct {
	Fields []RecordField `validate:"required,min=1"`
}

func (n *RecordLiteral) expressionNode()    {}
func (n *RecordLiteral) NodeType() string   { return NodeRecordLiteral }

// SetConstant represents a well-known set constant (Nat, Int, Real, BOOLEAN).
type SetConstant struct {
	Kind SetConstantKind `validate:"required,oneof=nat int real boolean"`
}

func (n *SetConstant) expressionNode()    {}
func (n *SetConstant) NodeType() string   { return NodeSetConstant }

// --- References ---

// SelfRef represents a reference to the current class instance.
type SelfRef struct{}

func (n *SelfRef) expressionNode()    {}
func (n *SelfRef) NodeType() string   { return NodeSelfRef }

// AttributeRef represents a reference to a class attribute, identified by key.
type AttributeRef struct {
	AttributeKey identity.Key `validate:"required"`
}

func (n *AttributeRef) expressionNode()    {}
func (n *AttributeRef) NodeType() string   { return NodeAttributeRef }

// LocalVar represents a quantifier-bound or parameter-bound variable.
type LocalVar struct {
	Name string `validate:"required"`
}

func (n *LocalVar) expressionNode()    {}
func (n *LocalVar) NodeType() string   { return NodeLocalVar }

// PriorFieldValue represents the value of a field before a record update (replaces TLA+ @).
type PriorFieldValue struct {
	Field string `validate:"required"`
}

func (n *PriorFieldValue) expressionNode()    {}
func (n *PriorFieldValue) NodeType() string   { return NodePriorFieldValue }

// NextState wraps an expression to reference its next-state value (safety rules only).
type NextState struct {
	Expr Expression
}

func (n *NextState) expressionNode()    {}
func (n *NextState) NodeType() string   { return NodeNextState }

// --- Binary operators ---

// BinaryArith represents an arithmetic binary operation (add, sub, mul, div, mod, pow).
type BinaryArith struct {
	Op    ArithOp    `validate:"required,oneof=add sub mul div mod pow"`
	Left  Expression
	Right Expression
}

func (n *BinaryArith) expressionNode()    {}
func (n *BinaryArith) NodeType() string   { return NodeBinaryArith }

// BinaryLogic represents a logical binary operation (and, or, implies, equiv).
type BinaryLogic struct {
	Op    LogicOp    `validate:"required,oneof=and or implies equiv"`
	Left  Expression
	Right Expression
}

func (n *BinaryLogic) expressionNode()    {}
func (n *BinaryLogic) NodeType() string   { return NodeBinaryLogic }

// Compare represents a comparison operation (lt, gt, lte, gte, eq, neq).
type Compare struct {
	Op    CompareOp  `validate:"required,oneof=lt gt lte gte eq neq"`
	Left  Expression
	Right Expression
}

func (n *Compare) expressionNode()    {}
func (n *Compare) NodeType() string   { return NodeCompare }

// SetOp represents a set operation (union, intersect, difference).
type SetOp struct {
	Op    SetOpKind  `validate:"required,oneof=union intersect difference"`
	Left  Expression
	Right Expression
}

func (n *SetOp) expressionNode()    {}
func (n *SetOp) NodeType() string   { return NodeSetOp }

// SetCompare represents a set comparison operation (subset_eq, subset, superset_eq, superset).
type SetCompare struct {
	Op    SetCompareOp `validate:"required,oneof=subset_eq subset superset_eq superset"`
	Left  Expression
	Right Expression
}

func (n *SetCompare) expressionNode()    {}
func (n *SetCompare) NodeType() string   { return NodeSetCompare }

// BagOp represents a bag operation (sum, difference).
type BagOp struct {
	Op    BagOpKind  `validate:"required,oneof=sum difference"`
	Left  Expression
	Right Expression
}

func (n *BagOp) expressionNode()    {}
func (n *BagOp) NodeType() string   { return NodeBagOp }

// BagCompare represents a bag comparison operation.
type BagCompare struct {
	Op    BagCompareOp `validate:"required,oneof=proper_sub_bag sub_bag proper_sup_bag sup_bag"`
	Left  Expression
	Right Expression
}

func (n *BagCompare) expressionNode()    {}
func (n *BagCompare) NodeType() string   { return NodeBagCompare }

// Membership represents set membership (∈ or ∉).
type Membership struct {
	Element Expression
	Set     Expression
	Negated bool
}

func (n *Membership) expressionNode()    {}
func (n *Membership) NodeType() string   { return NodeMembership }

// --- Unary operators ---

// Negate represents arithmetic negation (-x).
type Negate struct {
	Expr Expression
}

func (n *Negate) expressionNode()    {}
func (n *Negate) NodeType() string   { return NodeNegate }

// Not represents logical negation (¬x).
type Not struct {
	Expr Expression
}

func (n *Not) expressionNode()    {}
func (n *Not) NodeType() string   { return NodeNot }

// --- Collections ---

// FieldAccess represents accessing a field on a base expression (base.field).
type FieldAccess struct {
	Base  Expression
	Field string     `validate:"required"`
}

func (n *FieldAccess) expressionNode()    {}
func (n *FieldAccess) NodeType() string   { return NodeFieldAccess }

// TupleIndex represents indexing into a tuple (tuple[index]).
type TupleIndex struct {
	Tuple Expression
	Index Expression
}

func (n *TupleIndex) expressionNode()    {}
func (n *TupleIndex) NodeType() string   { return NodeTupleIndex }

// RecordUpdate represents updating fields of a record (EXCEPT pattern).
type RecordUpdate struct {
	Base        Expression
	Alterations []FieldAlteration `validate:"required,min=1"`
}

func (n *RecordUpdate) expressionNode()    {}
func (n *RecordUpdate) NodeType() string   { return NodeRecordUpdate }

// StringIndex represents indexing into a string (str[index]).
type StringIndex struct {
	Str   Expression
	Index Expression
}

func (n *StringIndex) expressionNode()    {}
func (n *StringIndex) NodeType() string   { return NodeStringIndex }

// StringConcat represents string concatenation (s1 ∘ s2 ∘ ...).
type StringConcat struct {
	Operands []Expression `validate:"required,min=2"`
}

func (n *StringConcat) expressionNode()    {}
func (n *StringConcat) NodeType() string   { return NodeStringConcat }

// TupleConcat represents tuple concatenation (t1 ∘ t2 ∘ ...).
type TupleConcat struct {
	Operands []Expression `validate:"required,min=2"`
}

func (n *TupleConcat) expressionNode()    {}
func (n *TupleConcat) NodeType() string   { return NodeTupleConcat }

// --- Control flow ---

// IfThenElse represents a conditional expression (IF c THEN t ELSE e).
type IfThenElse struct {
	Condition Expression
	Then      Expression
	Else      Expression
}

func (n *IfThenElse) expressionNode()    {}
func (n *IfThenElse) NodeType() string   { return NodeIfThenElse }

// Case represents a CASE expression with branches and optional otherwise.
type Case struct {
	Branches  []CaseBranch `validate:"required,min=1"`
	Otherwise Expression
}

func (n *Case) expressionNode()    {}
func (n *Case) NodeType() string   { return NodeCase }

// --- Quantifiers ---

// Quantifier represents universal (∀) or existential (∃) quantification.
type Quantifier struct {
	Kind      QuantifierKind `validate:"required,oneof=forall exists"`
	Variable  string         `validate:"required"`
	Domain    Expression
	Predicate Expression
}

func (n *Quantifier) expressionNode()    {}
func (n *Quantifier) NodeType() string   { return NodeQuantifier }

// SetFilter represents a set filter expression: {x ∈ S : P(x)}.
type SetFilter struct {
	Variable  string     `validate:"required"`
	Set       Expression
	Predicate Expression
}

func (n *SetFilter) expressionNode()    {}
func (n *SetFilter) NodeType() string   { return NodeSetFilter }

// SetRange represents a contiguous integer range: start..end.
type SetRange struct {
	Start Expression
	End   Expression
}

func (n *SetRange) expressionNode()    {}
func (n *SetRange) NodeType() string   { return NodeSetRange }

// --- Calls ---

// ActionCall represents a call to a class action or query, identified by key.
type ActionCall struct {
	ActionKey identity.Key `validate:"required"`
	Args      []Expression
}

func (n *ActionCall) expressionNode()    {}
func (n *ActionCall) NodeType() string   { return NodeActionCall }

// GlobalCall represents a call to a global function, identified by key.
type GlobalCall struct {
	FunctionKey identity.Key `validate:"required"`
	Args        []Expression
}

func (n *GlobalCall) expressionNode()    {}
func (n *GlobalCall) NodeType() string   { return NodeGlobalCall }

// BuiltinCall represents a call to a built-in function (e.g., Len, Cardinality).
type BuiltinCall struct {
	Module   string       `validate:"required"`
	Function string       `validate:"required"`
	Args     []Expression
}

func (n *BuiltinCall) expressionNode()    {}
func (n *BuiltinCall) NodeType() string   { return NodeBuiltinCall }
