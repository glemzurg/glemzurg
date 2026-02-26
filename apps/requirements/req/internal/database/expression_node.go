package database

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression"
)

// Synthetic node types used only in the database to represent composite structures.
// These have no corresponding Go Expression type.
const (
	dbNodeFieldAlteration = "field_alteration"
	dbNodeCaseBranch      = "case_branch"
)

// exprNodeRow is a flattened database row for an expression node.
type exprNodeRow struct {
	logicKey      identity.Key
	parentNodeKey *string // nil for root
	sortOrder     int
	nodeKey       string // synthetic key for this node
	nodeType      string

	// Scalar values.
	boolValue   *bool
	intValue    *int64
	numerator   *int64
	denominator *int64
	stringValue *string

	// Operator.
	operator *string

	// Model references.
	attributeKey      *string
	actionKey         *string
	globalFunctionKey *string
	builtinModule     *string
	builtinFunction   *string

	// Quantifier metadata.
	quantifierKind *string
	variableName   *string

	// Set constant kind.
	setConstantKind *string

	// Membership negation.
	negated *bool
}

// scanExprNode scans a database row into an exprNodeRow.
func scanExprNode(scanner Scanner) (row exprNodeRow, err error) {
	var logicKeyStr string

	if err = scanner.Scan(
		&row.nodeKey,
		&logicKeyStr,
		&row.parentNodeKey,
		&row.sortOrder,
		&row.nodeType,
		&row.boolValue,
		&row.intValue,
		&row.numerator,
		&row.denominator,
		&row.stringValue,
		&row.operator,
		&row.attributeKey,
		&row.actionKey,
		&row.globalFunctionKey,
		&row.builtinModule,
		&row.builtinFunction,
		&row.quantifierKind,
		&row.variableName,
		&row.setConstantKind,
		&row.negated,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return exprNodeRow{}, err
	}

	row.logicKey, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return exprNodeRow{}, err
	}

	return row, nil
}

// QueryExpressionNodes loads all expression nodes for a model, reconstructs trees,
// and returns a map from logic key to root Expression.
func QueryExpressionNodes(dbOrTx DbOrTx, modelKey string) (expressions map[identity.Key]me.Expression, err error) {
	var rows []exprNodeRow

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			row, err := scanExprNode(scanner)
			if err != nil {
				return errors.WithStack(err)
			}
			rows = append(rows, row)
			return nil
		},
		`SELECT
			expression_node_key,
			logic_key,
			parent_node_key,
			sort_order,
			node_type,
			bool_value,
			int_value,
			numerator,
			denominator,
			string_value,
			operator,
			attribute_key,
			action_key,
			global_function_key,
			builtin_module,
			builtin_function,
			quantifier_kind,
			variable_name,
			set_constant_kind,
			negated
		FROM
			expression_node
		WHERE
			model_key = $1
		ORDER BY logic_key, parent_node_key NULLS FIRST, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	expressions, err = rebuildExprTrees(rows)
	if err != nil {
		return nil, err
	}

	return expressions, nil
}

// rebuildExprTrees reconstructs expression trees from flat rows.
func rebuildExprTrees(rows []exprNodeRow) (map[identity.Key]me.Expression, error) {
	rowByKey := make(map[string]*exprNodeRow, len(rows))
	childrenByParent := make(map[string][]string)
	rootsByLogic := make(map[identity.Key][]string)

	for i := range rows {
		row := &rows[i]
		rowByKey[row.nodeKey] = row

		if row.parentNodeKey == nil {
			rootsByLogic[row.logicKey] = append(rootsByLogic[row.logicKey], row.nodeKey)
		} else {
			childrenByParent[*row.parentNodeKey] = append(childrenByParent[*row.parentNodeKey], row.nodeKey)
		}
	}

	ctx := &rebuildContext{rowByKey: rowByKey, childrenByParent: childrenByParent}

	result := make(map[identity.Key]me.Expression)
	for logicKey, roots := range rootsByLogic {
		if len(roots) != 1 {
			return nil, fmt.Errorf("logic %q has %d root expression nodes, expected 1", logicKey.String(), len(roots))
		}
		expr, err := ctx.buildNode(roots[0])
		if err != nil {
			return nil, errors.Wrapf(err, "logic %q expression tree", logicKey.String())
		}
		result[logicKey] = expr
	}

	return result, nil
}

// rebuildContext holds the indexed data needed during tree reconstruction.
type rebuildContext struct {
	rowByKey         map[string]*exprNodeRow
	childrenByParent map[string][]string
}

// buildNode recursively builds an Expression from a node key.
func (ctx *rebuildContext) buildNode(key string) (me.Expression, error) {
	row := ctx.rowByKey[key]
	childKeys := ctx.childrenByParent[key]

	switch row.nodeType {
	// --- Literals ---
	case me.NodeBoolLiteral:
		return &me.BoolLiteral{Value: derefBool(row.boolValue)}, nil
	case me.NodeIntLiteral:
		return &me.IntLiteral{Value: derefInt64(row.intValue)}, nil
	case me.NodeRationalLiteral:
		return &me.RationalLiteral{Numerator: derefInt64(row.numerator), Denominator: derefInt64(row.denominator)}, nil
	case me.NodeStringLiteral:
		return &me.StringLiteral{Value: derefString(row.stringValue)}, nil
	case me.NodeSetLiteral:
		children, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.SetLiteral{Elements: children}, nil
	case me.NodeTupleLiteral:
		children, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.TupleLiteral{Elements: children}, nil
	case me.NodeRecordLiteral:
		// Each child is a value expression. The field name is stored in the child row's
		// variable_name column (repurposed for record field names).
		fields := make([]me.RecordField, len(childKeys))
		for i, childKey := range childKeys {
			childRow := ctx.rowByKey[childKey]
			valueExpr, err := ctx.buildNode(childKey)
			if err != nil {
				return nil, err
			}
			fields[i] = me.RecordField{
				Name:  derefString(childRow.variableName),
				Value: valueExpr,
			}
		}
		return &me.RecordLiteral{Fields: fields}, nil
	case me.NodeSetConstant:
		return &me.SetConstant{Kind: me.SetConstantKind(derefString(row.setConstantKind))}, nil

	// --- References ---
	case me.NodeSelfRef:
		return &me.SelfRef{}, nil
	case me.NodeAttributeRef:
		attrKey, err := identity.ParseKey(derefString(row.attributeKey))
		if err != nil {
			return nil, fmt.Errorf("attribute_ref: %w", err)
		}
		return &me.AttributeRef{AttributeKey: attrKey}, nil
	case me.NodeLocalVar:
		return &me.LocalVar{Name: derefString(row.stringValue)}, nil
	case me.NodePriorFieldValue:
		return &me.PriorFieldValue{Field: derefString(row.stringValue)}, nil
	case me.NodeNextState:
		expr, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		return &me.NextState{Expr: expr}, nil

	// --- Binary operators ---
	case me.NodeBinaryArith:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.BinaryArith{Op: me.ArithOp(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeBinaryLogic:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.BinaryLogic{Op: me.LogicOp(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeCompare:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.Compare{Op: me.CompareOp(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeSetOp:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.SetOp{Op: me.SetOpKind(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeSetCompare:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.SetCompare{Op: me.SetCompareOp(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeBagOp:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.BagOp{Op: me.BagOpKind(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeBagCompare:
		left, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		right, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.BagCompare{Op: me.BagCompareOp(derefString(row.operator)), Left: left, Right: right}, nil
	case me.NodeMembership:
		elem, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		set, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.Membership{Element: elem, Set: set, Negated: derefBool(row.negated)}, nil

	// --- Unary operators ---
	case me.NodeNegate:
		expr, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		return &me.Negate{Expr: expr}, nil
	case me.NodeNot:
		expr, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		return &me.Not{Expr: expr}, nil

	// --- Collections ---
	case me.NodeFieldAccess:
		base, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		return &me.FieldAccess{Base: base, Field: derefString(row.stringValue)}, nil
	case me.NodeTupleIndex:
		tuple, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		index, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.TupleIndex{Tuple: tuple, Index: index}, nil
	case me.NodeRecordUpdate:
		// sort_order 0 = base expression, rest are field_alteration synthetic nodes.
		base, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		alts := make([]me.FieldAlteration, 0, len(childKeys)-1)
		for i := 1; i < len(childKeys); i++ {
			alt, err := ctx.buildFieldAlteration(childKeys[i])
			if err != nil {
				return nil, err
			}
			alts = append(alts, alt)
		}
		return &me.RecordUpdate{Base: base, Alterations: alts}, nil
	case me.NodeStringIndex:
		str, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		index, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.StringIndex{Str: str, Index: index}, nil
	case me.NodeStringConcat:
		children, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.StringConcat{Operands: children}, nil
	case me.NodeTupleConcat:
		children, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.TupleConcat{Operands: children}, nil

	// --- Control flow ---
	case me.NodeIfThenElse:
		cond, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		then, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		elseExpr, err := ctx.buildChildAt(childKeys, 2)
		if err != nil {
			return nil, err
		}
		return &me.IfThenElse{Condition: cond, Then: then, Else: elseExpr}, nil
	case me.NodeCase:
		// Children are case_branch synthetic nodes. The negated column on the case row
		// indicates whether an otherwise clause exists. If true, the last child is the
		// otherwise expression (a normal expression node, not a case_branch).
		hasOtherwise := derefBool(row.negated)
		branchCount := len(childKeys)
		var otherwise me.Expression
		if hasOtherwise && branchCount > 0 {
			branchCount--
			var err error
			otherwise, err = ctx.buildNode(childKeys[branchCount])
			if err != nil {
				return nil, err
			}
		}
		branches := make([]me.CaseBranch, branchCount)
		for i := 0; i < branchCount; i++ {
			branch, err := ctx.buildCaseBranch(childKeys[i])
			if err != nil {
				return nil, err
			}
			branches[i] = branch
		}
		return &me.Case{Branches: branches, Otherwise: otherwise}, nil

	// --- Quantifiers ---
	case me.NodeQuantifier:
		domain, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		predicate, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.Quantifier{
			Kind:      me.QuantifierKind(derefString(row.quantifierKind)),
			Variable:  derefString(row.variableName),
			Domain:    domain,
			Predicate: predicate,
		}, nil
	case me.NodeSetFilter:
		set, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		predicate, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.SetFilter{
			Variable:  derefString(row.variableName),
			Set:       set,
			Predicate: predicate,
		}, nil
	case me.NodeSetRange:
		start, err := ctx.buildChildAt(childKeys, 0)
		if err != nil {
			return nil, err
		}
		end, err := ctx.buildChildAt(childKeys, 1)
		if err != nil {
			return nil, err
		}
		return &me.SetRange{Start: start, End: end}, nil

	// --- Calls ---
	case me.NodeActionCall:
		actionKey, err := identity.ParseKey(derefString(row.actionKey))
		if err != nil {
			return nil, fmt.Errorf("action_call: %w", err)
		}
		args, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.ActionCall{ActionKey: actionKey, Args: args}, nil
	case me.NodeGlobalCall:
		funcKey, err := identity.ParseKey(derefString(row.globalFunctionKey))
		if err != nil {
			return nil, fmt.Errorf("global_call: %w", err)
		}
		args, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.GlobalCall{FunctionKey: funcKey, Args: args}, nil
	case me.NodeBuiltinCall:
		args, err := ctx.buildAllChildren(childKeys)
		if err != nil {
			return nil, err
		}
		return &me.BuiltinCall{
			Module:   derefString(row.builtinModule),
			Function: derefString(row.builtinFunction),
			Args:     args,
		}, nil

	default:
		return nil, fmt.Errorf("unknown expression node type: %q", row.nodeType)
	}
}

// buildAllChildren builds all child keys as Expression nodes.
func (ctx *rebuildContext) buildAllChildren(childKeys []string) ([]me.Expression, error) {
	result := make([]me.Expression, len(childKeys))
	for i, key := range childKeys {
		expr, err := ctx.buildNode(key)
		if err != nil {
			return nil, err
		}
		result[i] = expr
	}
	return result, nil
}

// buildChildAt builds a single child by index, returning nil if out of range.
func (ctx *rebuildContext) buildChildAt(childKeys []string, i int) (me.Expression, error) {
	if i >= len(childKeys) {
		return nil, nil
	}
	return ctx.buildNode(childKeys[i])
}

// buildFieldAlteration builds a FieldAlteration from a field_alteration synthetic node.
// The field name is in string_value, the single child is the value expression.
func (ctx *rebuildContext) buildFieldAlteration(key string) (me.FieldAlteration, error) {
	row := ctx.rowByKey[key]
	if row.nodeType != dbNodeFieldAlteration {
		return me.FieldAlteration{}, fmt.Errorf("expected field_alteration node, got %q", row.nodeType)
	}
	altChildKeys := ctx.childrenByParent[key]
	if len(altChildKeys) != 1 {
		return me.FieldAlteration{}, fmt.Errorf("field_alteration node expected 1 child, got %d", len(altChildKeys))
	}
	valueExpr, err := ctx.buildNode(altChildKeys[0])
	if err != nil {
		return me.FieldAlteration{}, err
	}
	return me.FieldAlteration{
		Field: derefString(row.stringValue),
		Value: valueExpr,
	}, nil
}

// buildCaseBranch builds a CaseBranch from a case_branch synthetic node.
// Child 0 = condition, child 1 = result.
func (ctx *rebuildContext) buildCaseBranch(key string) (me.CaseBranch, error) {
	row := ctx.rowByKey[key]
	if row.nodeType != dbNodeCaseBranch {
		return me.CaseBranch{}, fmt.Errorf("expected case_branch node, got %q", row.nodeType)
	}
	branchChildKeys := ctx.childrenByParent[key]
	if len(branchChildKeys) != 2 {
		return me.CaseBranch{}, fmt.Errorf("case_branch node expected 2 children, got %d", len(branchChildKeys))
	}
	condition, err := ctx.buildNode(branchChildKeys[0])
	if err != nil {
		return me.CaseBranch{}, err
	}
	result, err := ctx.buildNode(branchChildKeys[1])
	if err != nil {
		return me.CaseBranch{}, err
	}
	return me.CaseBranch{
		Condition: condition,
		Result:    result,
	}, nil
}

// --- Flatten: Expression tree -> flat rows ---

// FlattenExpression walks an expression tree and produces flat exprNodeRow slices
// in topological order (parent before children).
func FlattenExpression(logicKey identity.Key, expr me.Expression) []exprNodeRow {
	var rows []exprNodeRow
	counter := 0
	flattenExprRecursive(logicKey, nil, expr, 0, &rows, &counter, "")
	return rows
}

// flattenExprRecursive recursively flattens an expression tree.
// fieldName is used when flattening children of record_literal (stores the field name in variable_name).
func flattenExprRecursive(logicKey identity.Key, parentKey *string, expr me.Expression, sortOrder int, rows *[]exprNodeRow, counter *int, fieldName string) string {
	*counter++
	nodeKey := fmt.Sprintf("%s/expr/%d", logicKey.String(), *counter)

	row := exprNodeRow{
		logicKey:      logicKey,
		parentNodeKey: parentKey,
		sortOrder:     sortOrder,
		nodeKey:       nodeKey,
		nodeType:      expr.NodeType(),
	}

	// Set field name in variable_name for record_literal children.
	if fieldName != "" {
		row.variableName = exprStrPtr(fieldName)
	}

	switch n := expr.(type) {
	// --- Literals ---
	case *me.BoolLiteral:
		row.boolValue = &n.Value
	case *me.IntLiteral:
		row.intValue = &n.Value
	case *me.RationalLiteral:
		row.numerator = &n.Numerator
		row.denominator = &n.Denominator
	case *me.StringLiteral:
		row.stringValue = exprStrPtr(n.Value)
	case *me.SetLiteral:
		*rows = append(*rows, row)
		for i, elem := range n.Elements {
			flattenExprRecursive(logicKey, &nodeKey, elem, i, rows, counter, "")
		}
		return nodeKey
	case *me.TupleLiteral:
		*rows = append(*rows, row)
		for i, elem := range n.Elements {
			flattenExprRecursive(logicKey, &nodeKey, elem, i, rows, counter, "")
		}
		return nodeKey
	case *me.RecordLiteral:
		*rows = append(*rows, row)
		for i, field := range n.Fields {
			flattenExprRecursive(logicKey, &nodeKey, field.Value, i, rows, counter, field.Name)
		}
		return nodeKey
	case *me.SetConstant:
		row.setConstantKind = exprStrPtr(string(n.Kind))

	// --- References ---
	case *me.SelfRef:
		// No additional fields.
	case *me.AttributeRef:
		row.attributeKey = exprStrPtr(n.AttributeKey.String())
	case *me.LocalVar:
		row.stringValue = exprStrPtr(n.Name)
	case *me.PriorFieldValue:
		row.stringValue = exprStrPtr(n.Field)
	case *me.NextState:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Expr, 0, rows, counter, "")
		return nodeKey

	// --- Binary operators ---
	case *me.BinaryArith:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.BinaryLogic:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.Compare:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.SetOp:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.SetCompare:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.BagOp:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.BagCompare:
		row.operator = exprStrPtr(string(n.Op))
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Left, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Right, 1, rows, counter, "")
		return nodeKey
	case *me.Membership:
		row.negated = &n.Negated
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Element, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Set, 1, rows, counter, "")
		return nodeKey

	// --- Unary operators ---
	case *me.Negate:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Expr, 0, rows, counter, "")
		return nodeKey
	case *me.Not:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Expr, 0, rows, counter, "")
		return nodeKey

	// --- Collections ---
	case *me.FieldAccess:
		row.stringValue = exprStrPtr(n.Field)
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Base, 0, rows, counter, "")
		return nodeKey
	case *me.TupleIndex:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Tuple, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Index, 1, rows, counter, "")
		return nodeKey
	case *me.RecordUpdate:
		*rows = append(*rows, row)
		// sort_order 0 = base expression.
		flattenExprRecursive(logicKey, &nodeKey, n.Base, 0, rows, counter, "")
		// sort_order 1+ = field_alteration synthetic nodes.
		for i, alt := range n.Alterations {
			flattenFieldAlteration(logicKey, &nodeKey, alt, i+1, rows, counter)
		}
		return nodeKey
	case *me.StringIndex:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Str, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Index, 1, rows, counter, "")
		return nodeKey
	case *me.StringConcat:
		*rows = append(*rows, row)
		for i, op := range n.Operands {
			flattenExprRecursive(logicKey, &nodeKey, op, i, rows, counter, "")
		}
		return nodeKey
	case *me.TupleConcat:
		*rows = append(*rows, row)
		for i, op := range n.Operands {
			flattenExprRecursive(logicKey, &nodeKey, op, i, rows, counter, "")
		}
		return nodeKey

	// --- Control flow ---
	case *me.IfThenElse:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Condition, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Then, 1, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Else, 2, rows, counter, "")
		return nodeKey
	case *me.Case:
		// negated column repurposed: true = has otherwise clause.
		hasOtherwise := n.Otherwise != nil
		row.negated = &hasOtherwise
		*rows = append(*rows, row)
		for i, branch := range n.Branches {
			flattenCaseBranch(logicKey, &nodeKey, branch, i, rows, counter)
		}
		if n.Otherwise != nil {
			flattenExprRecursive(logicKey, &nodeKey, n.Otherwise, len(n.Branches), rows, counter, "")
		}
		return nodeKey

	// --- Quantifiers ---
	case *me.Quantifier:
		row.quantifierKind = exprStrPtr(string(n.Kind))
		row.variableName = exprStrPtr(n.Variable)
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Domain, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Predicate, 1, rows, counter, "")
		return nodeKey
	case *me.SetFilter:
		row.variableName = exprStrPtr(n.Variable)
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Set, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.Predicate, 1, rows, counter, "")
		return nodeKey
	case *me.SetRange:
		*rows = append(*rows, row)
		flattenExprRecursive(logicKey, &nodeKey, n.Start, 0, rows, counter, "")
		flattenExprRecursive(logicKey, &nodeKey, n.End, 1, rows, counter, "")
		return nodeKey

	// --- Calls ---
	case *me.ActionCall:
		row.actionKey = exprStrPtr(n.ActionKey.String())
		*rows = append(*rows, row)
		for i, arg := range n.Args {
			flattenExprRecursive(logicKey, &nodeKey, arg, i, rows, counter, "")
		}
		return nodeKey
	case *me.GlobalCall:
		row.globalFunctionKey = exprStrPtr(n.FunctionKey.String())
		*rows = append(*rows, row)
		for i, arg := range n.Args {
			flattenExprRecursive(logicKey, &nodeKey, arg, i, rows, counter, "")
		}
		return nodeKey
	case *me.BuiltinCall:
		row.builtinModule = exprStrPtr(n.Module)
		row.builtinFunction = exprStrPtr(n.Function)
		*rows = append(*rows, row)
		for i, arg := range n.Args {
			flattenExprRecursive(logicKey, &nodeKey, arg, i, rows, counter, "")
		}
		return nodeKey
	}

	// Leaf node (no children added above).
	*rows = append(*rows, row)
	return nodeKey
}

// flattenFieldAlteration creates a synthetic field_alteration node row with the value as its child.
func flattenFieldAlteration(logicKey identity.Key, parentKey *string, alt me.FieldAlteration, sortOrder int, rows *[]exprNodeRow, counter *int) {
	*counter++
	nodeKey := fmt.Sprintf("%s/expr/%d", logicKey.String(), *counter)

	altRow := exprNodeRow{
		logicKey:      logicKey,
		parentNodeKey: parentKey,
		sortOrder:     sortOrder,
		nodeKey:       nodeKey,
		nodeType:      dbNodeFieldAlteration,
		stringValue:   exprStrPtr(alt.Field),
	}
	*rows = append(*rows, altRow)

	flattenExprRecursive(logicKey, &nodeKey, alt.Value, 0, rows, counter, "")
}

// flattenCaseBranch creates a synthetic case_branch node row with condition and result as children.
func flattenCaseBranch(logicKey identity.Key, parentKey *string, branch me.CaseBranch, sortOrder int, rows *[]exprNodeRow, counter *int) {
	*counter++
	nodeKey := fmt.Sprintf("%s/expr/%d", logicKey.String(), *counter)

	branchRow := exprNodeRow{
		logicKey:      logicKey,
		parentNodeKey: parentKey,
		sortOrder:     sortOrder,
		nodeKey:       nodeKey,
		nodeType:      dbNodeCaseBranch,
	}
	*rows = append(*rows, branchRow)

	flattenExprRecursive(logicKey, &nodeKey, branch.Condition, 0, rows, counter, "")
	flattenExprRecursive(logicKey, &nodeKey, branch.Result, 1, rows, counter, "")
}

// --- Add/Delete ---

// AddExpressionNodes adds multiple expression node rows to the database.
// Rows must be in topological order (parent before children) due to self-referential FK.
func AddExpressionNodes(dbOrTx DbOrTx, modelKey string, rows []exprNodeRow) (err error) {
	if len(rows) == 0 {
		return nil
	}

	query := `INSERT INTO expression_node (model_key, expression_node_key, logic_key, parent_node_key, sort_order, node_type, bool_value, int_value, numerator, denominator, string_value, operator, attribute_key, action_key, global_function_key, builtin_module, builtin_function, quantifier_kind, variable_name, set_constant_kind, negated) VALUES `
	args := make([]interface{}, 0, len(rows)*21)

	for i, row := range rows {
		if i > 0 {
			query += ", "
		}
		base := i * 21
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7,
			base+8, base+9, base+10, base+11, base+12, base+13, base+14,
			base+15, base+16, base+17, base+18, base+19, base+20, base+21)

		args = append(args,
			modelKey,
			row.nodeKey,
			row.logicKey.String(),
			row.parentNodeKey,
			row.sortOrder,
			row.nodeType,
			row.boolValue,
			row.intValue,
			row.numerator,
			row.denominator,
			row.stringValue,
			row.operator,
			row.attributeKey,
			row.actionKey,
			row.globalFunctionKey,
			row.builtinModule,
			row.builtinFunction,
			row.quantifierKind,
			row.variableName,
			row.setConstantKind,
			row.negated,
		)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeleteExpressionNodes deletes all expression nodes for a given logic.
func DeleteExpressionNodes(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			expression_node
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// --- Helpers ---

// strPtr returns a pointer to s, or nil if s is empty.
func exprStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefBool returns the value of a *bool or false if nil.
func derefBool(p *bool) bool {
	if p != nil {
		return *p
	}
	return false
}

// derefInt64 returns the value of a *int64 or 0 if nil.
func derefInt64(p *int64) int64 {
	if p != nil {
		return *p
	}
	return 0
}

// derefString returns the value of a *string or "" if nil.
func derefString(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}
