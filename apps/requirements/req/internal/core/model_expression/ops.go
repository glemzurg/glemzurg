package model_expression

// ArithOp represents arithmetic operators.
type ArithOp string

const (
	ArithAdd = ArithOp("add")
	ArithSub = ArithOp("sub")
	ArithMul = ArithOp("mul")
	ArithDiv = ArithOp("div")
	ArithMod = ArithOp("mod")
	ArithPow = ArithOp("pow")
)

// LogicOp represents logical operators.
type LogicOp string

const (
	LogicAnd     = LogicOp("and")
	LogicOr      = LogicOp("or")
	LogicImplies = LogicOp("implies")
	LogicEquiv   = LogicOp("equiv")
)

// CompareOp represents comparison operators.
type CompareOp string

const (
	CompareLt  = CompareOp("lt")
	CompareGt  = CompareOp("gt")
	CompareLte = CompareOp("lte")
	CompareGte = CompareOp("gte")
	CompareEq  = CompareOp("eq")
	CompareNeq = CompareOp("neq")
)

// SetOpKind represents set operators.
type SetOpKind string

const (
	SetUnion      = SetOpKind("union")
	SetIntersect  = SetOpKind("intersect")
	SetDifference = SetOpKind("difference")
)

// SetCompareOp represents set comparison operators.
type SetCompareOp string

const (
	SetCompareSubsetEq   = SetCompareOp("subset_eq")
	SetCompareSubset     = SetCompareOp("subset")
	SetCompareSupersetEq = SetCompareOp("superset_eq")
	SetCompareSuperset   = SetCompareOp("superset")
)

// BagOpKind represents bag operators.
type BagOpKind string

const (
	BagSum        = BagOpKind("sum")
	BagDifference = BagOpKind("difference")
)

// BagCompareOp represents bag comparison operators.
type BagCompareOp string

const (
	BagCompareProperSubBag = BagCompareOp("proper_sub_bag")
	BagCompareSubBag       = BagCompareOp("sub_bag")
	BagCompareProperSupBag = BagCompareOp("proper_sup_bag")
	BagCompareSupBag       = BagCompareOp("sup_bag")
)

// QuantifierKind represents quantifier types.
type QuantifierKind string

const (
	QuantifierForall = QuantifierKind("forall")
	QuantifierExists = QuantifierKind("exists")
)

// SetConstantKind represents well-known set constants.
type SetConstantKind string

const (
	SetConstantNat     = SetConstantKind("nat")
	SetConstantInt     = SetConstantKind("int")
	SetConstantReal    = SetConstantKind("real")
	SetConstantBoolean = SetConstantKind("boolean")
)
