package convert_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/stretchr/testify/require"
)

func TestLowerRecursiveSumGlobalFunction(t *testing.T) {
	gfKey, err := identity.NewGlobalFunctionKey("_sumadjustmentamounts")
	require.NoError(t, err)

	ctx := &convert.LowerContext{
		Parameters: map[string]bool{"amounts": true},
		GlobalFunctions: map[string]identity.Key{
			"_SumAdjustmentAmounts": gfKey,
		},
	}
	spec := `IF amounts = {} THEN 0 ELSE LET x == CHOOSE y \in amounts : TRUE IN x + _SumAdjustmentAmounts(amounts \ {x})`
	pf := convert.NewExpressionParseFuncStrict(ctx)
	expr, _, err := pf(spec)
	require.NoError(t, err)
	require.NotNil(t, expr)
}
