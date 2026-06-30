package convert_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestParseEmptyAssociationCondition(t *testing.T) {
	_, err := parser.ParseExpression("AppliesSocialCurrencyLogic = {}")
	require.NoError(t, err)
}

func TestLowerAddOrUpdateGuaranteeCondition(t *testing.T) {
	ctx := associationSetMapFixture()
	spec := `IF AppliesSocialCurrencyLogic = {} THEN AppliesSocialCurrencyLogic \union {_new(MinimumBalance, TopoffBalance)} ELSE { Update(r, MinimumBalance, TopoffBalance) : r \in AppliesSocialCurrencyLogic }`
	astExpr, err := parser.ParseExpression(spec)
	require.NoError(t, err)
	_, err = convert.Lower(astExpr, ctx)
	require.NoError(t, err)
}
