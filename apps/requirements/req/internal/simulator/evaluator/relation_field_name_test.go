package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAssociationFieldName(t *testing.T) {
	t.Parallel()
	require.Equal(t, "ConfiguresCustomersFor", NormalizeAssociationFieldName("Configures Customers For"))
	require.Equal(t, "JurisdictionalWalletDefinition", NormalizeAssociationFieldName("Jurisdictional Wallet Definition"))
	require.Equal(t, "Lines", NormalizeAssociationFieldName("Lines"))
}
