package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestAssociationUniquenessSuite(t *testing.T) {
	suite.Run(t, new(AssociationUniquenessSuite))
}

type AssociationUniquenessSuite struct {
	suite.Suite
}

func (suite *AssociationUniquenessSuite) TestValidate() {
	jurisdictionAttrKey := helper.Must(identity.ParseKey("domain/d/subdomain/s/class/jurisdiction/attribute/jurisdiction_code"))
	currencyAttrKey := helper.Must(identity.ParseKey("domain/d/subdomain/s/class/currency/attribute/abbr"))

	tests := []struct {
		name       string
		uniqueness AssociationUniqueness
		errstr     string
	}{
		{
			name: "valid to attributes only",
			uniqueness: NewAssociationUniqueness(
				nil,
				[]identity.Key{jurisdictionAttrKey},
			),
		},
		{
			name: "valid composite",
			uniqueness: NewAssociationUniqueness(
				[]identity.Key{currencyAttrKey},
				[]identity.Key{jurisdictionAttrKey},
			),
		},
		{
			name:       "error empty keys",
			uniqueness: AssociationUniqueness{},
			errstr:     "at least one",
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx := coreerr.NewContext("test", "")
			err := tc.uniqueness.Validate(ctx)
			if tc.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tc.errstr)
			}
		})
	}
}

func (suite *AssociationUniquenessSuite) TestValidateAttributeReferences() {
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

	uniqueness := NewAssociationUniqueness(nil, []identity.Key{jurisdictionAttrKey})
	ctx := coreerr.NewContext("test", "")
	suite.Require().NoError(uniqueness.ValidateAttributeReferences(ctx, partner, jurisdiction))

	bad := NewAssociationUniqueness(nil, []identity.Key{missingAttrKey})
	err := bad.ValidateAttributeReferences(ctx, partner, jurisdiction)
	suite.Require().ErrorContains(err, "missing")
}