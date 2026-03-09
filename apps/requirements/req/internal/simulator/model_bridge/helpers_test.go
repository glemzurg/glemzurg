package model_bridge

import (
	"math/big"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

// parsedSpec creates a TLA+ ExpressionSpec with the expression parsed via the convert pipeline.
// Any unresolved identifiers in the expression should be passed as extra parameters.
func parsedSpec(tla string, params ...string) logic_spec.ExpressionSpec {
	return parsedSpecWithParams(tla, params)
}

// parsedSpecWithParams creates a TLA+ ExpressionSpec with parameter names in scope.
func parsedSpecWithParams(tla string, params []string) logic_spec.ExpressionSpec {
	ctx := &convert.LowerContext{}
	if len(params) > 0 {
		ctx.Parameters = make(map[string]bool, len(params))
		for _, p := range params {
			ctx.Parameters[p] = true
		}
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// emptySpec creates a TLA+ ExpressionSpec with no specification body (nil Expression).
func emptySpec() logic_spec.ExpressionSpec {
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", "", nil))
	return spec
}

// big0 returns a *big.Int with value 0.
func big0() *big.Int {
	return big.NewInt(0)
}
