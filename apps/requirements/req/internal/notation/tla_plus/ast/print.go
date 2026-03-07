package ast

import (
	"fmt"
	"strings"
)

// Print converts a TLA+ AST expression into a TLA+ Unicode string
// with minimal parentheses based on operator precedence.
//
// Unlike the per-node String() methods, Print produces clean output
// that only parenthesizes where structurally required by precedence
// and associativity.
func Print(expr Expression) string {
	p := &printer{}
	return p.print(expr)
}

type printer struct{}

// print returns the TLA+ string for an expression.
func (p *printer) print(expr Expression) string {
	switch e := expr.(type) {
	// --- Literals ---
	case *BooleanLiteral:
		if e.Value {
			return "TRUE"
		}
		return "FALSE"

	case *NumberLiteral:
		return e.String()

	case *StringLiteral:
		return `"` + escapeString(e.Value) + `"`

	case *SetLiteral:
		return p.printDelimited("{", e.Elements, "}")

	case *SetLiteralEnum:
		parts := make([]string, len(e.Values))
		for i, v := range e.Values {
			parts[i] = `"` + escapeString(v) + `"`
		}
		return "{" + strings.Join(parts, ", ") + "}"

	case *SetLiteralInt:
		parts := make([]string, len(e.Values))
		for i, v := range e.Values {
			parts[i] = fmt.Sprintf("%d", v)
		}
		return "{" + strings.Join(parts, ", ") + "}"

	case *TupleLiteral:
		return p.printDelimited("⟨", e.Elements, "⟩")

	case *RecordInstance:
		parts := make([]string, len(e.Bindings))
		for i, b := range e.Bindings {
			parts[i] = b.Field.Value + " ↦ " + p.print(b.Expression)
		}
		return "[" + strings.Join(parts, ", ") + "]"

	case *RecordTypeExpr:
		parts := make([]string, len(e.Fields))
		for i, f := range e.Fields {
			parts[i] = f.Name.Value + ": " + p.print(f.Type)
		}
		return "[" + strings.Join(parts, ", ") + "]"

	case *SetConstant:
		return e.Value

	case *Identifier:
		return e.Value

	case *ExistingValue:
		return "@"

	// --- Parenthesized ---
	case *Parenthesized:
		return "(" + p.print(e.Inner) + ")"

	// --- Binary logic ---
	case *BinaryLogic:
		info := logicOpInfo(e.Operator)
		return p.wrap(e.Left, info.prec, info.assoc, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, info.prec, info.assoc, posRight)

	// --- Unary logic ---
	case *UnaryLogic:
		return "¬" + p.wrap(e.Right, precNot, assocPrefix, posOnly)

	// --- Binary equality ---
	case *BinaryEquality:
		return p.wrap(e.Left, precEquality, assocNone, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, precEquality, assocNone, posRight)

	// --- Binary comparison ---
	case *BinaryComparison:
		return p.wrap(e.Left, precNumCompare, assocNone, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, precNumCompare, assocNone, posRight)

	// --- Binary arithmetic ---
	case *BinaryArithmetic:
		info := arithOpInfo(e.Operator)
		return p.wrap(e.Left, info.prec, info.assoc, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, info.prec, info.assoc, posRight)

	// --- Fraction ---
	case *Fraction:
		return p.wrap(e.Numerator, precFraction, assocLeft, posLeft) +
			" / " +
			p.wrap(e.Denominator, precFraction, assocLeft, posRight)

	// --- Unary negation ---
	case *UnaryNegation:
		return "-" + p.wrap(e.Right, precNegate, assocPrefix, posOnly)

	// --- Set operations ---
	case *BinarySetOperation:
		info := setOpInfo(e.Operator)
		opStr := e.Operator
		if opStr == `\` {
			// Space around set difference to avoid ambiguity with \/ etc.
			opStr = ` \ `
		} else {
			opStr = " " + opStr + " "
		}
		return p.wrap(e.Left, info.prec, info.assoc, posLeft) +
			opStr +
			p.wrap(e.Right, info.prec, info.assoc, posRight)

	// --- Set comparison ---
	case *BinarySetComparison:
		return p.wrap(e.Left, precSetCompare, assocNone, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, precSetCompare, assocNone, posRight)

	// --- Bag operations ---
	case *BinaryBagOperation:
		info := bagOpInfo(e.Operator)
		return p.wrap(e.Left, info.prec, info.assoc, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, info.prec, info.assoc, posRight)

	// --- Bag comparison ---
	case *BinaryBagComparison:
		return p.wrap(e.Left, precBagCompare, assocNone, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, precBagCompare, assocNone, posRight)

	// --- Membership ---
	case *Membership:
		return p.wrap(e.Left, precMembership, assocNone, posLeft) +
			" " + e.Operator + " " +
			p.wrap(e.Right, precMembership, assocNone, posRight)

	// --- Cartesian product ---
	case *CartesianProduct:
		parts := make([]string, len(e.Operands))
		for i, op := range e.Operands {
			pos := posLeft
			if i == len(e.Operands)-1 {
				pos = posRight
			}
			parts[i] = p.wrap(op, precCartesian, assocLeft, pos)
		}
		return strings.Join(parts, " × ")

	// --- Set range ---
	case *SetRangeExpr:
		return p.wrap(e.Start, precRange, assocNone, posLeft) +
			".." +
			p.wrap(e.End, precRange, assocNone, posRight)

	case *SetRange:
		return fmt.Sprintf("%d..%d", e.Start, e.End)

	// --- Concat ---
	case *TupleConcat:
		return p.printNaryOp(e.Operands, "∘", precConcat, assocLeft)

	case *StringConcat:
		return p.printNaryOp(e.Operands, "∘", precConcat, assocLeft)

	// --- Field access ---
	case *FieldAccess:
		base := p.wrap(e.GetBase(), precFieldIndex, assocLeft, posLeft)
		if e.GetBase() == nil {
			base = "!"
		}
		return base + "." + e.Member

	// --- Tuple index ---
	case *TupleIndex:
		return p.wrap(e.Tuple, precFieldIndex, assocLeft, posLeft) +
			"[" + p.print(e.Index) + "]"

	// --- String index ---
	case *StringIndex:
		return p.wrap(e.Str, precFieldIndex, assocLeft, posLeft) +
			"[" + p.print(e.Index) + "]"

	// --- Prime ---
	case *Primed:
		return p.wrap(e.Base, precPrime, assocPostfix, posOnly) + "'"

	// --- Set filter ---
	case *SetFilter:
		return "{" + p.print(e.Membership) + " : " + p.print(e.Predicate) + "}"

	// --- Quantifier ---
	case *Quantifier:
		return e.Quantifier + " " + p.print(e.Membership) + " : " + p.print(e.Predicate)

	// --- IF/THEN/ELSE ---
	case *IfThenElse:
		return "IF " + p.print(e.Condition) +
			" THEN " + p.print(e.Then) +
			" ELSE " + p.print(e.Else)

	// --- CASE ---
	case *CaseExpr:
		return p.printCase(e)

	// --- Record altered (EXCEPT) ---
	case *RecordAltered:
		var sb strings.Builder
		sb.WriteString("[")
		sb.WriteString(p.print(e.Base))
		sb.WriteString(" EXCEPT ")
		for i, alt := range e.Alterations {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(alt.Field.String())
			sb.WriteString(" = ")
			sb.WriteString(p.print(alt.Expression))
		}
		sb.WriteString("]")
		return sb.String()

	// --- Function call ---
	case *FunctionCall:
		var sb strings.Builder
		for _, seg := range e.ScopePath {
			sb.WriteString(seg.Value)
			sb.WriteString("!")
		}
		sb.WriteString(e.Name.Value)
		sb.WriteString("(")
		for i, arg := range e.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.print(arg))
		}
		sb.WriteString(")")
		return sb.String()

	// --- Builtin call ---
	case *BuiltinCall:
		var sb strings.Builder
		sb.WriteString(e.Name)
		sb.WriteString("(")
		for i, arg := range e.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.print(arg))
		}
		sb.WriteString(")")
		return sb.String()

	default:
		// Fallback to the node's own String() method.
		return expr.String()
	}
}

// wrap prints a child expression, wrapping it in parentheses if needed
// given the parent's precedence and associativity.
func (p *printer) wrap(child Expression, parentPrec int, parentAssoc associativity, pos childPosition) string {
	if child == nil {
		return ""
	}
	childInfo := precedenceOf(child)
	s := p.print(child)
	if needsParens(childInfo.prec, parentPrec, pos, parentAssoc) {
		return "(" + s + ")"
	}
	return s
}

// printDelimited prints a list of expressions with delimiters.
func (p *printer) printDelimited(open string, elems []Expression, closeStr string) string {
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = p.print(e)
	}
	return open + strings.Join(parts, ", ") + closeStr
}

// printNaryOp prints n-ary operators (concat, etc.) with correct wrapping.
func (p *printer) printNaryOp(operands []Expression, op string, prec int, assoc associativity) string {
	parts := make([]string, len(operands))
	for i, operand := range operands {
		pos := posLeft
		if i == len(operands)-1 {
			pos = posRight
		}
		parts[i] = p.wrap(operand, prec, assoc, pos)
	}
	return strings.Join(parts, " "+op+" ")
}

// printCase prints a CASE expression.
// CASE conditions and results are parsed at OrExpr level, so we must
// wrap them if they contain implies (⇒) or equiv (≡) at the top level.
func (p *printer) printCase(e *CaseExpr) string {
	var sb strings.Builder
	sb.WriteString("CASE ")
	for i, branch := range e.Branches {
		if i > 0 {
			sb.WriteString(" □ ")
		}
		sb.WriteString(p.wrapCaseExpr(branch.Condition))
		sb.WriteString(" → ")
		sb.WriteString(p.wrapCaseExpr(branch.Result))
	}
	if e.Other != nil {
		sb.WriteString(" □ OTHER → ")
		sb.WriteString(p.wrapCaseExpr(e.Other))
	}
	return sb.String()
}

// wrapCaseExpr wraps a CASE condition/result expression in parentheses
// if it would be ambiguous at the OrExpr precedence level.
func (p *printer) wrapCaseExpr(expr Expression) string {
	info := precedenceOf(expr)
	s := p.print(expr)
	// CASE internals are parsed at OrExpr level, so anything with
	// precedence below precOr (i.e. implies or equiv) must be wrapped.
	if info.prec < precOr {
		return "(" + s + ")"
	}
	return s
}

// escapeString escapes special characters in a TLA+ string literal.
func escapeString(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		case '\r':
			sb.WriteString(`\r`)
		case '\f':
			sb.WriteString(`\f`)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
