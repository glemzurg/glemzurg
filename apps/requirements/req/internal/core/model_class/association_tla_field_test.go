package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestAssociationTLAFieldName(t *testing.T) {
	require.Equal(t, "IsSubdividedInto", AssociationTLAFieldName("Is Subdivided Into"))
	require.Equal(t, "ConfiguresCustomersFor", AssociationTLAFieldName("Configures Customers For"))
}

func TestAttributeTLAFieldName(t *testing.T) {
	require.Equal(t, "Amount", AttributeTLAFieldName("Amount"))
	require.Equal(t, "JurisdictionCode", AttributeTLAFieldName("Jurisdiction Code"))
	require.Equal(t, "IsSocialOnly", AttributeTLAFieldName("Is Social Only"))
}

func TestOutgoingAssociationTLAFieldSet(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "container"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "part"))
	otherKey := helper.Must(identity.NewClassKey(subdomainKey, "other"))

	assoc := NewAssociation(
		helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "is_subdivided_into")),
		AssociationDetails{Name: "Is Subdivided Into", Details: ""},
		AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(NewMultiplicity("1"))},
		AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(NewMultiplicity("any"))},
		AssociationOptions{},
	)

	got := OutgoingAssociationTLAFieldSet(fromKey, map[identity.Key]Association{assoc.Key: assoc})
	require.True(t, got["IsSubdividedInto"])
	require.Len(t, got, 1)

	require.Nil(t, OutgoingAssociationTLAFieldSet(otherKey, map[identity.Key]Association{assoc.Key: assoc}))
}

func TestReverseAssociationTLAFieldName(t *testing.T) {
	require.Equal(t, "_IsSubdividedInto", ReverseAssociationTLAFieldName("Is Subdivided Into"))
}

func TestIncomingAssociationReverseTLAFieldSet(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "wallet"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "account"))

	assoc := NewAssociation(
		helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "is_subdivided_into")),
		AssociationDetails{Name: "Is Subdivided Into", Details: ""},
		AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(NewMultiplicity("1"))},
		AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(NewMultiplicity("any"))},
		AssociationOptions{},
	)

	got := IncomingAssociationReverseTLAFieldSet(toKey, map[identity.Key]Association{assoc.Key: assoc})
	require.True(t, got["_IsSubdividedInto"])
	require.Nil(t, IncomingAssociationReverseTLAFieldSet(fromKey, map[identity.Key]Association{assoc.Key: assoc}))
}
