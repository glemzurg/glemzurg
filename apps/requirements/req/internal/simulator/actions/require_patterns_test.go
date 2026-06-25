package actions

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestDetectPeerFieldDistinctFromParam(t *testing.T) {
	classKey := identity.Key{SubKey: "currency"}
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "c",
		Domain:   &me.ClassRef{ClassKey: classKey, Name: "Currency"},
		Predicate: &me.Compare{
			Op: me.CompareNeq,
			Left: &me.FieldAccess{
				Base:  &me.LocalVar{Name: "c"},
				Field: "abbr",
			},
			Right: &me.LocalVar{Name: "Abbr"},
		},
	}

	pattern, ok := detectPeerFieldDistinctFromParam(expr)
	require.True(t, ok)
	require.Equal(t, classKey, pattern.ClassKey)
	require.Equal(t, "Currency", pattern.ClassName)
	require.Equal(t, "abbr", pattern.FieldSubKey)
	require.Equal(t, "Abbr", pattern.ParamName)
}

func TestAssessPeerFieldDistinctFromParamExcludesSelf(t *testing.T) {
	classKey := identity.Key{SubKey: "currency"}
	self := object.NewRecord()
	self.Set("abbr", object.NewString("ZZZ"))

	peer := object.NewRecord()
	peer.Set("abbr", object.NewString("USD"))

	classSet := object.NewSet()
	classSet.Add(self)
	classSet.Add(peer)

	bindings := evaluator.NewBindings()
	bindings.Set("Currency", classSet, evaluator.NamespaceGlobal)
	bindings.Set("Abbr", object.NewString("ZZZ"), evaluator.NamespaceLocal)
	child := bindings.WithSelf(self)
	child.Set("Abbr", object.NewString("ZZZ"), evaluator.NamespaceLocal)

	pattern := peerFieldDistinctFromParamPattern{
		ClassKey:    classKey,
		ClassName:   "Currency",
		FieldSubKey: "abbr",
		ParamName:   "Abbr",
	}
	require.True(t, assessPeerFieldDistinctFromParam(pattern, child))
}

func TestAssessPeerFieldDistinctFromParamFailsOnDuplicatePeer(t *testing.T) {
	classKey := identity.Key{SubKey: "currency"}
	self := object.NewRecord()
	self.Set("abbr", object.NewString("NEW"))

	peer := object.NewRecord()
	peer.Set("abbr", object.NewString("USD"))

	classSet := object.NewSet()
	classSet.Add(self)
	classSet.Add(peer)

	bindings := evaluator.NewBindings()
	bindings.Set("Currency", classSet, evaluator.NamespaceGlobal)
	child := bindings.WithSelf(self)
	child.Set("Abbr", object.NewString("USD"), evaluator.NamespaceLocal)

	pattern := peerFieldDistinctFromParamPattern{
		ClassKey:    classKey,
		ClassName:   "Currency",
		FieldSubKey: "abbr",
		ParamName:   "Abbr",
	}
	require.False(t, assessPeerFieldDistinctFromParam(pattern, child))
}

func TestAssessPeerFieldDistinctFromParamFailsOnDuplicateNullPeer(t *testing.T) {
	classKey := identity.Key{SubKey: "jurisdiction"}
	self := object.NewRecord()
	peer := object.NewRecord()
	peer.Set("jurisdiction_code", object.Null())

	classSet := object.NewSet()
	classSet.Add(self)
	classSet.Add(peer)

	bindings := evaluator.NewBindings()
	bindings.Set("Jurisdiction", classSet, evaluator.NamespaceGlobal)
	child := bindings.WithSelf(self)
	child.Set("JurisdictionCode", object.Null(), evaluator.NamespaceLocal)

	pattern := peerFieldDistinctFromParamPattern{
		ClassKey:    classKey,
		ClassName:   "Jurisdiction",
		FieldSubKey: "jurisdiction_code",
		ParamName:   "JurisdictionCode",
	}
	require.False(t, assessPeerFieldDistinctFromParam(pattern, child))
}
