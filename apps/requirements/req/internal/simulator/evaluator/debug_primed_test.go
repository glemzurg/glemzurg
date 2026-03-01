package evaluator

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

func TestDebugPrimed(t *testing.T) {
	expr, err := parser.ParseExpression("record.value' > record.value")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	fmt.Printf("Parsed AST: %s\n", expr.String())
	fmt.Printf("AST type: %T\n", expr)

	// Inspect the AST structure
	comp, ok := expr.(*ast.BinaryComparison)
	if !ok {
		t.Fatalf("Expected BinaryComparison, got %T", expr)
	}
	fmt.Printf("Left: %s (type: %T)\n", comp.Left.String(), comp.Left)
	fmt.Printf("Right: %s (type: %T)\n", comp.Right.String(), comp.Right)

	// Current state: record.value = 10
	currentRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(10),
	})

	// Next state: record'.value = 15 (increased)
	primedRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(15),
	})

	bindings := NewBindings()
	bindings.Set("record", currentRecord, NamespaceGlobal)
	bindings.SetPrimed("record", primedRecord)

	// Evaluate left (record.value')
	leftResult := Eval(comp.Left, bindings)
	if leftResult.IsError() {
		t.Fatalf("Left eval error: %v", leftResult.Error.Message)
	}
	fmt.Printf("Left result: %v\n", leftResult.Value.Inspect())

	// Evaluate right (record.value)
	rightResult := Eval(comp.Right, bindings)
	if rightResult.IsError() {
		t.Fatalf("Right eval error: %v", rightResult.Error.Message)
	}
	fmt.Printf("Right result: %v\n", rightResult.Value.Inspect())

	result := Eval(expr, bindings)
	if result.IsError() {
		t.Fatalf("Eval error: %v", result.Error.Message)
	}
	fmt.Printf("Result: %v (type: %T)\n", result.Value.Inspect(), result.Value)
}
