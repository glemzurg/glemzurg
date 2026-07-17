package state

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestAssociationNavigationKeepsIdenticalPeerDataDistinct(t *testing.T) {
	t.Parallel()

	sim := NewSimulationState()
	txnKey := testClassKey(t, "finance", "wallet", "transaction")
	acctKey := testClassKey(t, "finance", "wallet", "account")
	abcKey := testClassKey(t, "finance", "wallet", "account_balance_change")
	assocKey := testAssocKey(t, "finance", "wallet", "transaction", "account", "adjusts")

	txn := sim.CreateInstance(txnKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Recorded"),
	}))
	// Three accounts with identical attribute data — historically collapsed to one.
	identical := map[string]object.Object{"_state": object.NewString("Exists")}
	a1 := sim.CreateInstance(acctKey, object.NewRecordFromFields(identical))
	a2 := sim.CreateInstance(acctKey, object.NewRecordFromFields(identical))
	a3 := sim.CreateInstance(acctKey, object.NewRecordFromFields(identical))
	abc1 := sim.CreateInstance(abcKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Recorded"),
		"amount": object.NewInteger(75),
	}))
	abc2 := sim.CreateInstance(abcKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Recorded"),
		"amount": object.NewInteger(-51),
	}))
	abc3 := sim.CreateInstance(abcKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Recorded"),
		"amount": object.NewInteger(-25),
	}))

	require.NoError(t, sim.AddAssociationLink(assocKey, txn.ID, a1.ID, abc1.ID))
	require.NoError(t, sim.AddAssociationLink(assocKey, txn.ID, a2.ID, abc2.ID))
	require.NoError(t, sim.AddAssociationLink(assocKey, txn.ID, a3.ID, abc3.ID))

	builder := NewBindingsBuilder(sim)
	builder.AddAssociationClassHost(
		assocKey,
		"Adjusts",
		evaluator.AssociationHostEndpoints{
			FromClassKey: txnKey.String(),
			ToClassKey:   acctKey.String(),
		},
		"Account Balance Change",
		evaluator.AssociationHostMultiplicities{},
	)

	bindings := builder.BuildForInstance(txn)

	result := evaluator.Eval(&me.FieldAccess{Base: &me.SelfRef{}, Field: "Adjusts"}, bindings)
	require.False(t, result.IsError(), "%v", result.Error)
	assocRel, ok := result.Value.(*object.AssociationRelation)
	require.True(t, ok)
	require.Equal(t, 3, assocRel.Endpoints().Size(), "identical peer data must not collapse")

	for _, elem := range assocRel.Endpoints().Elements() {
		rec, ok := elem.(*object.Record)
		require.True(t, ok)
		require.True(t, object.IsExtentElement(rec), "association endpoints must be [id, data]")
	}

	// Multi-endpoint AC member → set of three link rows (not a sole scalar).
	acAccess := &me.FieldAccess{
		Base:  &me.FieldAccess{Base: &me.SelfRef{}, Field: "Adjusts"},
		Field: "AccountBalanceChange",
	}
	acResult := evaluator.Eval(acAccess, bindings)
	require.False(t, acResult.IsError(), "%v", acResult.Error)
	acSet, ok := acResult.Value.(*object.Set)
	require.True(t, ok, "multi-endpoint AC member should be a set, got %T", acResult.Value)
	require.Equal(t, 3, acSet.Size())
}

func testClassKey(t *testing.T, domain, subdomain, class string) identity.Key {
	t.Helper()
	domainKey, err := identity.NewDomainKey(domain)
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomain)
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, class)
	require.NoError(t, err)
	return classKey
}

func testAssocKey(t *testing.T, domain, subdomain, fromClass, toClass, name string) identity.Key {
	t.Helper()
	domainKey, err := identity.NewDomainKey(domain)
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomain)
	require.NoError(t, err)
	fromKey, err := identity.NewClassKey(subdomainKey, fromClass)
	require.NoError(t, err)
	toKey, err := identity.NewClassKey(subdomainKey, toClass)
	require.NoError(t, err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, name)
	require.NoError(t, err)
	return assocKey
}
