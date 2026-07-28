package convert_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestAssociationSetAddGuaranteeTLARoundTrip(t *testing.T) {
	t.Run("ASCII _new input raises as «new»", func(t *testing.T) {
		ctx := associationSetAddFixture()
		spec := `IsSubdividedInto \union {_new(PartId, Label)}`

		astExpr, err := parser.ParseExpression(spec)
		require.NoError(t, err)

		lowered, err := convert.Lower(astExpr, ctx)
		require.NoError(t, err)

		raised, err := convert.Raise(lowered, convert.RaiseContextFromLower(ctx))
		require.NoError(t, err)

		printed := ast.Print(raised)
		require.Contains(t, printed, "«new»")
		require.NotContains(t, printed, "_new")
		require.Contains(t, printed, "IsSubdividedInto")
		require.Contains(t, printed, "∪")
	})

	t.Run("canonical «new» input round-trips", func(t *testing.T) {
		ctx := associationSetAddFixture()
		spec := `IsSubdividedInto ∪ {«new»(PartId, Label)}`

		astExpr, err := parser.ParseExpression(spec)
		require.NoError(t, err)

		lowered, err := convert.Lower(astExpr, ctx)
		require.NoError(t, err)

		raised, err := convert.Raise(lowered, convert.RaiseContextFromLower(ctx))
		require.NoError(t, err)

		printed := ast.Print(raised)
		require.Contains(t, printed, "«new»")
		require.Contains(t, printed, "IsSubdividedInto")
		require.Contains(t, printed, "∪")

		reparsed, err := parser.ParseExpression(printed)
		require.NoError(t, err)
		relowered, err := convert.Lower(reparsed, ctx)
		require.NoError(t, err)
		require.Equal(t, lowered, relowered)
	})
}

func associationSetAddFixture() *convert.LowerContext {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	containerKey := helper.Must(identity.NewClassKey(subdomainKey, "container"))
	partKey := helper.Must(identity.NewClassKey(subdomainKey, "part"))
	eventNewKey := helper.Must(identity.NewEventKey(containerKey, "_new"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, containerKey, partKey, "is_subdivided_into"))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Is Subdivided Into", Details: ""},
		model_class.AssociationEnd{ClassKey: containerKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: partKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{},
	)

	class := model_class.NewClass(containerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Container"})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventNewKey: model_state.NewEvent(eventNewKey, "_new", "", []string{"PartId", "Label"}),
	})

	associations := map[identity.Key]model_class.Association{assocKey: assoc}
	return &convert.LowerContext{
		ClassKey:         containerKey,
		AssociationNames: convert.BuildOutgoingAssociationFieldNameMap(containerKey, associations),
		SystemEventNames: convert.BuildSystemEventNameMap(&class),
		Parameters:       map[string]bool{"PartId": true, "Label": true},
	}
}
