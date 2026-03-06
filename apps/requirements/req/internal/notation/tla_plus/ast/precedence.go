package ast

// Precedence and associativity data for TLA+ operators.
//
// The precedence hierarchy is derived from the PEG grammar
// (parser/peg/tla_expression.peg). Lower numbers bind less tightly.
// This data drives the pretty-printer's parenthesization decisions.

// associativity describes how an operator groups with repeated application.
type associativity int

const (
	assocLeft    associativity = iota // (a ∘ b) ∘ c
	assocRight                        // a ⇒ (b ⇒ c)
	assocNone                         // non-associative (comparisons, membership)
	assocPrefix                       // unary prefix (¬, -)
	assocPostfix                      // postfix (')
)

// childPosition identifies which operand of a parent node a child occupies.
type childPosition int

const (
	posLeft  childPosition = iota // left operand of binary op
	posRight                      // right operand of binary op
	posOnly                       // sole operand of unary op
)

// Precedence levels. Each constant corresponds to a named grammar rule.
// Lower number = binds less tightly.
const (
	precImplies    = 1  // ⇒  right-associative
	precEquiv      = 2  // ≡  left-associative
	precOr         = 3  // ∨  left-associative
	precAnd        = 4  // ∧  left-associative
	precNot        = 5  // ¬  prefix
	precQuantifier = 6  // ∀/∃  prefix
	precMembership = 7  // ∈/∉  non-associative
	precSetCompare = 8  // ⊆/⊂/⊇/⊃  non-associative
	precBagCompare = 9  // ⊏/⊑/⊐/⊒  non-associative
	precEquality   = 10 // =/≠  non-associative
	precNumCompare = 11 // </>/≤/≥  non-associative
	precSetDiff    = 12 // \  left-associative
	precIntersect  = 13 // ∩  left-associative
	precUnion      = 14 // ∪  left-associative
	precCartesian  = 15 // ×  left-associative
	precRange      = 16 // ..  non-associative
	precBagSum     = 17 // ⊕  left-associative
	precModulo     = 18 // %  left-associative
	precAdd        = 19 // +  left-associative
	precBagDiff    = 20 // ⊖  left-associative
	precSub        = 21 // -  left-associative
	precNegate     = 22 // - (prefix)
	precDiv        = 23 // ÷  left-associative
	precConcat     = 24 // ∘  left-associative
	precMul        = 25 // *  left-associative
	precFraction   = 26 // /  left-associative
	precPow        = 27 // ^  right-associative
	precPrime      = 28 // '  postfix
	precFieldIndex = 29 // . []  left-associative
	precAtom       = 30 // literals, identifiers, parens, etc.
)

// opInfo holds the precedence and associativity for an operator.
type opInfo struct {
	prec  int
	assoc associativity
}

// precedenceOf returns the precedence and associativity of an AST expression.
// Atomic/compound expressions (literals, IF/THEN/ELSE, CASE, etc.) return
// precAtom — they never need wrapping as children.
func precedenceOf(expr Expression) opInfo {
	switch e := expr.(type) {
	// --- Binary logic ---
	case *BinaryLogic:
		return logicOpInfo(e.Operator)
	// --- Unary logic ---
	case *UnaryLogic:
		return opInfo{precNot, assocPrefix}
	// --- Quantifiers ---
	case *Quantifier:
		return opInfo{precQuantifier, assocPrefix}
	// --- Membership ---
	case *Membership:
		return opInfo{precMembership, assocNone}
	// --- Set comparison ---
	case *BinarySetComparison:
		return opInfo{precSetCompare, assocNone}
	// --- Bag comparison ---
	case *BinaryBagComparison:
		return opInfo{precBagCompare, assocNone}
	// --- Equality ---
	case *BinaryEquality:
		return opInfo{precEquality, assocNone}
	// --- Numeric comparison ---
	case *BinaryComparison:
		return opInfo{precNumCompare, assocNone}
	// --- Set operations ---
	case *BinarySetOperation:
		return setOpInfo(e.Operator)
	// --- Cartesian product ---
	case *CartesianProduct:
		return opInfo{precCartesian, assocLeft}
	// --- Bag operations ---
	case *BinaryBagOperation:
		return bagOpInfo(e.Operator)
	// --- Arithmetic ---
	case *BinaryArithmetic:
		return arithOpInfo(e.Operator)
	// --- Fraction ---
	case *Fraction:
		return opInfo{precFraction, assocLeft}
	// --- Unary negation ---
	case *UnaryNegation:
		return opInfo{precNegate, assocPrefix}
	// --- Concat ---
	case *TupleConcat:
		return opInfo{precConcat, assocLeft}
	case *StringConcat:
		return opInfo{precConcat, assocLeft}
	// --- Power ---
	// (handled in arithOpInfo via BinaryArithmetic)
	// --- Prime ---
	case *Primed:
		return opInfo{precPrime, assocPostfix}
	// --- Field/index access ---
	case *FieldAccess:
		return opInfo{precFieldIndex, assocLeft}
	case *TupleIndex:
		return opInfo{precFieldIndex, assocLeft}
	case *StringIndex:
		return opInfo{precFieldIndex, assocLeft}
	// --- Set range ---
	case *SetRangeExpr:
		return opInfo{precRange, assocNone}
	// --- Everything else is atomic ---
	default:
		return opInfo{precAtom, assocNone}
	}
}

// logicOpInfo maps a logic operator string to its opInfo.
func logicOpInfo(op string) opInfo {
	switch op {
	case "⇒":
		return opInfo{precImplies, assocRight}
	case "≡":
		return opInfo{precEquiv, assocLeft}
	case "∨":
		return opInfo{precOr, assocLeft}
	case "∧":
		return opInfo{precAnd, assocLeft}
	default:
		return opInfo{precAtom, assocNone}
	}
}

// setOpInfo maps a set operation operator string to its opInfo.
func setOpInfo(op string) opInfo {
	switch op {
	case `\`:
		return opInfo{precSetDiff, assocLeft}
	case "∩":
		return opInfo{precIntersect, assocLeft}
	case "∪":
		return opInfo{precUnion, assocLeft}
	default:
		return opInfo{precAtom, assocNone}
	}
}

// bagOpInfo maps a bag operation operator string to its opInfo.
func bagOpInfo(op string) opInfo {
	switch op {
	case "⊕":
		return opInfo{precBagSum, assocLeft}
	case "⊖":
		return opInfo{precBagDiff, assocLeft}
	default:
		return opInfo{precAtom, assocNone}
	}
}

// arithOpInfo maps an arithmetic operator string to its opInfo.
func arithOpInfo(op string) opInfo {
	switch op {
	case "+":
		return opInfo{precAdd, assocLeft}
	case "-":
		return opInfo{precSub, assocLeft}
	case "*":
		return opInfo{precMul, assocLeft}
	case "÷":
		return opInfo{precDiv, assocLeft}
	case "%":
		return opInfo{precModulo, assocLeft}
	case "^":
		return opInfo{precPow, assocRight}
	default:
		return opInfo{precAtom, assocNone}
	}
}

// needsParens determines whether a child expression must be wrapped in
// parentheses when it appears at the given position within a parent
// expression of the given precedence and associativity.
//
// Rules:
//  1. child binds less tightly → must wrap
//  2. child binds more tightly → no wrap needed
//  3. equal precedence → depends on associativity:
//     - assocLeft:  wrap right child only
//     - assocRight: wrap left child only
//     - assocNone:  always wrap (non-associative operators)
//     - assocPrefix/assocPostfix: no wrap
func needsParens(childPrec, parentPrec int, pos childPosition, parentAssoc associativity) bool {
	if childPrec < parentPrec {
		return true
	}
	if childPrec > parentPrec {
		return false
	}
	// Equal precedence — associativity decides.
	switch parentAssoc {
	case assocLeft:
		return pos == posRight
	case assocRight:
		return pos == posLeft
	case assocNone:
		return true
	default: // prefix, postfix
		return false
	}
}
