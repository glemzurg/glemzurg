package parser_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestParseBagAmountsGlobals(t *testing.T) {
	specs := []string{
		`IF adjustments = {} THEN _Bags!SetToBag({}) ELSE LET c == CHOOSE x \in adjustments : TRUE IN _Bags!SetToBag({c.amount}) ⊕ _AmountsBag(adjustments \ {c})`,
		`IF _Bags!BagToSet(amounts) = {} THEN 0 ELSE LET x == CHOOSE y \in _Bags!BagToSet(amounts) : _Bags!CopiesIn(y, amounts) > 0 IN x + _SumAmounts(amounts ⊖ _Bags!SetToBag({x}))`,
		`IF _Bags!BagCardinality(self.Adjusts) = 1 THEN TRUE ELSE _SumAmounts(_AmountsBag(self.Adjusts.AcountBalanceChange.Amount)) = 0`,
	}
	for _, s := range specs {
		_, err := parser.ParseExpression(s)
		require.NoError(t, err, "spec %q", s)
	}
}
