package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomValueUniqueUnorderedEnumIsSet(t *testing.T) {
	parent := helper.Must(identity.NewAttributeKey(
		helper.Must(identity.NewClassKey(
			helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
			"c",
		)),
		"allowed_operations",
	))
	key := helper.Must(identity.NewDataTypeKey(parent, "self"))
	dt, err := model_data_type.New(key, "unique unordered of enum of withdraw, deposit, wager", nil)
	require.NoError(t, err)

	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed
	for i := range 20 {
		val := generateRandomValue(dt, rng)
		set, ok := val.(*object.Set)
		require.Truef(t, ok, "iteration %d: got %T %s", i, val, val.Inspect())
		require.Positive(t, set.Size())
		require.LessOrEqual(t, set.Size(), 3)
		for _, elem := range set.Elements() {
			str, ok := elem.(*object.String)
			require.True(t, ok)
			require.Contains(t, []string{"withdraw", "deposit", "wager"}, str.Value())
		}
	}
}
