package evaluator

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestAssociationClassHostNavigationEndpointsAndMember(t *testing.T) {
	t.Parallel()

	ctx := NewRelationContext()
	hostKey := AssociationKey("domain/d/subdomain/s/cassociation/class/partner/class/jurisdiction/configures")
	partnerKey := "domain/d/subdomain/s/class/partner"
	jurisdictionKey := "domain/d/subdomain/s/class/jurisdiction"

	ctx.AddAssociationClassHost(
		hostKey,
		"Configures",
		AssociationHostEndpoints{FromClassKey: partnerKey, ToClassKey: jurisdictionKey},
		"Link Def",
		AssociationHostMultiplicities{},
	)

	partner := object.NewRecord()
	partner.Set("_state", object.NewString("Active"))
	jurisdiction := object.NewRecord()
	jurisdiction.Set("_state", object.NewString("Active"))
	link := object.NewRecord()
	link.Set("_state", object.NewString("Active"))
	link.Set("Fee", object.NewInteger(5))

	ctx.CreateLink(hostKey, partner, jurisdiction)
	ctx.AddAssociationClassRow(hostKey, partner, jurisdiction, link)

	relInfo := ctx.GetForwardRelation(partnerKey, "Configures")
	require.NotNil(t, relInfo)
	require.Equal(t, jurisdictionKey, relInfo.TargetClassKey)
	require.Equal(t, "LinkDef", relInfo.LinkClassMember)

	endpointResult := evalRelationTraversal(partner, relInfo, ctx)
	require.False(t, endpointResult.IsError())
	assocRel, ok := endpointResult.Value.(*object.AssociationRelation)
	require.True(t, ok)
	require.Equal(t, 1, assocRel.Endpoints().Size())
	resolved, ok := assocRel.LinkForEndpoint(jurisdiction)
	require.True(t, ok)
	require.Equal(t, link, resolved)

	bindings := NewBindings().WithSelfAndClass(partner, partnerKey)
	bindings.SetRelationContext(ctx)

	memberAccess := &me.FieldAccess{
		Base:  &me.FieldAccess{Base: &me.SelfRef{}, Field: "Configures"},
		Field: "LinkDef",
	}
	memberResult := Eval(memberAccess, bindings)
	require.False(t, memberResult.IsError())
	linkRecord, ok := memberResult.Value.(*object.Record)
	require.True(t, ok)
	require.Equal(t, link, linkRecord)

	nestedField := &me.FieldAccess{
		Base:  memberAccess,
		Field: "Fee",
	}
	feeResult := Eval(nestedField, bindings)
	require.False(t, feeResult.IsError())
	fee, ok := feeResult.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, int64(5), fee.Rat().Num().Int64())
}

func TestAssociationClassMemberResolvesViaQuantifierEndpoint(t *testing.T) {
	t.Parallel()

	ctx := NewRelationContext()
	hostKey := AssociationKey("domain/d/subdomain/s/cassociation/class/partner/class/jurisdiction/configures")
	partnerKey := "domain/d/subdomain/s/class/partner"
	jurisdictionKey := "domain/d/subdomain/s/class/jurisdiction"

	ctx.AddAssociationClassHost(
		hostKey,
		"Configures",
		AssociationHostEndpoints{FromClassKey: partnerKey, ToClassKey: jurisdictionKey},
		"Link Def",
		AssociationHostMultiplicities{},
	)

	partner := object.NewRecord()
	j1 := object.NewRecord()
	j1.Set("Code", object.NewString("US"))
	j2 := object.NewRecord()
	j2.Set("Code", object.NewString("UK"))
	link1 := object.NewRecord()
	link1.Set("Fee", object.NewInteger(1))
	link2 := object.NewRecord()
	link2.Set("Fee", object.NewInteger(2))

	ctx.CreateLink(hostKey, partner, j1)
	ctx.CreateLink(hostKey, partner, j2)
	ctx.AddAssociationClassRow(hostKey, partner, j1, link1)
	ctx.AddAssociationClassRow(hostKey, partner, j2, link2)

	bindings := NewBindings().WithSelfAndClass(partner, partnerKey)
	bindings.SetRelationContext(ctx)

	child := NewEnclosedBindings(bindings)
	child.Set("j", j1, NamespaceLocal)

	memberAccess := &me.FieldAccess{
		Base:  &me.FieldAccess{Base: &me.SelfRef{}, Field: "Configures"},
		Field: "LinkDef",
	}
	memberResult := Eval(memberAccess, child)
	require.False(t, memberResult.IsError())
	linkRecord, ok := memberResult.Value.(*object.Record)
	require.True(t, ok)
	require.Equal(t, link1, linkRecord)
}

func TestAssociationClassMemberAmbiguousWithoutEndpointContext(t *testing.T) {
	t.Parallel()

	ctx := NewRelationContext()
	hostKey := AssociationKey("domain/d/subdomain/s/cassociation/class/partner/class/jurisdiction/configures")
	partnerKey := "domain/d/subdomain/s/class/partner"
	jurisdictionKey := "domain/d/subdomain/s/class/jurisdiction"

	ctx.AddAssociationClassHost(
		hostKey,
		"Configures",
		AssociationHostEndpoints{FromClassKey: partnerKey, ToClassKey: jurisdictionKey},
		"Link Def",
		AssociationHostMultiplicities{},
	)

	partner := object.NewRecord()
	j1 := object.NewRecord()
	j1.Set("Code", object.NewString("US"))
	j2 := object.NewRecord()
	j2.Set("Code", object.NewString("UK"))
	link1 := object.NewRecord()
	link2 := object.NewRecord()

	ctx.CreateLink(hostKey, partner, j1)
	ctx.CreateLink(hostKey, partner, j2)
	ctx.AddAssociationClassRow(hostKey, partner, j1, link1)
	ctx.AddAssociationClassRow(hostKey, partner, j2, link2)

	bindings := NewBindings().WithSelfAndClass(partner, partnerKey)
	bindings.SetRelationContext(ctx)

	memberAccess := &me.FieldAccess{
		Base:  &me.FieldAccess{Base: &me.SelfRef{}, Field: "Configures"},
		Field: "LinkDef",
	}
	memberResult := Eval(memberAccess, bindings)
	require.True(t, memberResult.IsError())
	require.Contains(t, memberResult.Error.Inspect(), "exactly one association row")
}
