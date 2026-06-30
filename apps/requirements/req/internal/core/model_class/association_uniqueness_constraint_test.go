package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestAssociationUniquenessConstraintSuite(t *testing.T) {
	suite.Run(t, new(AssociationUniquenessConstraintSuite))
}

type AssociationUniquenessConstraintSuite struct {
	suite.Suite
}

func (suite *AssociationUniquenessConstraintSuite) TestValidate() {
	jurisdictionAttrKey := helper.Must(identity.ParseKey("domain/d/subdomain/s/class/jurisdiction/attribute/jurisdiction_code"))
	tests := []struct {
		name       string
		constraint AssociationUniquenessConstraint
		errstr     string
	}{
		{
			name: "valid per_from_instance",
			constraint: NewAssociationUniquenessConstraint(
				AssociationUniquenessScopePerFromInstance,
				AssociationUniquenessKey{ToAttributeKeys: []identity.Key{jurisdictionAttrKey}},
				0,
			),
		},
		{
			name: "valid global composite",
			constraint: NewAssociationUniquenessConstraint(
				AssociationUniquenessScopeGlobal,
				AssociationUniquenessKey{
					FromAttributeKeys: []identity.Key{helper.Must(identity.ParseKey("domain/d/subdomain/s/class/currency/attribute/abbr"))},
					ToAttributeKeys:   []identity.Key{jurisdictionAttrKey},
				},
				2,
			),
		},
		{
			name: "error invalid scope",
			constraint: AssociationUniquenessConstraint{
				Scope:    "per_pair",
				Key:      AssociationUniquenessKey{ToAttributeKeys: []identity.Key{jurisdictionAttrKey}},
				MaxCount: 1,
			},
			errstr: "scope",
		},
		{
			name: "error empty key",
			constraint: AssociationUniquenessConstraint{
				Scope:    AssociationUniquenessScopeGlobal,
				MaxCount: 1,
			},
			errstr: "at least one",
		},
		{
			name: "error max zero",
			constraint: AssociationUniquenessConstraint{
				Scope:    AssociationUniquenessScopeGlobal,
				Key:      AssociationUniquenessKey{ToAttributeKeys: []identity.Key{jurisdictionAttrKey}},
				MaxCount: 0,
			},
			errstr: "max must be at least 1",
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx := coreerr.NewContext("test", "")
			err := tc.constraint.Validate(ctx)
			if tc.errstr == "" {
				suite.Require().NoError(err)
				if tc.constraint.MaxCount == 0 {
					suite.Equal(uint(1), NewAssociationUniquenessConstraint(tc.constraint.Scope, tc.constraint.Key, 0).MaxCount)
				}
			} else {
				suite.Require().ErrorContains(err, tc.errstr)
			}
		})
	}
}

func (suite *AssociationUniquenessConstraintSuite) TestValidateAttributeReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	jurisdictionAttrKey := helper.Must(identity.NewAttributeKey(jurisdictionKey, "jurisdiction_code"))
	missingAttrKey := helper.Must(identity.NewAttributeKey(jurisdictionKey, "missing"))

	partner := NewClass(partnerKey, ClassLinks{}, ClassDetails{Name: "Partner"})
	jurisdiction := NewClass(jurisdictionKey, ClassLinks{}, ClassDetails{Name: "Jurisdiction"})
	jurisdiction.SetAttributes([]Attribute{
		helper.Must(NewAttribute(jurisdictionAttrKey, AttributeDetails{Name: "Jurisdiction Code"}, "unconstrained", nil, true, AttributeAnnotations{})),
	})

	constraint := NewAssociationUniquenessConstraint(
		AssociationUniquenessScopePerFromInstance,
		AssociationUniquenessKey{ToAttributeKeys: []identity.Key{jurisdictionAttrKey}},
		1,
	)
	ctx := coreerr.NewContext("test", "")
	suite.Require().NoError(constraint.ValidateAttributeReferences(ctx, partner, jurisdiction))

	bad := NewAssociationUniquenessConstraint(
		AssociationUniquenessScopePerFromInstance,
		AssociationUniquenessKey{ToAttributeKeys: []identity.Key{missingAttrKey}},
		1,
	)
	err := bad.ValidateAttributeReferences(ctx, partner, jurisdiction)
	suite.Require().ErrorContains(err, "missing")
}
