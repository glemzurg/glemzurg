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
		`IF (\E c \in self.Adjusts : c.amount > 0) /\ (\E c \in self.Adjusts : c.amount < 0) THEN _SumAmounts(_AmountsBag(self.Adjusts)) = 0 ELSE TRUE`,
	}
	for _, s := range specs {
		_, err := parser.ParseExpression(s)
		require.NoError(t, err, "spec %q", s)
	}
}
